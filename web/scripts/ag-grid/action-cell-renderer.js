import { showToast } from "../toast.js";

export function actionCellRenderer(params) {
  const gridApi = params.api;

  const container = document.createElement("div");
  container.className =
    "flex space-x-2 justify-around items-center w-full h-full";

  const refreshBtn = document.createElement("button");
  refreshBtn.className = "text-blue-600 hover:text-blue-900 text-sm";
  refreshBtn.title = "Refresh";
  refreshBtn.innerHTML = `
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
		</svg>
  `;
  refreshBtn.addEventListener("click", () =>
    refreshShipment(gridApi, params.data.id),
  );

  const deleteBtn = document.createElement("button");
  deleteBtn.className = "text-red-600 hover:text-red-900 text-sm";
  deleteBtn.title = "Delete";
  deleteBtn.innerHTML = `
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
		</svg>
  `;
  deleteBtn.addEventListener("click", () => deleteShipment(params.data.id));

  container.appendChild(refreshBtn);
  container.appendChild(deleteBtn);

  return container;
}

function refreshShipment(gridApi, id) {
  console.log("Refreshing shipment:", id);

  fetch(`/api/shipments/${id}/refresh`, { method: "POST" })
    .then((response) => response.json)
    .then((data) => {
      gridApi.applyTransaction({ update: data.shipment });
      console.log("Shipment Refreshed");
      showToast("Shipment refreshed", "info");
    })
    .catch((err) => {
      console.log("Refresh error:", err);
      showToast("Error refreshing shipment", "error");
    });
}

function deleteShipment(id) {
  console.log("Shipment to delete:", id);
}
