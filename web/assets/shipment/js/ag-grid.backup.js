console.log("Ola");

let gridApi;
let selectedRows = [];

// Column definitions
const columnDefs = [
  {
    field: "select",
    headerName: "",
    checkboxSelection: true,
    headerCheckboxSelection: true,
    width: 50,
    pinned: "left",
  },
  {
    field: "shipment_number",
    headerName: "Shipment Number",
    sortable: true,
    filter: "agTextColumnFilter",
    width: 150,
  },
  {
    field: "shipment_type",
    headerName: "Type",
    sortable: true,
    filter: "agSetColumnFilter",
    width: 80,
  },
  {
    field: "sealine_code",
    headerName: "Sealine",
    sortable: true,
    filter: "agTextColumnFilter",
    width: 100,
  },
  {
    field: "sealine_name",
    headerName: "Sealine Name",
    sortable: true,
    filter: "agTextColumnFilter",
    width: 150,
  },
  {
    field: "shipping_status",
    headerName: "Status",
    sortable: true,
    filter: "agSetColumnFilter",
    width: 120,
    cellRenderer: function (params) {
      const status = params.value || "Unknown";
      const statusClasses = {
        "In Transit": "bg-blue-100 text-blue-800",
        Delivered: "bg-green-100 text-green-800",
        Delayed: "bg-yellow-100 text-yellow-800",
        Unknown: "bg-gray-100 text-gray-800",
      };
      const classes = statusClasses[status] || statusClasses["Unknown"];
      return `<span class="px-2 py-1 text-xs font-semibold rounded-full ${classes}">${status}</span>`;
    },
  },
  {
    field: "created_at",
    headerName: "Created",
    sortable: true,
    filter: "agDateColumnFilter",
    width: 120,
    cellRenderer: function (params) {
      return new Date(params.value).toLocaleDateString();
    },
  },
  {
    field: "actions",
    headerName: "Actions",
    width: 100,
    pinned: "right",
    cellRenderer: function (params) {
      return `
								<div class="flex space-x-2">
									<button
										onclick="refreshShipment('${params.data.id}')"
										class="text-blue-600 hover:text-blue-900 text-sm"
										title="Refresh"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
										</svg>
									</button>
									<button
										onclick="deleteShipment('${params.data.id}')"
										class="text-red-600 hover:text-red-900 text-sm"
										title="Delete"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
										</svg>
									</button>
								</div>
							`;
    },
  },
];

// Grid options
const gridOptions = {
  columnDefs: columnDefs,
  defaultColDef: {
    resizable: true,
    sortable: true,
    filter: true,
  },
  rowSelection: "multiple",
  suppressRowClickSelection: true,
  onSelectionChanged: onSelectionChanged,
  pagination: true,
  paginationPageSize: 20,
  rowModelType: "serverSide",
};

// Initialize the grid when the page loads
document.addEventListener("DOMContentLoaded", function () {
  const gridDiv = document.querySelector("#grid");
  agGrid.createGrid(gridDiv, gridOptions);

  // Set up server-side row model
  const datasource = {
    getRows: function (params) {
      fetch("/api/shipments/grid-data", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          startRow: params.startRow,
          endRow: params.endRow,
          sortModel: params.sortModel,
          filterModel: params.filterModel,
        }),
      })
        .then((response) => response.json())
        .then((data) => {
          params.success({
            rowData: data.rows,
            rowCount: data.lastRow,
          });
        })
        .catch((error) => {
          console.error("Error fetching data:", error);
          params.fail();
        });
    },
  };

  gridApi.setGridOption("serverSideDatasource", datasource);
});

// Selection changed handler
function onSelectionChanged() {
  selectedRows = gridApi.getSelectedRows();
}

// Refresh grid data
function refreshGrid() {
  gridApi.refreshServerSide();
}

// Refresh individual shipment
function refreshShipment(shipmentId) {
  fetch(`/api/shipments/${shipmentId}/refresh`, {
    method: "GET",
  })
    .then((response) => response.json())
    .then((data) => {
      if (data.message === "success") {
        gridApi.refreshServerSide();
        showMessage("Shipment refreshed successfully", "success");
      } else {
        showMessage("Failed to refresh shipment", "error");
      }
    })
    .catch((error) => {
      console.error("Error refreshing shipment:", error);
      showMessage("Failed to refresh shipment", "error");
    });
}

// Delete individual shipment
function deleteShipment(shipmentId) {
  if (confirm("Are you sure you want to delete this shipment?")) {
    fetch(`/api/shipments/${shipmentId}`, {
      method: "DELETE",
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.message === "success") {
          gridApi.refreshServerSide();
          showMessage("Shipment deleted successfully", "success");
        } else {
          showMessage("Failed to delete shipment", "error");
        }
      })
      .catch((error) => {
        console.error("Error deleting shipment:", error);
        showMessage("Failed to delete shipment", "error");
      });
  }
}

// Delete selected shipments
function deleteSelectedShipments() {
  if (selectedRows.length === 0) {
    showMessage("Please select shipments to delete", "warning");
    return;
  }

  if (
    confirm(
      `Are you sure you want to delete ${selectedRows.length} shipment(s)?`,
    )
  ) {
    const shipmentIds = selectedRows.map((row) => row.id);

    fetch("/api/shipments/bulk-delete", {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ shipment_ids: shipmentIds }),
    })
      .then((response) => response.json())
      .then((data) => {
        if (data.message === "success") {
          gridApi.refreshServerSide();
          showMessage(
            `${selectedRows.length} shipment(s) deleted successfully`,
            "success",
          );
          selectedRows = [];
        } else {
          showMessage("Failed to delete shipments", "error");
        }
      })
      .catch((error) => {
        console.error("Error deleting shipments:", error);
        showMessage("Failed to delete shipments", "error");
      });
  }
}

// Show message helper
function showMessage(message, type) {
  const messageDiv = document.getElementById("form-messages");
  const alertClasses = {
    success: "bg-green-100 border-green-500 text-green-700",
    error: "bg-red-100 border-red-500 text-red-700",
    warning: "bg-yellow-100 border-yellow-500 text-yellow-700",
  };

  messageDiv.innerHTML = `
						<div class="border-l-4 p-4 mt-4 ${alertClasses[type] || alertClasses["error"]}">
							<p>${message}</p>
						</div>
					`;

  // Auto-hide success messages
  if (type === "success") {
    setTimeout(() => {
      messageDiv.innerHTML = "";
    }, 3000);
  }
}

// HTMX event handlers
document.body.addEventListener("htmx:afterRequest", function (event) {
  const response = event.detail.xhr.response;
  let data;

  try {
    data = JSON.parse(response);
  } catch (e) {
    data = { error: "Invalid response" };
  }

  if (event.detail.successful) {
    showMessage(data.message || "Operation successful", "success");
  } else {
    showMessage(data.error || "Operation failed", "error");
  }
});
