export const REPOSITORY_STACKVIEW_STYLES: string = `
.option-right {
  padding-right: 16px;
}
.sub-grid-custom {
  position: relative;
  left: 40px;
}
.refresh-btn {
    cursor: pointer;
}
.refresh-btn:hover {
    color: #007CBB;
}
:host >>> .datagrid .datagrid-body .datagrid-row {
  overflow-x: hidden;
  overflow-y: hidden;
  background-color: #ccc;
}

:host >>> .datagrid-body .datagrid-row .datagrid-row-master{
  background-color: #FFFFFF;
}

:host >>> .datagrid .datagrid-placeholder-container {
  display: none;
}
`;
