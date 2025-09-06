export function initModalFunctions() {
  window.openModal = openModal;
  window.closeModal = closeModal;
  window.openModalFetchDetails = openModalFetchDetails;
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
