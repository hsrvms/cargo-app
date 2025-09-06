export function initDrawerFunctions() {
  window.openDrawer = openDrawer;
  window.closeDrawer = closeDrawer;
  window.openDrawerFetchDetails = openDrawerFetchDetails;
}

function openDrawer() {
  document
    .getElementById("shipment-drawer")
    .classList.remove("translate-x-full");
}

function closeDrawer() {
  document.getElementById("shipment-drawer").classList.add("translate-x-full");
}

function openDrawerFetchDetails(shipmentID) {
  openDrawer();
  htmx.ajax(
    "GET",
    `/api/shipments/${shipmentID}/details-html`,
    "#drawer-content",
  );
}
