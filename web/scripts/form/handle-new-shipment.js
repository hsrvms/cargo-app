import { showToast } from "../toast.js";

export function handleNewShipment(event, gridApi) {
  const { xhr } = event.detail;
  if (!event.detail.successful) {
    const responseJSON = JSON.parse(xhr.responseText);
    console.error("Failed to add shipment:", responseJSON.error);
    showToast(responseJSON.error, "error");
    return;
  }

  try {
    const response = JSON.parse(xhr.responseText);

    const newShipment = response.shipment || response;

    if (gridApi && newShipment) {
      gridApi.applyTransaction({ add: [newShipment] });
    }
    showToast("Shipment added successfully", "success");
  } catch (err) {
    console.error("Failed to add shipment:", err);
    showToast("Error adding shipment", "error");
  }
}
