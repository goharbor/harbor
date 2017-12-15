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

:host >>> .datagrid {
  margin: 0;
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

`;