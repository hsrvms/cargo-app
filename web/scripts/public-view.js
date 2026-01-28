import { initPublicGrid } from "./ag-grid/public-grid.js";
import { initModalFunctions } from "./ag-grid/modal-functions.js";

document.addEventListener("DOMContentLoaded", () => {
    initModalFunctions();

    const gridContainer = document.querySelector("#public-grid");

    if (!gridContainer) {
        console.error("Public Grid container #public-grid not found!");
        return;
    }

    const viewID = gridContainer.dataset.viewId;

    if (viewID) {
        console.log("Initializing Public Grid for View ID:", viewID);
        initPublicGrid(viewID);
    } else {
        console.error("No view ID found in data-view-id attribute");
    }
});
