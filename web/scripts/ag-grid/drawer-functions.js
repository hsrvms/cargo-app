export function initDrawerFunctions() {
  window.openDrawer = openDrawer;
  window.closeDrawer = closeDrawer;
}

function openDrawer() {
  document
    .getElementById("shipment-drawer")
    .classList.remove("translate-x-full");
}

function closeDrawer() {
  document.getElementById("shipment-drawer").classList.add("translate-x-full");
}
