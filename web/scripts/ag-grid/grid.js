import { actionCellRenderer } from "./action-cell-renderer.js";
import { deleteSelectedBtnEvent } from "./ag-grid-toolbar.js";

let gridApi;

const rowSelection = {
  mode: "multiRow",
  checkboxes: true,
  // enableClickSelection: true,
};

const columnDefs = [
  {
    field: "shipmentNumber",
    headerName: "Shipment Number",
    sortable: true,
    filter: "agTextColumnFilter",
    width: 200,
  },
  {
    field: "shipmentType",
    headerName: "Type",
    sortable: true,
    filter: "agSetColumnFilter", // Enterprise
    width: 90,
  },
  {
    field: "sealineCode",
    headerName: "Sealine",
    sortable: true,
    filter: "agSetColumnFilter",
    width: 100,
  },
  {
    field: "sealineName",
    headerName: "Sealine Name",
    sortable: true,
    filter: "agTextColumnFilter",
    width: 150,
  },
  {
    field: "shippingStatus",
    headerName: "Status",
    sortable: true,
    filter: "agSetColumnFilter",
    width: 120,
    cellRenderer: (params) => {
      const statusMap = {
        IN_TRANSIT: "In Transit",
        DELIVERED: "Delivered",
        PLANNED: "Planned",
        UNKNOWN: "Unknown",
      };

      const status = statusMap[params.value] || "Unknown";
      const statusClasses = {
        "In Transit": "bg-blue-100 text-blue-800",
        Delivered: "bg-green-100 text-green-800",
        Planned: "bg-yellow-100 text-yellow-800",
        Unknown: "bg-gray-100 text-gray-800",
      };
      const classes = statusClasses[status] || statusClasses["Unknown"];
      return `<span class="px-2 py-1 text-xs font-semibold rounded-full ${classes}">${status}</span>`;
    },
  },
  {
    field: "actions",
    headerName: "Actions",
    width: 100,
    pinned: "right",
    sortable: false,
    cellRenderer: actionCellRenderer,
  },
];

const gridOptions = {
  columnDefs,
  defaultColDef: {
    flex: 1,
  },
  rowSelection,
  pagination: true,
  paginationPageSize: 20,
  // rowModelType: "serverSide",
};

export function getGridApi() {
  const gridDiv = document.querySelector("#grid");
  gridApi = agGrid.createGrid(gridDiv, gridOptions);

  const deleteSelectedBtn = document.getElementById("deleteSelectedBtn");
  deleteSelectedBtn.addEventListener("click", () =>
    deleteSelectedBtnEvent(gridApi),
  );

  const refreshGridBtn = document.getElementById("refreshGridBtn");
  refreshGridBtn.addEventListener("click", () => loadShipments(gridApi));
  return gridApi;
}

export function loadShipments(gridApi) {
  fetch("/api/shipments/grid-data", {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  })
    .then((response) => response.json())
    .then((data) => {
      console.log("Data:", data);
      gridApi.setGridOption("rowData", data.rows);
    })
    .catch((error) => {
      console.error("Error fetching data:", error);
      params.fail();
    });
}
