import { actionCellRenderer } from "./action-cell-renderer.js";
import { ContainerDropdownRenderer } from "./container-dropdown-renderer.js";
import { deleteSelectedBtnEvent } from "./ag-grid-toolbar.js";
import { mapDataService } from "/scripts/map/map-data-service.js";
import { FilterManager } from "./filter-manager.js";
import { getVisibleShipments } from "./get-visible-shipments.js";

let gridApi;
let filterManager;

const rowSelection = {
    mode: "multiRow",
    // enableClickSelection: false,
};

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
        minWidth: 120,
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
        field: "shipmentDate",
        headerName: "Shipment Date",
        width: 160,
        filter: "agDateColumnFilter",
        valueGetter: (params) => {
            const dateVal = params.data.containers[0]?.events[0]?.date;
            return dateVal ? new Date(dateVal) : null;
        },
        cellRenderer: (params) => {
            if (!params.value) return '<span class="text-gray-400 italic">N/A</span>';

            return params.value.toLocaleDateString("en-US", {
                year: "numeric",
                month: "short",
                day: "numeric",
            });
        },
    },
    {
        field: "originPort",
        headerName: "Origin",
        minWidth: 180,
        flex: 1,
        filter: "agTextColumnFilter",
        valueGetter: (params) => {
            const pol = params.data?.route?.pol;
            if (pol && pol.location) {
                return `${pol.location.name} (${pol.location.locode})`;
            }
            return "N/A";
        },
        cellRenderer: createTruncatedRenderer(),
    },
    {
        field: "destinationPort",
        headerName: "Destination",
        minWidth: 180,
        flex: 1,
        filter: "agTextColumnFilter",
        valueGetter: (params) => {
            const pod = params.data?.route?.pod;
            if (pod && pod.location) {
                return `${pod.location.name} (${pod.location.locode})`;
            }
            return "N/A";
        },
        cellRenderer: createTruncatedRenderer(),
    },
    {
        field: "vesselInfo",
        headerName: "Vessel",
        minWidth: 180,
        flex: 1,
        filter: "agSetColumnFilter",
        valueGetter: (params) => {
            const vessels = params.data?.vessels;
            if (vessels && vessels.length > 0) {
                const vessel = vessels[0];
                return `${vessel.name || "Unknown"} (${vessel.imo || "N/A"})`;
            }
            return "N/A";
        },
        cellRenderer: createTruncatedRenderer(),
    },
    {
        field: "containerCount",
        headerName: "Containers",
        width: 120,
        filter: "agTextColumnFilter",
        filterValueGetter: (params) => {
            const containers = params.data?.containers;
            if (containers && containers.length > 0) {
                return containers.map((c) => c.number || c.name || c.id || "").join(" ");
            }
            return "";
        },
        valueGetter: (params) => {
            const containers = params.data?.containers;
            return containers && containers.length > 0 ? containers.length : 0;
        },
        cellStyle: { overflow: "visible" },
        cellRenderer: ContainerDropdownRenderer,
    },
    {
        field: "nextETA",
        headerName: "Next ETA",
        width: 100,
        filter: "agDateColumnFilter",
        valueGetter: (params) => {
            const route = params.data?.route;
            if (!route) return null;

            const ports = [route.prepol, route.pol, route.pod, route.postpod].filter(Boolean);
            const now = new Date();

            for (const port of ports) {
                if (port.date && new Date(port.date) > now) {
                    return new Date(port.date);
                }
            }

            return null;
        },
        cellRenderer: (params) => {
            if (!params.value) return "N/A";
            const date = new Date(params.value);
            return date.toLocaleDateString("en-US", {
                month: "short",
                day: "numeric",
            });
        },
    },
    {
        field: "consignee",
        headerName: "Consignee",
        minWidth: 180,
        flex: 1,
        filter: "agTextColumnFilter",
        editable: true,
        cellEditor: "agLargeTextCellEditor",
        cellEditorParams: {
            maxLength: 1000,
            rows: 3,
            cols: 40,
        },
        tooltipField: "consignee",
        cellRenderer: createTruncatedRenderer(),
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "consignee", params.newValue || "");
        },
    },
    {
        field: "recipient",
        headerName: "Recipient",
        minWidth: 180,
        flex: 1,
        filter: "agTextColumnFilter",
        editable: true,
        cellEditor: "agLargeTextCellEditor",
        cellEditorParams: {
            maxLength: 1000,
            rows: 3,
            cols: 40,
        },
        tooltipField: "recipient",
        cellRenderer: createTruncatedRenderer(),
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "recipient", params.newValue || "");
        },
    },
    {
        field: "shipper",
        headerName: "Shipper",
        minWidth: 180,
        flex: 1,
        filter: "agTextColumnFilter",
        editable: true,
        cellEditor: "agLargeTextCellEditor",
        cellEditorParams: {
            maxLength: 1000,
            rows: 3,
            cols: 40,
        },
        tooltipField: "shipper",
        cellRenderer: createTruncatedRenderer(),
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "shipper", params.newValue || "");
        },
    },
    {
        field: "assignedTo",
        headerName: "Assigned To",
        minWidth: 180,
        flex: 1,
        filter: "agTextColumnFilter",
        editable: true,
        cellEditor: "agLargeTextCellEditor",
        cellEditorParams: {
            maxLength: 1000,
            rows: 3,
            cols: 40,
        },
        tooltipField: "assignedTo",
        cellRenderer: createTruncatedRenderer(),
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "assignedTo", params.newValue || "");
        },
    },
    {
        field: "placeOfLoading",
        headerName: "Place of Loading",
        minWidth: 180,
        flex: 1,
        filter: "agTextColumnFilter",
        editable: true,
        cellEditor: "agLargeTextCellEditor",
        cellEditorParams: {
            maxLength: 1000,
            rows: 3,
            cols: 40,
        },
        tooltipField: "placeOfLoading",
        cellRenderer: createTruncatedRenderer(),
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "placeOfLoading", params.newValue || "");
        },
    },
    {
        field: "placeOfDelivery",
        headerName: "Place of Delivery",
        minWidth: 180,
        flex: 1,
        filter: "agTextColumnFilter",
        editable: true,
        cellEditor: "agLargeTextCellEditor",
        cellEditorParams: {
            maxLength: 1000,
            rows: 3,
            cols: 40,
        },
        tooltipField: "placeOfDelivery",
        cellRenderer: createTruncatedRenderer(),
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "placeOfDelivery", params.newValue || "");
        },
    },
    {
        field: "finalDestination",
        headerName: "Final Destination",
        minWidth: 180,
        flex: 1,
        filter: "agTextColumnFilter",
        editable: true,
        cellEditor: "agLargeTextCellEditor",
        cellEditorParams: {
            maxLength: 1000,
            rows: 3,
            cols: 40,
        },
        tooltipField: "finalDestination",
        cellRenderer: createTruncatedRenderer(),
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "finalDestination", params.newValue || "");
        },
    },
    {
        field: "containerType",
        headerName: "Container Type",
        minWidth: 120,
        filter: "agSetColumnFilter",
        editable: true,
        cellEditor: "agSelectCellEditor",
        cellEditorParams: {
            values: ["20GP", "40GP", "40HC", "45HC", "20FR", "40FR", "20OT", "40OT", "20RF", "40RF"],
        },
        tooltipField: "containerType",
        cellRenderer: (params) => {
            if (!params.value) return '<span class="text-gray-400 italic">N/A</span>';
            return `<span class="px-2 py-1 text-xs font-mono bg-blue-100 text-blue-800 rounded">${params.value}</span>`;
        },
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "containerType", params.newValue || "");
        },
    },
    {
        field: "mbl",
        headerName: "MBL",
        minWidth: 100,
        filter: "agTextColumnFilter",
        editable: true,
        cellEditor: "agLargeTextCellEditor",
        cellEditorParams: {
            maxLength: 1000,
            rows: 3,
            cols: 40,
        },
        tooltipField: "mbl",
        cellRenderer: (params) => {
            if (!params.value) return '<span class="text-gray-400 italic">N/A</span>';
            return `<span class="font-mono text-blue-700 bg-blue-50 px-2 py-1 rounded text-xs">${params.value}</span>`;
        },
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "mbl", params.newValue || "");
        },
    },
    {
        field: "customs",
        headerName: "Customs",
        minWidth: 120,
        filter: "agTextColumnFilter",
        editable: true,
        cellEditor: "agLargeTextCellEditor",
        cellEditorParams: {
            maxLength: 1000,
            rows: 3,
            cols: 40,
        },
        tooltipField: "customs",
        cellRenderer: createTruncatedRenderer(),
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "customs", params.newValue || "");
        },
    },
    {
        field: "invoiceAmount",
        headerName: "Invoice Amount",
        minWidth: 110,
        filter: "agTextColumnFilter",
        editable: true,
        cellEditor: "agTextCellEditor",
        tooltipField: "invoiceAmount",
        cellRenderer: createTruncatedRenderer(),
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "invoiceAmount", params.newValue || "");
        },
    },
    {
        field: "cost",
        headerName: "Cost",
        width: 100,
        filter: "agTextColumnFilter",
        editable: true,
        cellEditor: "agTextCellEditor",
        tooltipField: "cost",
        cellRenderer: (params) => {
            if (!params.value) return '<span class="text-gray-400 italic">N/A</span>';
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
        filter: false,
        editable: true,
        cellRenderer: "agCheckboxCellRenderer",
        cellEditor: "agCheckboxCellEditor",
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "customsProcessed", params.newValue === true);
        },
    },
    {
        field: "invoiced",
        headerName: "Invoiced ✓",
        width: 110,
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
        filter: false,
        editable: true,
        cellRenderer: "agCheckboxCellRenderer",
        cellEditor: "agCheckboxCellEditor",
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "paymentReceived", params.newValue === true);
        },
    },
    {
        field: "notes",
        headerName: "Notes",
        minWidth: 150,
        flex: 1.5,
        filter: "agTextColumnFilter",
        editable: true,
        cellEditor: "agLargeTextCellEditor",
        cellEditorParams: {
            maxLength: 2000,
            rows: 4,
            cols: 50,
        },
        tooltipField: "notes",
        cellRenderer: createTruncatedRenderer('<span class="text-gray-400 italic">No notes</span>'),
        onCellValueChanged: (params) => {
            updateShipmentField(params.data.id, "notes", params.newValue || "");
        },
    },
    {
        field: "actions",
        headerName: "Actions",
        pinned: "right",
        minWidth: 80,
        maxWidth: 100,
        sortable: false,
        resizable: false,
        filter: false,
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
    },
    rowSelection,
    pagination: true,
    paginationPageSize: 50,
    paginationPageSizeSelector: [20, 50, 100],
    domLayout: "autoHeight",
    rowBuffer: 15,
    suppressCellFocus: true,
    animateRows: true,
    suppressColumnVirtualisation: false,
    suppressRowVirtualisation: false,
    suppressHorizontalScroll: false,
    alwaysShowHorizontalScroll: false,
    suppressScrollOnNewData: true,
    scrollbarWidth: 17,
    suppressMiddleClickScrolls: false,
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
    },

    // Enable horizontal scrolling - don't auto-size columns
    onFirstDataRendered: (params) => {
        const columnsToHide = ["customs", "mbl", "placeOfLoading", "placeOfDelivery", "containerType"];
        params.api.setColumnsVisible(columnsToHide, false);
    },

    onGridReady: (event) => {
        event.api.setFilterModel({
            shippingStatus: { values: ["IN_TRANSIT", "UNKNOWN", "PLANNED"] },
        });

        if (!filterManager) {
            try {
                filterManager = new FilterManager(event.api);
                window.filterManager = filterManager;
            } catch (error) {
                console.error("Failed to initialize filter manager:", error);
            }
        }
    },

    // Broadcast selection changes to map service
    onSelectionChanged: (event) => {
        const selectedRows = event.api.getSelectedRows();
        mapDataService.broadcastSelection(selectedRows);
    },

    getRowId: (params) => String(params.data.id),
};

export function getGridApi() {
    const gridDiv = document.querySelector("#grid");
    gridApi = agGrid.createGrid(gridDiv, gridOptions);

    window.gridApi = gridApi;

    return gridApi;
}

export function handleToolbar(gridApi) {
    const deleteSelectedBtn = document.getElementById("deleteSelectedBtn");
    deleteSelectedBtn.addEventListener("click", () => deleteSelectedBtnEvent(gridApi));

    const refreshGridBtn = document.getElementById("refreshGridBtn");
    refreshGridBtn.addEventListener("click", () => {
        loadShipments(gridApi);
    });
}

export function getFilterManager() {
    return filterManager;
}

export function initializeFilterManager(gridApi) {
    if (!gridApi) {
        console.error("Grid API required for filter manager initialization");
        return null;
    }

    if (!filterManager) {
        try {
            filterManager = new FilterManager(gridApi);
            window.filterManager = filterManager;
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
            console.log("GridData:", data);
            gridApi.setGridOption("rowData", data.rows);
            setTimeout(() => {
                const visibleShipments = getVisibleShipments(gridApi, { debug: false });
                mapDataService.broadcastShipments(visibleShipments, []);
            }, 100);
        })
        .catch((error) => {
            console.error("Error fetching data:", error);
        });
}

function updateShipmentField(shipmentId, field, value) {
    if (!shipmentId) {
        console.error("Invalid shipment ID");
        return;
    }

    const fieldMapping = {
        // Most fields use the same name, but we can add exceptions here if needed
        // Example: "gridFieldName": "apiFieldName"
    };

    const apiField = fieldMapping[field] || field;

    const payload = {};

    if (typeof value === "boolean") {
        payload[apiField] = value;
    } else if (value === null || value === undefined) {
        payload[apiField] = "";
    } else {
        payload[apiField] = String(value).trim();
    }

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
                if (window.gridApi) {
                    loadShipments(window.gridApi);
                }
            } else {
                // You could add success toast notification here if available
                // showToast(`${field} updated successfully`, "success");
            }
        })
        .catch((error) => {
            console.error("Error updating field:", error);
            // You could add toast notification here if available
            // showToast("Failed to update " + field, "error");

            if (window.gridApi) {
                loadShipments(window.gridApi);
            }
        });
}

/**
 * Generates a cell renderer that truncates text with CSS and shows a tooltip.
 * @param {string} placeholder - Text to show if value is null/empty (default: "N/A")
 */
function createTruncatedRenderer(placeholder = '<span class="text-gray-400 italic">N/A</span>') {
    return (params) => {
        if (!params.value) return placeholder;

        return `
              <span
                title="${params.value}"
                class="block w-full truncate cursor-help"
              >
                ${params.value}
              </span>
            `;
    };
}
