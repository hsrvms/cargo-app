export function initModalFunctions() {
  window.openModal = openModal;
  window.closeModal = closeModal;
  window.openModalFetchDetails = openModalFetchDetails;
  window.openFilterManagementPanel = openFilterManagementPanel;
  window.closeFilterManagementPanel = closeFilterManagementPanel;
}

function openModal() {
  const modal = document.getElementById("shipment-modal");
  const modalContent = document.getElementById("modal-content");

  modal.classList.remove("hidden");

  // Add animation classes for smooth entrance
  setTimeout(() => {
    modal.classList.remove("opacity-0");
    modalContent.classList.remove("scale-95", "opacity-0");
    modalContent.classList.add("scale-100", "opacity-100");
  }, 10);
}

function closeModal() {
  const modal = document.getElementById("shipment-modal");
  const modalContent = document.getElementById("modal-content");

  // Add exit animation classes
  modal.classList.add("opacity-0");
  modalContent.classList.remove("scale-100", "opacity-100");
  modalContent.classList.add("scale-95", "opacity-0");

  // Hide modal after animation completes
  setTimeout(() => {
    modal.classList.add("hidden");
  }, 300);
}

function openModalFetchDetails(shipmentID) {
  openModal();
  htmx.ajax("GET", `/api/shipments/${shipmentID}/details-html`, "#modal-body");
}

function openFilterManagementPanel() {
  const panel = document.getElementById("filter-management-panel");
  const panelContent = document.getElementById("filter-panel-content");

  if (!panel || !panelContent) {
    console.error("Filter management panel elements not found");
    return;
  }

  panel.classList.remove("hidden");

  // Add animation classes for smooth entrance
  setTimeout(() => {
    panel.classList.remove("opacity-0");
    panelContent.classList.remove("scale-95", "opacity-0");
    panelContent.classList.add("scale-100", "opacity-100");
  }, 10);

  // Trigger filter manager to render the panel content if available
  if (
    window.filterManager &&
    typeof window.filterManager.renderFilterList === "function"
  ) {
    window.filterManager.renderFilterList();
  }
}

function closeFilterManagementPanel() {
  const panel = document.getElementById("filter-management-panel");
  const panelContent = document.getElementById("filter-panel-content");

  if (!panel || !panelContent) return;

  // Add exit animation classes
  panel.classList.add("opacity-0");
  panelContent.classList.remove("scale-100", "opacity-100");
  panelContent.classList.add("scale-95", "opacity-0");

  // Hide panel after animation completes
  setTimeout(() => {
    panel.classList.add("hidden");
  }, 300);
}
