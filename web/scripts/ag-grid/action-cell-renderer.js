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
  deleteBtn.addEventListener("click", () =>
    deleteShipment(gridApi, params.data.id),
  );

  const detailsBtn = document.createElement("button");
  detailsBtn.className = "text-yellow-400 hover:text-yellow-800 text-sm";
  detailsBtn.title = "View Details";
  detailsBtn.innerHTML = `
    <svg
      xmlns="http://www.w3.org/2000/svg"
      class="w-4 h-4"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
      <polyline points="14 2 14 8 20 8" />
      <line x1="16" y1="13" x2="8" y2="13" />
      <line x1="16" y1="17" x2="8" y2="17" />
      <line x1="10" y1="9" x2="8" y2="9" />
    </svg>
    `;
  detailsBtn.addEventListener("click", () => {
    openDrawer();
    htmx.ajax(
      "GET",
      `/api/shipments/${params.data.id}/details-html`,
      "#drawer-content",
    );
    // window.location.href = `/shipments/${params.data.id}`;
  });

  container.appendChild(refreshBtn);
  container.appendChild(deleteBtn);
  container.appendChild(detailsBtn);

  return container;
}

function refreshShipment(gridApi, id) {
  fetch(`/api/shipments/${id}/refresh`, { method: "POST" })
    .then((response) => response.json())
    .then((data) => {
      gridApi.applyTransaction({ update: [data.shipment] });
      showToast("Shipment refreshed", "info");
      console.log("Shipment Refreshed");
    })
    .catch((err) => {
      showToast("Error refreshing shipment", "error");
      console.error("Refresh error:", err);
    });
}

function deleteShipment(gridApi, id) {
  fetch(`/api/shipments/${id}`, { method: "DELETE" })
    .then((response) => {
      if (!response.ok) throw new Error("Failed to delete");
      return response.json();
    })
    .then(() => {
      gridApi.applyTransaction({ remove: [{ id }] });
      showToast("Shipment deleted");
    })
    .catch((err) => {
      showToast("Error deleting shipment", "error");
      console.error("Delete error:", err);
    });
}
