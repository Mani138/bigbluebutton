package writer

import (
	"context"
	"errors"
	"github.com/iMDT/bbb-graphql-middleware/internal/common"
	"github.com/iMDT/bbb-graphql-middleware/internal/msgpatch"
	log "github.com/sirupsen/logrus"
	"nhooyr.io/websocket/wsjson"
	"strings"
	"sync"
)

// HasuraConnectionWriter
// process messages (middleware to hasura)
func HasuraConnectionWriter(hc *common.HasuraConnection, fromBrowserToHasuraChannel *common.SafeChannel, wg *sync.WaitGroup, initMessage map[string]interface{}) {
	log := log.WithField("_routine", "HasuraConnectionWriter")

	browserConnection := hc.BrowserConn

	log = log.WithField("browserConnectionId", browserConnection.Id).WithField("hasuraConnectionId", hc.Id)

	defer wg.Done()
	defer hc.ContextCancelFunc()
	defer log.Debugf("finished")

	//Send authentication (init) message at first
	//It will not use the channel (fromBrowserToHasuraChannel) because this msg must bypass ChannelFreeze
	if initMessage != nil {
		log.Infof("it's a reconnection, injecting authentication (init) message")
		err := wsjson.Write(hc.Context, hc.Websocket, initMessage)
		if err != nil {
			log.Errorf("error on write authentication (init) message (we're disconnected from hasura): %v", err)
			return
		}
	}

RangeLoop:
	for {
		select {
		case <-hc.Context.Done():
			break RangeLoop
		case <-hc.FreezeMsgFromBrowserChan.ReceiveChannel():
			if !fromBrowserToHasuraChannel.Frozen() {
				log.Debug("freezing channel fromBrowserToHasuraChannel")
				//Freeze channel once it's about to close Hasura connection
				fromBrowserToHasuraChannel.FreezeChannel()
			}
		case fromBrowserMessage := <-fromBrowserToHasuraChannel.ReceiveChannel():
			{
				if fromBrowserMessage == nil {
					continue
				}

				var fromBrowserMessageAsMap = fromBrowserMessage.(map[string]interface{})

				if fromBrowserMessageAsMap["type"] == "start" {
					var queryId = fromBrowserMessageAsMap["id"].(string)

					//Identify type based on query string
					messageType := common.Query
					var lastReceivedDataChecksum uint32
					streamCursorField := ""
					streamCursorVariableName := ""
					var streamCursorInitialValue interface{}
					payload := fromBrowserMessageAsMap["payload"].(map[string]interface{})

					query, ok := payload["query"].(string)
					if ok {
						if strings.HasPrefix(query, "subscription") {
							messageType = common.Subscription

							browserConnection.ActiveSubscriptionsMutex.RLock()
							existingSubscriptionData, queryIdExists := browserConnection.ActiveSubscriptions[queryId]
							browserConnection.ActiveSubscriptionsMutex.RUnlock()
							if queryIdExists {
								lastReceivedDataChecksum = existingSubscriptionData.LastReceivedDataChecksum
							}

							if strings.Contains(query, "_stream(") && strings.Contains(query, "cursor: {") {
								messageType = common.Streaming

								if !queryIdExists {
									streamCursorField, streamCursorVariableName, streamCursorInitialValue = common.GetStreamCursorPropsFromQuery(payload, query)

									//It's necessary to assure the cursor field will return in the result of the query
									//To be able to store the last received cursor value
									payload["query"] = common.PatchQueryIncludingCursorField(query, streamCursorField)
									fromBrowserMessageAsMap["payload"] = payload
								}
							}

							if strings.Contains(query, "_aggregate") && strings.Contains(query, "aggregate {") {
								messageType = common.SubscriptionAggregate
							}
						}

						if strings.HasPrefix(query, "mutation") {
							messageType = common.Mutation
						}
					}

					//Identify if the client that requested this subscription expects to receive json-patch
					//Client append `Patched_` to the query operationName to indicate that it supports
					jsonPatchSupported := false
					operationName, ok := payload["operationName"].(string)
					if ok && strings.HasPrefix(operationName, "Patched_") {
						jsonPatchSupported = true
					}

					browserConnection.ActiveSubscriptionsMutex.Lock()
					browserConnection.ActiveSubscriptions[queryId] = common.GraphQlSubscription{
						Id:                         queryId,
						Message:                    fromBrowserMessageAsMap,
						OperationName:              operationName,
						StreamCursorField:          streamCursorField,
						StreamCursorVariableName:   streamCursorVariableName,
						StreamCursorCurrValue:      streamCursorInitialValue,
						LastSeenOnHasuraConnection: hc.Id,
						JsonPatchSupported:         jsonPatchSupported,
						Type:                       messageType,
						LastReceivedDataChecksum:   lastReceivedDataChecksum,
					}
					// log.Tracef("Current queries: %v", browserConnection.ActiveSubscriptions)
					browserConnection.ActiveSubscriptionsMutex.Unlock()
				}

				if fromBrowserMessageAsMap["type"] == "stop" {
					var queryId = fromBrowserMessageAsMap["id"].(string)
					browserConnection.ActiveSubscriptionsMutex.RLock()
					jsonPatchSupported := browserConnection.ActiveSubscriptions[queryId].JsonPatchSupported
					browserConnection.ActiveSubscriptionsMutex.RUnlock()
					if jsonPatchSupported {
						msgpatch.RemoveConnSubscriptionCacheFile(browserConnection, queryId)
					}
					browserConnection.ActiveSubscriptionsMutex.Lock()
					delete(browserConnection.ActiveSubscriptions, queryId)
					// log.Tracef("Current queries: %v", browserConnection.ActiveSubscriptions)
					browserConnection.ActiveSubscriptionsMutex.Unlock()
				}

				if fromBrowserMessageAsMap["type"] == "connection_init" {
					browserConnection.ConnectionInitMessage = fromBrowserMessageAsMap
				}

				log.Tracef("sending to hasura: %v", fromBrowserMessageAsMap)
				err := wsjson.Write(hc.Context, hc.Websocket, fromBrowserMessageAsMap)
				if err != nil {
					if !errors.Is(err, context.Canceled) {
						log.Errorf("error on write (we're disconnected from hasura): %v", err)
					}
					return
				}
			}
		}
	}
}
