import { actionCellRenderer } from "./action-cell-renderer.js";
import { deleteSelectedBtnEvent } from "./ag-grid-toolbar.js";
import { mapDataService } from "/scripts/map/map-data-service.js";
import { FilterManager } from "./filter-manager.js";
import { getVisibleShipments } from "./get-visible-shipments.js";
// Map integration now handled by enhanced map service

let gridApi;
let filterManager;

const rowSelection = {
  mode: "multiRow",
  // enableClickSelection: false,
};

// TODO:
// Correct Checkboxes
// Remove filter buttons on some columns
// Test it if its working correctly
const columnDefs = [
  {
    field: "shipmentNumber",
    headerName: "Shipment Number",
    filter: "agTextColumnFilter",
    width: 200,
    minWidth: 180,
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
    width: 120,
    minWidth: 100,
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
    minWidth: 150,
    filter: "agTextColumnFilter",
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
    minWidth: 150,
    filter: "agTextColumnFilter",
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
    width: 180,
    minWidth: 150,
    filter: "agSetColumnFilter",
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
    width: 120,
    minWidth: 100,
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
    minWidth: 100,
    filter: "agDateColumnFilter",
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
    field: "consignee",
    headerName: "Consignee",
    width: 150,
    minWidth: 120,
    filter: "agTextColumnFilter",
    editable: true,
    cellEditor: "agTextCellEditor",
    tooltipField: "consignee",
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
      updateShipmentField(params.data.id, "consignee", params.newValue || "");
    },
  },
  {
    field: "recipient",
    headerName: "Recipient",
    width: 150,
    minWidth: 120,
    filter: "agTextColumnFilter",
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
    field: "shipper",
    headerName: "Shipper",
    width: 150,
    minWidth: 120,
    filter: "agTextColumnFilter",
    editable: true,
    cellEditor: "agTextCellEditor",
    tooltipField: "shipper",
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
      updateShipmentField(params.data.id, "shipper", params.newValue || "");
    },
  },
  {
    field: "assignedTo",
    headerName: "Assigned To",
    width: 140,
    minWidth: 120,
    filter: "agTextColumnFilter",
    editable: true,
    cellEditor: "agTextCellEditor",
    tooltipField: "assignedTo",
    cellRenderer: (params) => {
      if (!params.value)
        return '<span class="text-gray-400 italic">Not assigned</span>';
      const value = params.value;
      if (value.length > 18) {
        return `<span title="${value}">${value.substring(0, 15)}...</span>`;
      }
      return value;
    },
    onCellValueChanged: (params) => {
      updateShipmentField(params.data.id, "assignedTo", params.newValue || "");
    },
  },
  {
    field: "placeOfLoading",
    headerName: "Place of Loading",
    width: 180,
    minWidth: 150,
    filter: "agTextColumnFilter",
    editable: true,
    cellEditor: "agTextCellEditor",
    tooltipField: "placeOfLoading",
    cellRenderer: (params) => {
      if (!params.value)
        return '<span class="text-gray-400 italic">Not specified</span>';
      const value = params.value;
      if (value.length > 25) {
        return `<span title="${value}">${value.substring(0, 22)}...</span>`;
      }
      return value;
    },
    onCellValueChanged: (params) => {
      updateShipmentField(
        params.data.id,
        "placeOfLoading",
        params.newValue || "",
      );
    },
  },
  {
    field: "placeOfDelivery",
    headerName: "Place of Delivery",
    width: 180,
    minWidth: 150,
    filter: "agTextColumnFilter",
    editable: true,
    cellEditor: "agTextCellEditor",
    tooltipField: "placeOfDelivery",
    cellRenderer: (params) => {
      if (!params.value)
        return '<span class="text-gray-400 italic">Not specified</span>';
      const value = params.value;
      if (value.length > 25) {
        return `<span title="${value}">${value.substring(0, 22)}...</span>`;
      }
      return value;
    },
    onCellValueChanged: (params) => {
      updateShipmentField(
        params.data.id,
        "placeOfDelivery",
        params.newValue || "",
      );
    },
  },
  {
    field: "finalDestination",
    headerName: "Final Destination",
    width: 200,
    minWidth: 150,
    filter: "agTextColumnFilter",
    editable: true,
    cellEditor: "agLargeTextCellEditor",
    cellEditorParams: {
      maxLength: 1000,
      rows: 3,
      cols: 40,
    },
    tooltipField: "finalDestination",
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
      updateShipmentField(
        params.data.id,
        "finalDestination",
        params.newValue || "",
      );
    },
  },
  {
    field: "containerType",
    headerName: "Container Type",
    width: 140,
    minWidth: 120,
    filter: "agSetColumnFilter",
    editable: true,
    cellEditor: "agSelectCellEditor",
    cellEditorParams: {
      values: [
        "20GP",
        "40GP",
        "40HC",
        "45HC",
        "20FR",
        "40FR",
        "20OT",
        "40OT",
        "20RF",
        "40RF",
      ],
    },
    tooltipField: "containerType",
    cellRenderer: (params) => {
      if (!params.value)
        return '<span class="text-gray-400 italic">Not specified</span>';
      return `<span class="px-2 py-1 text-xs font-mono bg-blue-100 text-blue-800 rounded">${params.value}</span>`;
    },
    onCellValueChanged: (params) => {
      updateShipmentField(
        params.data.id,
        "containerType",
        params.newValue || "",
      );
    },
  },
  {
    field: "mbl",
    headerName: "MBL",
    width: 140,
    minWidth: 100,
    filter: "agTextColumnFilter",
    editable: true,
    cellEditor: "agTextCellEditor",
    tooltipField: "mbl",
    cellRenderer: (params) => {
      if (!params.value)
        return '<span class="text-gray-400 italic">Not specified</span>';
      return `<span class="font-mono text-blue-700 bg-blue-50 px-2 py-1 rounded text-xs">${params.value}</span>`;
    },
    onCellValueChanged: (params) => {
      updateShipmentField(params.data.id, "mbl", params.newValue || "");
    },
  },
  {
    field: "customs",
    headerName: "Customs",
    width: 150,
    minWidth: 120,
    filter: "agTextColumnFilter",
    editable: true,
    cellEditor: "agTextCellEditor",
    tooltipField: "customs",
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
      updateShipmentField(params.data.id, "customs", params.newValue || "");
    },
  },
  {
    field: "invoiceAmount",
    headerName: "Invoice Amount",
    width: 130,
    minWidth: 110,
    filter: "agTextColumnFilter",
    editable: true,
    cellEditor: "agTextCellEditor",
    tooltipField: "invoiceAmount",
    cellRenderer: (params) => {
      if (!params.value)
        return '<span class="text-gray-400 italic">Not specified</span>';
      return `<span class="font-mono text-green-700">${params.value}</span>`;
    },
    onCellValueChanged: (params) => {
      updateShipmentField(
        params.data.id,
        "invoiceAmount",
        params.newValue || "",
      );
    },
  },
  {
    field: "cost",
    headerName: "Cost",
    width: 120,
    minWidth: 100,
    filter: "agTextColumnFilter",
    editable: true,
    cellEditor: "agTextCellEditor",
    tooltipField: "cost",
    cellRenderer: (params) => {
      if (!params.value)
        return '<span class="text-gray-400 italic">Not specified</span>';
      return `<span class="font-mono text-orange-700">${params.value}</span>`;
    },
    onCellValueChanged: (params) => {
      updateShipmentField(params.data.id, "cost", params.newValue || "");
    },
  },
  {
    field: "customsProcessed",
    headerName: "Customs ✓",
    width: 110,
    minWidth: 90,
    filter: false,
    editable: true,
    cellRenderer: "agCheckboxCellRenderer",
    cellEditor: "agCheckboxCellEditor",
    onCellValueChanged: (params) => {
      updateShipmentField(
        params.data.id,
        "customsProcessed",
        params.newValue === true,
      );
    },
  },
  {
    field: "invoiced",
    headerName: "Invoiced ✓",
    width: 110,
    minWidth: 90,
    filter: false,
    editable: true,
    cellRenderer: "agCheckboxCellRenderer",
    cellEditor: "agCheckboxCellEditor",
    onCellValueChanged: (params) => {
      updateShipmentField(params.data.id, "invoiced", params.newValue === true);
    },
  },
  {
    field: "paymentReceived",
    headerName: "Payment ✓",
    width: 110,
    minWidth: 90,
    filter: false,
    editable: true,
    cellRenderer: "agCheckboxCellRenderer",
    cellEditor: "agCheckboxCellEditor",
    onCellValueChanged: (params) => {
      updateShipmentField(
        params.data.id,
        "paymentReceived",
        params.newValue === true,
      );
    },
  },
  {
    field: "notes",
    headerName: "Notes",
    width: 250,
    minWidth: 150,
    filter: "agTextColumnFilter",
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
    minWidth: 80,
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
    filter: true,
    suppressHeaderFilterButton: false,
    minWidth: 120, // Minimum width for readability
  },
  rowSelection,
  pagination: true,
  paginationPageSize: 20, // Increased due to more columns
  paginationPageSizeSelector: [20, 50, 100],

  // Performance optimizations for richer data
  rowBuffer: 15,
  suppressCellFocus: true,
  animateRows: true,

  // Handle large datasets better with many columns
  suppressColumnVirtualisation: false,
  suppressRowVirtualisation: false,

  // Ensure single horizontal scrollbar
  suppressHorizontalScroll: false,
  alwaysShowHorizontalScroll: false,
  suppressScrollOnNewData: true,

  // Prevent auto-sizing to viewport width
  suppressSizeToFit: true,

  // Allow columns to maintain their natural widths
  enableColResize: true,

  // Fix scrollbar synchronization
  scrollbarWidth: 17,
  suppressMiddleClickScrolls: false,
  // Column management
  sideBar: {
    toolPanels: [
      {
        id: "columns",
        labelDefault: "Columns",
        labelKey: "columns",
        iconKey: "columns",
        toolPanel: "agColumnsToolPanel",
        toolPanelParams: {
          suppressRowGroups: true,
          suppressValues: true,
          suppressPivots: true,
          suppressPivotMode: true,
          suppressColumnFilter: false,
          suppressColumnSelectAll: false,
          suppressColumnExpandAll: false,
        },
      },
      {
        id: "filters",
        labelDefault: "Filters",
        labelKey: "filters",
        iconKey: "filter",
        toolPanel: "agFiltersToolPanel",
      },
    ],
    defaultToolPanel: "columns",
  },

  // Enable horizontal scrolling - don't auto-size columns
  onFirstDataRendered: (params) => {
    // Hide some columns by default to reduce initial width but keep essential ones visible
    const columnsToHide = [
      "customs",
      "mbl",
      "placeOfLoading",
      "placeOfDelivery",
      "containerType",
    ];
    params.api.setColumnsVisible(columnsToHide, false);
  },

  onGridReady: (event) => {
    event.api.setFilterModel({
      shippingStatus: { values: ["IN_TRANSIT", "UNKNOWN", "PLANNED"] },
    });

    // Initialize filter manager after grid is ready
    if (!filterManager) {
      try {
        filterManager = new FilterManager(event.api);
        window.filterManager = filterManager;
      } catch (error) {
        console.error("Failed to initialize filter manager:", error);
      }
    }
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

  // Make gridApi globally accessible
  window.gridApi = gridApi;

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

  // Filter manager is now initialized in onGridReady callback
  // This ensures the grid API is fully ready with all methods available
}

// Export filter manager for external access
export function getFilterManager() {
  return filterManager;
}

// Initialize filter manager manually if needed
export function initializeFilterManager(gridApi) {
  if (!gridApi) {
    console.error("Grid API required for filter manager initialization");
    return null;
  }

  if (!filterManager) {
    try {
      filterManager = new FilterManager(gridApi);
      window.filterManager = filterManager;
      console.log("Filter manager manually initialized");
      return filterManager;
    } catch (error) {
      console.error("Failed to manually initialize filter manager:", error);
      return null;
    }
  }

  return filterManager;
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
      // Wait for grid to process the data, then broadcast visible shipments
      setTimeout(() => {
        const visibleShipments = getVisibleShipments(gridApi, { debug: false });
        mapDataService.broadcastShipments(visibleShipments, []);
        console.log(
          `📡 Broadcasted ${visibleShipments.length} visible shipments to map service`,
        );
      }, 100); // Small delay to ensure grid has processed the data
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

  // Map field names from ag-grid to API expectations if needed
  const fieldMapping = {
    // Most fields use the same name, but we can add exceptions here if needed
    // Example: "gridFieldName": "apiFieldName"
  };

  const apiField = fieldMapping[field] || field;

  // Create payload with ONLY the changed field
  const payload = {};

  // Handle different field types properly
  if (typeof value === "boolean") {
    payload[apiField] = value;
  } else if (value === null || value === undefined) {
    payload[apiField] = "";
  } else {
    payload[apiField] = String(value).trim();
  }

  console.log(
    `Updating ONLY ${apiField} for shipment ${shipmentId} with value:`,
    payload[apiField],
  );

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

        // Revert the change in the grid on error
        if (window.gridApi) {
          loadShipments(window.gridApi); // Reload data to get correct values
        }
      } else {
        console.log(
          `Successfully updated ${apiField} for shipment ${shipmentId}`,
        );
        // You could add success toast notification here if available
        // showToast(`${field} updated successfully`, "success");
      }
    })
    .catch((error) => {
      console.error("Error updating field:", error);
      // You could add toast notification here if available
      // showToast("Failed to update " + field, "error");

      // Revert the change in the grid on error
      if (window.gridApi) {
        loadShipments(window.gridApi); // Reload data to get correct values
      }
    });
}
