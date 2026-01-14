import { initEnhancedMap, updateMapWithShipments } from "./handle-map-enhanced.js";
import { mapDataService } from "./map-data-service.js";
import { initModalFunctions } from "../ag-grid/modal-functions.js";

class StandaloneMapController {
    constructor() {
        this.map = null;
        this.hasReceivedData = false;
        this.gridPageDetected = false;
        this.shipmentCount = 0;
        this.selectedCount = 0;
        this.config = {
            mapContainerId: "standaloneMap",
            standalone: true,
            initialView: [20, 0],
            initialZoom: 2,
            statusUpdateInterval: 2000,
            fallbackDelay: 2000,
            gridFallbackDelay: 4000,
        };
    }

    async init() {
        initModalFunctions();

        this.map = initEnhancedMap(null, {
            mapContainerId: this.config.mapContainerId,
            standalone: this.config.standalone,
            initialView: this.config.initialView,
            initialZoom: this.config.initialZoom,
        });

        if (!this.map) {
            console.error("❌ Failed to initialize map");
            this.updateAllStatusDisplays("error", "Failed to initialize map");
            return;
        }

        this.setupEventListeners();
        mapDataService.requestInitialData();
        this.setupDataFallback();
        this.startStatusUpdates();
    }

    setupEventListeners() {
        mapDataService.addEventListener("shipmentsUpdate", (data) => {
            this.shipmentCount = data.shipmentCount || data.visibleShipments?.length || 0;
            this.selectedCount = data.selectedCount || data.selectedShipments?.length || 0;

            updateMapWithShipments(data.visibleShipments || [], data.selectedShipments || []);
            this.hasReceivedData = true;
            this.updateAllStatusDisplays("connected");
        });

        mapDataService.addEventListener("initialDataReceived", (data) => {
            this.gridPageDetected = true;

            if (data.hasData && data.visibleShipments && data.visibleShipments.length > 0) {
                this.shipmentCount = data.visibleShipments.length;
                this.selectedCount = data.selectedShipments?.length || 0;

                updateMapWithShipments(data.visibleShipments, data.selectedShipments || []);
                this.hasReceivedData = true;
                this.updateAllStatusDisplays("connected");
            }
        });
    }

    setupDataFallback() {
        setTimeout(async () => {
            if (!this.hasReceivedData) {
                if (this.gridPageDetected) {
                    setTimeout(async () => {
                        if (!this.hasReceivedData) {
                            await this.loadShipmentsFromAPI();
                        }
                    }, this.config.gridFallbackDelay - this.config.fallbackDelay);
                } else {
                    await this.loadShipmentsFromAPI();
                }
            }
        }, this.config.fallbackDelay);
    }

    async loadShipmentsFromAPI() {
        this.updateAllStatusDisplays("loading", "Loading shipments...");

        try {
            const response = await fetch("/api/shipments/grid-data");

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            const shipments = data.rows || [];

            this.shipmentCount = shipments.length;
            this.selectedCount = 0;

            updateMapWithShipments(shipments, []);
            this.hasReceivedData = true;

            mapDataService.currentData.visibleShipments = shipments;
            mapDataService.currentData.selectedShipments = [];
            mapDataService.currentData.lastUpdate = new Date().toISOString();

            mapDataService.broadcastShipments(shipments, []);

            this.updateAllStatusDisplays("connected");
        } catch (error) {
            console.error("❌ Error loading shipments from API:", error);
            this.updateAllStatusDisplays("error", "Error loading data");
        }
    }

    updateAllStatusDisplays(status, customMessage = null) {
        this.updateNavbarStatus(status, customMessage);
        this.updateMapStatus(status, customMessage);
    }

    updateNavbarStatus(status, customMessage = null) {
        const connectionStatus = document.getElementById("connectionStatus");
        const dataStatus = document.getElementById("dataStatus");

        if (!connectionStatus || !dataStatus) {
            console.warn("Navbar status elements not found");
            return;
        }

        const indicator = connectionStatus.querySelector(".connection-indicator");
        const statusText = connectionStatus.querySelector(".status-text");
        const shipmentCount = dataStatus.querySelector(".shipment-count");

        let indicatorClass = "disconnected";
        let statusMessage = customMessage;
        let textColorClass = "text-red-600 dark:text-red-400";

        switch (status) {
            case "connected":
                indicatorClass = "connected";
                statusMessage = statusMessage || "Connected";
                textColorClass = "text-green-600 dark:text-green-400";
                if (shipmentCount) {
                    shipmentCount.textContent = this.shipmentCount;
                }
                break;

            case "loading":
                indicatorClass = "connecting";
                statusMessage = statusMessage || "Loading...";
                textColorClass = "text-blue-600 dark:text-blue-400";
                break;

            case "disconnected":
                indicatorClass = "disconnected";
                statusMessage = statusMessage || "Disconnected";
                textColorClass = "text-red-600 dark:text-red-400";
                if (shipmentCount) {
                    shipmentCount.textContent = "0";
                }
                break;

            case "error":
                indicatorClass = "disconnected";
                statusMessage = statusMessage || "Error";
                textColorClass = "text-red-600 dark:text-red-400";
                if (shipmentCount) {
                    shipmentCount.textContent = "0";
                }
                break;
        }

        if (indicator) {
            indicator.className = `connection-indicator ${indicatorClass}`;
        }

        if (statusText) {
            statusText.textContent = statusMessage;
            statusText.className = `status-text text-sm ${textColorClass}`;
        }
    }

    updateMapStatus(status, customMessage = null) {
        const mapStatusContent = document.getElementById("mapStatusContent");
        if (!mapStatusContent) return;

        let indicator = "disconnected";
        let message = customMessage;

        switch (status) {
            case "connected":
                indicator = "connected";
                if (!message) {
                    const shipmentText = this.shipmentCount === 1 ? "shipment" : "shipments";
                    const selectedText = this.selectedCount > 0 ? ` • ${this.selectedCount} selected` : "";
                    message = `${this.shipmentCount} ${shipmentText}${selectedText}`;
                }
                break;

            case "loading":
                indicator = "connecting";
                message = message || "Loading shipments...";
                break;

            case "disconnected":
                indicator = "disconnected";
                message = message || "No connection to grid data";
                break;

            case "error":
                indicator = "disconnected";
                message = message || "Error loading data";
                break;
        }

        mapStatusContent.innerHTML = `
      <span class="connection-indicator ${indicator}"></span>
      <span>${message}</span>
    `;
    }

    startStatusUpdates() {
        setInterval(() => {
            const status = mapDataService.getStatus();

            if (status.hasData) {
                this.shipmentCount = status.shipmentCount;
                this.selectedCount = status.selectedCount;
                this.hasReceivedData = true;
            }

            if (this.hasReceivedData && status.hasData) {
                this.updateAllStatusDisplays("connected");
            } else if (status.connected && !status.hasData) {
                this.updateAllStatusDisplays("loading");
            } else {
                this.updateAllStatusDisplays("disconnected");
            }
        }, this.config.statusUpdateInterval);
    }
}

document.addEventListener("DOMContentLoaded", () => {
    const controller = new StandaloneMapController();
    controller.init();
});

export { StandaloneMapController };
