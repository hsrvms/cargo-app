import { handleShipmentForm } from "./add-shipment-form.js";
import { getGridApi, loadShipments } from "./ag-grid/grid.js";

function main() {
  document.addEventListener("DOMContentLoaded", () => {
    const gridApi = getGridApi();
    loadShipments(gridApi);
    handleShipmentForm(gridApi);
  });
}

main();
