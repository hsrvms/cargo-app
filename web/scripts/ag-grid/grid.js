import { actionCellRenderer } from "./action-cell-renderer.js";
import { deleteSelectedBtnEvent } from "./ag-grid-toolbar.js";
import { mapDataService } from "/scripts/map/map-data-service.js";
// Map integration now handled by enhanced map service

let gridApi;

const rowSelection = {
  mode: "multiRow",
  checkboxes: true,
  // enableClickSelection: false,
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
          onclick="openModalFetchDetails('${params.data.id}')"
        >
          ${params.value} - ${params.data.shipmentType}
        </button>
      `;
    },
  },
  {
    field: "shippingStatus",
    headerName: "Status",
    filter: "agSetColumnFilter",
    maxWidth: 120,
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
    field: "originPort",
    headerName: "Origin",
    width: 200,
    cellRenderer: (params) => {
      const pol = params.data?.route?.pol;
      if (pol && pol.location) {
        return `${pol.location.name} (${pol.location.locode})`;
      }
      return "N/A";
    },
  },
  {
    field: "destinationPort",
    headerName: "Destination",
    width: 200,
    cellRenderer: (params) => {
      const pod = params.data?.route?.pod;
      if (pod && pod.location) {
        return `${pod.location.name} (${pod.location.locode})`;
      }
      return "N/A";
    },
  },
  {
    field: "vesselInfo",
    headerName: "Vessel",
    maxWidth: 180,
    cellRenderer: (params) => {
      const vessels = params.data?.vessels;
      if (vessels && vessels.length > 0) {
        const vessel = vessels[0];
        return `${vessel.name || "Unknown"} (${vessel.imo || "N/A"})`;
      }
      return "N/A";
    },
  },
  {
    field: "containerCount",
    headerName: "Containers",
    maxWidth: 120,
    cellRenderer: (params) => {
      const containers = params.data?.containers;
      if (containers && containers.length > 0) {
        return containers.length.toString();
      }
      return "0";
    },
  },
  {
    field: "nextETA",
    headerName: "Next ETA",
    width: 120,
    cellRenderer: (params) => {
      const route = params.data?.route;
      if (!route) return "N/A";

      // Find next port with a future date
      const ports = [route.prepol, route.pol, route.pod, route.postpod].filter(
        Boolean,
      );
      const now = new Date();

      for (const port of ports) {
        if (port.date && new Date(port.date) > now) {
          const date = new Date(port.date);
          return date.toLocaleDateString("en-US", {
            month: "short",
            day: "numeric",
          });
        }
      }

      return "N/A";
    },
  },
  {
    field: "recipient",
    headerName: "Recipient",
    width: 150,
    editable: true,
    cellEditor: "agTextCellEditor",
    tooltipField: "recipient",
    cellRenderer: (params) => {
      if (!params.value)
        return '<span class="text-gray-400 italic">Not specified</span>';
      const value = params.value;
      if (value.length > 20) {
        return `<span title="${value}">${value.substring(0, 17)}...</span>`;
      }
      return value;
    },
    onCellValueChanged: (params) => {
      updateShipmentField(params.data.id, "recipient", params.newValue || "");
    },
  },
  {
    field: "address",
    headerName: "Address",
    width: 200,
    editable: true,
    cellEditor: "agTextCellEditor",
    tooltipField: "address",
    cellRenderer: (params) => {
      if (!params.value)
        return '<span class="text-gray-400 italic">Not specified</span>';
      const value = params.value;
      if (value.length > 30) {
        return `<span title="${value}">${value.substring(0, 27)}...</span>`;
      }
      return value;
    },
    onCellValueChanged: (params) => {
      updateShipmentField(params.data.id, "address", params.newValue || "");
    },
  },
  {
    field: "notes",
    headerName: "Notes",
    width: 200,
    editable: true,
    cellEditor: "agLargeTextCellEditor",
    cellEditorParams: {
      maxLength: 2000,
      rows: 4,
      cols: 50,
    },
    tooltipField: "notes",
    cellRenderer: (params) => {
      if (!params.value)
        return '<span class="text-gray-400 italic">No notes</span>';
      const value = params.value;
      if (value.length > 50) {
        return `<span title="${value}" class="cursor-help">${value.substring(0, 47)}...</span>`;
      }
      return value;
    },
    onCellValueChanged: (params) => {
      updateShipmentField(params.data.id, "notes", params.newValue || "");
    },
  },
  {
    field: "actions",
    headerName: "Actions",
    width: 100,
    pinned: "right",
    sortable: false,
    resizable: false,
    suppressMovable: true,
    cellRenderer: actionCellRenderer,
  },
];

const gridOptions = {
  columnDefs,

  defaultColDef: {
    suppressHeaderMenuButton: true,
    resizable: true,
    sortable: true,
    // filter: true,
    // suppressHeaderFilterButton: true,
  },
  rowSelection,
  pagination: true,
  paginationPageSize: 15, // Reduced due to richer data per row
  paginationPageSizeSelector: [15, 25, 50],

  // Performance optimizations for richer data
  rowBuffer: 10,
  suppressCellFocus: true,
  animateRows: true,

  // Handle large datasets better
  suppressColumnVirtualisation: false,
  suppressRowVirtualisation: false,

  onGridReady: (event) => {
    event.api.setFilterModel({
      shippingStatus: { values: ["IN_TRANSIT", "UNKNOWN", "PLANNED"] },
    });
  },

  onSelectionChanged: (event) => {
    const selectedRows = event.api.getSelectedRows();
    // Broadcast selection changes to map service
    mapDataService.broadcastSelection(selectedRows);
  },

  // Map integration now handled by MapDataService in enhanced map
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
  refreshGridBtn.addEventListener("click", () => {
    console.log("Refreshing grid data and updating map...");
    loadShipments(gridApi);
  });
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
      // Broadcast data to MapDataService for map synchronization
      mapDataService.broadcastShipments(data.rows, []);
    })
    .catch((error) => {
      console.error("Error fetching data:", error);
    });
}

// Function to update shipment fields
function updateShipmentField(shipmentId, field, value) {
  if (!shipmentId) {
    console.error("Invalid shipment ID");
    return;
  }

  const payload = {};
  payload[field] = value || "";

  fetch(`/api/shipments/${shipmentId}/update-info`, {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  })
    .then((response) => {
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      return response.json();
    })
    .then((data) => {
      if (data.error) {
        console.error("Error updating field:", data.error);
        // You could add toast notification here if available
        // showToast("Error: " + data.error, "error");
      } else {
        console.log(`Successfully updated ${field} for shipment ${shipmentId}`);
        // You could add success toast notification here if available
        // showToast(`${field} updated successfully`, "success");
      }
    })
    .catch((error) => {
      console.error("Error updating field:", error);
      // You could add toast notification here if available
      // showToast("Failed to update " + field, "error");

      // Optionally revert the change in the grid
      // This would require storing the original value before the edit
    });
}
