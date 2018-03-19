// Copyright (c) 2017-2018 VMware, Inc. All Rights Reserved.
// This software is released under MIT license.
// The full license information can be found in LICENSE in the root directory of this project.

// @import 'node_modules/admiral-ui-common/css/mixins';

export const GRIDVIEW_STYLE = `
.grid-content {
  position: relative;
  top: 36px;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: auto;
  max-height: 65vh;
}

.card-item {
  display: block;
  max-width: 400px;
  min-width: 300px;
  position: absolute;
  margin-right: 40px;
  transition: width 0.4s, transform 0.4s;
}

.content-empty {
  text-align: center;
  display: block;
  margin-top: 100px;
}

.central-block-loading {
  position: absolute;
  z-index: 10;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  @include animation(fadein 0.4s);
  text-align: center;
  background-color: rgba(255, 255, 255, 0.5);
}
.central-block-loading-more {
  position: relative;
  z-index: 10;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  @include animation(fadein 0.4s);
  text-align: center;
  background-color: rgba(255, 255, 255, 0.5);
}
.vertical-helper {
  display: inline-block;
  height: 100%;
  vertical-align: middle;
}

.spinner {
  width: 100px;
  height: 100px;
  vertical-align: middle;
}

`