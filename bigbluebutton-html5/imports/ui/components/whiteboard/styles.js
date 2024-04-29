import styled, { createGlobalStyle } from 'styled-components';

const TldrawV2GlobalStyle = createGlobalStyle`
  ${({ isPresenter, hasWBAccess }) => (!isPresenter && hasWBAccess) && `
    [data-testid="tools.hand"] {
      display: none;
    }
  `}

  ${({ isMultiUserActive }) => !isMultiUserActive && `
    .tl-nametag {
      display: none;
    }
  `}

  ${({ isRTL }) => (!isRTL) && `
    .tlui-menu-zone {
      right: auto;
      left: 3.5rem;
    }
  `}

  ${({ isRTL }) => (isRTL) && `
    .tlui-menu-zone {
      right: 3.5rem;
      left: auto;
    }
  `}

  ${({ isToolbarVisible }) => (!isToolbarVisible) && `
    .tlui-toolbar,
    .tlui-style-panel__wrapper,
    .tlui-menu-zone {
      visibility: hidden;
    }
    #WhiteboardOptionButton {
      opacity: 0.2;
    }
  `}

  #presentationInnerWrapper > div:last-child {
    position: relative;
    height: 100%;
  }

  #presentationInnerWrapper > div:last-child > * {
    position: relative; 
    height: 100%;
  }

  #presentationInnerWrapper > div:last-child .tl-overlays {
    left: 0px;
    bottom: 0px;
  }

  .tlui-navigation-zone,
  .tlui-help-menu,
  .tlui-debug-panel {
    display: none !important;
  }

  .tlui-style-panel__wrapper {
    right: 0px;
    top: -0.35rem;
    position: relative;
  }

  // Add the following lines to override height and width attributes for .tl-overlays__item
  .tl-overlays__item {
    height: auto !important;
    width: auto !important;
  }

  ${({ isPresenter, isMultiUserActive }) => !isPresenter && !isMultiUserActive && `
    .tl-cursor use {
      transform: scale(0.05)!important;
    }

    .tl-collaborator__cursor {
      position: absolute !important;
      left: -7px !important;
      top: -6px !important;
    }
  `}

  [data-testid="main.page-menu"],
  [data-testid="main.menu"],
  [data-testid="tools.more.laser"],
  [data-testid="tools.asset"],
  .tlui-buttons__horizontal > :nth-child(1),
  .tlui-buttons__horizontal > :nth-child(2) {
    display: none !important;
  }

  .tl-collaborator__cursor {
    height: auto !important;
    width: auto !important;
  }
`;

const EditableWBWrapper = styled.div`
  &, & > :first-child {
    cursor: inherit !important;
  }
`;

export default {
  TldrawV2GlobalStyle,
  EditableWBWrapper,
};
