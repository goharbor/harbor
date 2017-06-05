export const TAG_STYLE = `
.sub-header-title {
  margin: 12px 0;
}

.embeded-datagrid {
  width: 98%;
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

:host >>> .datagrid .datagrid-body {
  background-color: #eee;
}

:host >>> .datagrid .datagrid-head .datagrid-row {
  background-color: #eee;
}

:host >>> .datagrid .datagrid-body .datagrid-row-master {
  background-color: #eee;
}
`;