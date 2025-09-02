let gripApi;

const rowSelection = {
  mode: "multiRow",
  checkboxes: true,
  enableClickSelection: true,
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
    filter: "agTextColumnFilter",
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
    // filter: "agSetColumnFilter" // Enterprise
    width: 120,
    cellRenderer: (params) => {
      const status = params.value || "Unknown";
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
    cellRenderer: (params) => {
      return `
				<div class="flex space-x-2 justify-around items-center w-full h-full">
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

const rowData = [
  {
    shipmentNumber: "Hello Mello",
    shipmentType: "CT",
    sealineCode: "MSCU",
  },
  {
    shipmentNumber: "Ola",
    shipmentType: "BK",
    sealineCode: "MSCU",
  },
  {
    shipmentNumber: "Obarey",
    shipmentType: "BL",
    sealineCode: "MSCU",
  },
];

const gridOptions = {
  rowData,
  columnDefs,
  rowSelection,
  pagination: true,
  paginationPageSize: 20,
};

document.addEventListener("DOMContentLoaded", () => {
  const gridDiv = document.querySelector("#grid");
  gridApi = agGrid.createGrid(gridDiv, gridOptions);

  // fetch("/");
});
