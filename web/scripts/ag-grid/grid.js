import { actionCellRenderer } from "./action-cell-renderer.js";
import { deleteSelectedBtnEvent } from "./ag-grid-toolbar.js";
import { handleMap, updateMapMarkers } from "../map/handle-map.js";

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

  // Map update events
  onFilterChanged: (event) => {
    if (window.currentMap && window.currentMap._shipmentMarkers) {
      setTimeout(() => updateMapMarkers(window.currentMap, event.api), 100);
    }
  },

  onPaginationChanged: (event) => {
    if (window.currentMap && window.currentMap._shipmentMarkers) {
      setTimeout(() => updateMapMarkers(window.currentMap, event.api), 100);
    }
  },

  onSortChanged: (event) => {
    if (window.currentMap && window.currentMap._shipmentMarkers) {
      setTimeout(() => updateMapMarkers(window.currentMap, event.api), 100);
    }
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
      // Update map after data is loaded
      setTimeout(() => {
        const map = handleMap(gridApi);
        window.currentMap = map; // Store map globally for event handlers
      }, 100); // Small delay to ensure grid is fully rendered
    })
    .catch((error) => {
      console.error("Error fetching data:", error);
    });
}
