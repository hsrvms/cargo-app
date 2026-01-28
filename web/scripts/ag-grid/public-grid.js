import { createTruncatedRenderer } from "./grid.js";
import { ContainerDropdownRenderer } from "./container-dropdown-renderer.js";

let publicGridApi;

const publicColumnDefs = [
    {
        field: "shipmentNumber",
        headerName: "Shipment Number",
        filter: "agTextColumnFilter",
        maxWidth: 200,
        minWidth: 180,
        cellRenderer: (params) => {
            if (!params.value) return "";
            return `
            <button
                class="text-blue-600 dark:text-blue-300 hover:underline font-medium"
                onclick="openModalFetchPublicDetails('${params.data.id}')"
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
            const dateVal = params.data.shipmentDate;
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
        width: 140,
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
        width: 140,
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
        field: "sealineName",
        headerName: "Sealine",
        minWidth: 150,
        flex: 1,
        tooltipField: "sealine_name",
        cellRenderer: createTruncatedRenderer(),
    },
];

const gridOptions = {
    columnDefs: publicColumnDefs,
    defaultColDef: {
        sortable: true,
        filter: true,
        resizable: true,
        suppressMenu: true,
    },
    pagination: true,
    paginationPageSize: 20,
    domLayout: "autoHeight",
    suppressCellFocus: true,
    rowSelection: "single",

    editable: false,
    suppressClickEdit: true,
};

export function initPublicGrid(viewId, gridDivId = "#public-grid") {
    const gridDiv = document.querySelector(gridDivId);

    publicGridApi = agGrid.createGrid(gridDiv, gridOptions);

    loadPublicData(viewId);
}

function loadPublicData(viewId) {
    publicGridApi.showLoadingOverlay();

    fetch(`/api/shipments/public/${viewId}`)
        .then((response) => {
            if (!response.ok) throw new Error("Failed to load data");
            return response.json();
        })
        .then((data) => {
            const rowData = data.shipments || data;
            console.log("Data:", rowData);
            publicGridApi.setGridOption("rowData", rowData);
        })
        .catch((err) => {
            console.error(err);
            publicGridApi.showNoRowsOverlay();
        });
}

window.openPublicModal = function (shipmentId) {
    const modal = document.getElementById("shipment-modal");
    const modalBody = document.getElementById("modal-body");

    modal.classList.remove("hidden", "opacity-0");
    modalBody.innerHTML = '<div class="text-center p-4">Loading details...</div>';

    htmx.ajax("GET", `/track/shipment/${shipmentId}/public-details`, "#modal-body");
};

window.closeModal = function () {
    const modal = document.getElementById("shipment-modal");
    modal.classList.add("hidden", "opacity-0");
};
