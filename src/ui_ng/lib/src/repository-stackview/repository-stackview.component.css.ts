export const REPOSITORY_STACKVIEW_STYLES: string = `
.option-right {
  padding-right: 16px;
}
.sub-grid-custom {
  left: 40px;
}
.refresh-btn {
    cursor: pointer;
}
.refresh-btn:hover {
    color: #007CBB;
}

:host >>> .datagrid .datagrid-body {
  overflow-x: hidden;
}

:host >>> .datagrid .datagrid-foot {
  border-top: 1px solid #ccc;
}

:host >>> .datagrid .datagrid-body .datagrid-row {
  background-color: #ccc;
}

:host >>> .datagrid-body .datagrid-row .datagrid-row-master{
  background-color: #FFFFFF;
}

:host >>> .datagrid .datagrid-placeholder-container {
  display: none;
}
:host >>> .datagrid-overlay-wrapper{margin-top:24px;}

.db-status-warning {
  position: absolute;
  left: 24px;
  display: inline-block;
}
.rightPos{
    position: absolute;
    z-index: 100;
    right: 35px;
    margin-top: 4px;
}
`;
