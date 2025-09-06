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
    filter: "agTextColumnFilter",
    maxWidth: 200,
    cellRenderer: (params) => {
      if (!params.value) return "";

      return `
        <button
          class="text-blue-600 dark:text-blue-300 hover:underline font-medium"
          onclick="openDrawerFetchDetails('${params.data.id}')"
        >
          ${params.value}
        </button>
      `;
    },
  },
  {
    field: "shipmentType",
    headerName: "Type",
    filter: "agSetColumnFilter", // Enterprise
    maxWidth: 90,
  },
  {
    field: "sealineCode",
    headerName: "Sealine",
    filter: "agSetColumnFilter",
    maxWidth: 120,
  },
  {
    field: "sealineName",
    headerName: "Sealine Name",
    filter: "agTextColumnFilter",
  },
  {
    field: "shippingStatus",
    headerName: "Status",
    filter: "agSetColumnFilter",
    maxWidth: 120,
    // cellStyle: { textAlign: "center" },
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
    field: "updatedAt",
    headerName: "Last Updated",
  },
  {
    field: "actions",
    headerName: "Actions",
    width: 100,
    pinned: "right",
    sortable: false,
    suppressMovable: true,
    cellRenderer: actionCellRenderer,
  },
];

const gridOptions = {
  columnDefs,

  defaultColDef: {
    flex: 1,
    suppressHeaderMenuButton: true,
    // suppressHeaderFilterButton: true,
  },
  rowSelection,
  pagination: true,
  paginationPageSize: 20,
  onGridReady: (event) => {
    event.api.setFilterModel({
      shippingStatus: { values: ["IN_TRANSIT", "UNKNOWN", "PLANNED"] },
    });
  },
  getRowId: (params) => String(params.data.id),
};

export function getGridApi() {
  const gridDiv = document.querySelector("#grid");
  gridApi = agGrid.createGrid(gridDiv, gridOptions);

  return gridApi;
}

export function handleToolbar(gridApi) {
  const deleteSelectedBtn = document.getElementById("deleteSelectedBtn");
  deleteSelectedBtn.addEventListener("click", () =>
    deleteSelectedBtnEvent(gridApi),
  );

  const refreshGridBtn = document.getElementById("refreshGridBtn");
  refreshGridBtn.addEventListener("click", () => loadShipments(gridApi));
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
      gridApi.setGridOption("rowData", data.rows);
    })
    .catch((error) => {
      console.error("Error fetching data:", error);
      params.fail();
    });
}
