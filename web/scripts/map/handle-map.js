import { getVisibleShipments } from "../ag-grid/get-visible-shipments.js";

export function handleMap(gridApi) {
  // Check if map container already has a map instance and remove it
  const mapContainer = document.getElementById("miniMap");
  if (mapContainer._leaflet_id) {
    mapContainer._leaflet_id = null;
    mapContainer.innerHTML = "";
  }

  var map = L.map("miniMap").setView([20, 0], 2);
  const visibleShipments = getVisibleShipments(gridApi);
  console.log("Visible shipments in map:", visibleShipments);

  L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
    maxZoom: 19,
    attribution:
      '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>',
  }).addTo(map);

  // // --- Add Markers for Locations ---
  // locations.forEach((loc) => {
  //   const marker = L.marker([loc.latitude, loc.longitude]).addTo(map);
  //   marker.bindPopup(`<b>${loc.name}</b><br/>${loc.country}`);
  // });

  // // --- Draw Routes as Polylines ---
  // routes.forEach((route) => {
  //   const latlngs = route.path.map((p) => [p.latitude, p.longitude]);
  //   L.polyline(latlngs, { color: "blue" }).addTo(map);
  // });

  // // Auto fit map to markers
  // if (locations.length > 0) {
  //   const bounds = L.latLngBounds(
  //     locations.map((loc) => [loc.latitude, loc.longitude]),
  //   );
  //   map.fitBounds(bounds);
  // }

  return map;
}
