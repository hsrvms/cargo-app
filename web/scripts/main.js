import { initModalFunctions } from "./ag-grid/modal-functions.js";
import { handleNewShipment } from "./form/handle-new-shipment.js";
import { handleMap } from "./map/handle-map.js";
import {
  getGridApi,
  handleToolbar,
  loadShipments,
} from "/scripts/ag-grid/grid.js";

function main() {
  const gridApi = getGridApi();
  document.addEventListener("DOMContentLoaded", async () => {
    initModalFunctions();
    handleToolbar(gridApi);
    loadShipments(gridApi);
    window.handleNewShipment = (e) => handleNewShipment(e, gridApi);
  });
}

main();
