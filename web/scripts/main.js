import { initModalFunctions } from "./ag-grid/modal-functions.js";
import { handleNewShipment } from "./form/handle-new-shipment.js";
import { initEnhancedMap } from "./map/handle-map-enhanced.js";
import { mapDataService } from "./map/map-data-service.js";
import {
  getGridApi,
  handleToolbar,
  loadShipments,
} from "/scripts/ag-grid/grid.js";

function main() {
  const gridApi = getGridApi();
  document.addEventListener("DOMContentLoaded", async () => {
    // Initialize modal functions
    initModalFunctions();

    // Initialize toolbar
    handleToolbar(gridApi);

    // Initialize enhanced map with grid integration
    const map = initEnhancedMap(gridApi, {
      mapContainerId: "miniMap",
      standalone: false,
    });

    // Store map globally for debugging
    window.currentEnhancedMap = map;

    // Load initial shipments data
    loadShipments(gridApi);

    // Set up new shipment handler
    window.handleNewShipment = (e) => handleNewShipment(e, gridApi);

    // Add service status to global for debugging
    window.mapDataService = mapDataService;

    // Log initialization status
    // setTimeout(() => {
    //   const status = mapDataService.getStatus();
    //   console.log("✅ Cargo App initialized:", {
    //     mapInitialized: !!map,
    //     serviceConnected: status.connected,
    //     hasData: status.hasData,
    //   });
    // }, 1000);
  });
}

main();
