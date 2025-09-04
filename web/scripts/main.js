import { handleNewShipment } from "./form/handle-new-shipment.js";
import {
  getGridApi,
  handleToolbar,
  loadShipments,
} from "/scripts/ag-grid/grid.js";

function main() {
  const gridApi = getGridApi();
  document.addEventListener("DOMContentLoaded", async () => {
    handleToolbar(gridApi);
    loadShipments(gridApi);
    window.handleNewShipment = (e) => handleNewShipment(e, gridApi);
  });
}

main();
