import { initModalFunctions } from "./ag-grid/modal-functions.js";
import { handleNewShipment } from "./form/handle-new-shipment.js";
import { initEnhancedMap } from "./map/handle-map-enhanced.js";
import { mapDataService } from "./map/map-data-service.js";
import { PublicViewManager } from "./ag-grid/public-view-manager.js";
import { getGridApi, handleToolbar, loadShipments } from "/scripts/ag-grid/grid.js";

function main() {
    const gridApi = getGridApi();
    document.addEventListener("DOMContentLoaded", async () => {
        initModalFunctions();

        handleToolbar(gridApi);

        const map = initEnhancedMap(gridApi, {
            mapContainerId: "miniMap",
            standalone: false,
        });

        const publicViewManager = new PublicViewManager(gridApi);

        const shareBtn = document.getElementById("createPublicViewBtn");
        if (shareBtn) {
            shareBtn.addEventListener("click", () => {
                publicViewManager.createFromSelection();
            });
        }

        window.currentEnhancedMap = map;

        loadShipments(gridApi);

        window.handleNewShipment = (e) => handleNewShipment(e, gridApi);

        window.mapDataService = mapDataService;
    });
}

main();
