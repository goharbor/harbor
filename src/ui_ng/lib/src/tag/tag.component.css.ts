export const TAG_STYLE = `
.option-right {
  padding-right: 18px;
  padding-bottom: 6px;
}

.refresh-btn {
  cursor: pointer;
}

.refresh-btn:hover {
  color: #007CBB;
}

.sub-header-title {
  margin: 12px 0;
}

.embeded-datagrid {
  width: 98%;
  float:right; /*add for issue #2688*/
}

.hidden-tag {
  display: block; height: 0;
}

:host >>> .datagrid-placeholder {
  display: none;
}

.truncated {
  display: inline-block;
  overflow: hidden;
  white-space: nowrap;
  text-overflow:ellipsis;
}

.copy-failed {
  color: red;
  margin-right: 6px;
}

:host >>> .datagrid clr-dg-column {
    min-width: 80px;
}
.rightPos{
    position: absolute;
    z-index: 100;
    right: 35px;
    margin-top: 4px;
}

.btn-group .dropdown-menu clr-icon{display: block;}
.dropdown-menu .dropdown-item{position: relative;padding-left:.5rem; padding-right:.5rem;}
.dropdown-menu input{position: relative;margin-left:.5rem; margin-right:.5rem;}
.pull-left{display:inline-block;float:left;}
.pull-right{display:inline-block; float:right;}
.btn-link{display:inline-flex;width: 15px;min-width:15px; color:black; vertical-align: super; }
.trigger-item, .signpost-item{display: inline;}
.signpost-content-body .label{margin:.3rem;}
.labelDiv{position: absolute; left:34px;top:3px;}
.datagrid-action-bar{z-index:10;}
.trigger-item hbr-label-piece{display: flex !important;margin: 6px 0;}
:host >>> .signpost-content{min-width:4rem;}
:host >>> .signpost-content-body{padding:0 .4rem;}
:host >>> .signpost-content-header{display:none;}
.filterLabelPiece{position: absolute; bottom :0px;z-index:1;}
.dropdown .dropdown-toggle.btn {
    margin: .25rem .5rem .25rem 0;
}
`;