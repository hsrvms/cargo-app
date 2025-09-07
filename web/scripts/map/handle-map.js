import { getVisibleShipments } from "../ag-grid/get-visible-shipments.js";

// Create custom marker icons for different shipping statuses
function createStatusIcon(status) {
  const statusConfig = {
    IN_TRANSIT: { color: "#3B82F6", bgColor: "#DBEAFE" }, // Blue
    DELIVERED: { color: "#10B981", bgColor: "#D1FAE5" }, // Green
    PLANNED: { color: "#F59E0B", bgColor: "#FEF3C7" }, // Yellow/Amber
    UNKNOWN: { color: "#6B7280", bgColor: "#F3F4F6" }, // Gray
  };

  const config = statusConfig[status] || statusConfig.UNKNOWN;

  return L.divIcon({
    className: "custom-shipment-marker",
    html: `
      <div style="
        background-color: ${config.bgColor};
        border: 2px solid ${config.color};
        border-radius: 50%;
        width: 20px;
        height: 20px;
        display: flex;
        align-items: center;
        justify-content: center;
        box-shadow: 0 2px 4px rgba(0,0,0,0.2);
      ">
        <div style="
          background-color: ${config.color};
          border-radius: 50%;
          width: 8px;
          height: 8px;
        "></div>
      </div>
    `,
    iconSize: [20, 20],
    iconAnchor: [10, 10],
    popupAnchor: [0, -10],
  });
}

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
    minZoom: 2,
    maxZoom: 19,
    attribution:
      '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>',
  }).addTo(map);

  // Store markers for potential cleanup
  const markers = [];
  const validCoordinates = [];
  const shipmentsWithoutCoordinates = [];

  // --- Add Markers for Visible Shipments ---
  visibleShipments.forEach((shipment) => {
    // Check if shipment has current coordinates
    if (
      shipment.routeData &&
      shipment.routeData.coordinates &&
      shipment.routeData.coordinates.latitude &&
      shipment.routeData.coordinates.longitude
    ) {
      const coords = shipment.routeData.coordinates;
      const lat = coords.latitude;
      const lng = coords.longitude;

      // Validate coordinates
      if (lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180) {
        validCoordinates.push([lat, lng]);

        // Create marker with custom status-based icon
        const customIcon = createStatusIcon(shipment.shippingStatus);
        const marker = L.marker([lat, lng], { icon: customIcon }).addTo(map);

        // Create popup content with shipment information
        const popupContent = createShipmentPopup(shipment);
        marker.bindPopup(popupContent);

        markers.push(marker);

        console.log(
          `Added marker for ${shipment.shipmentNumber} at [${lat}, ${lng}]`,
        );
      } else {
        console.warn(
          `Invalid coordinates for ${shipment.shipmentNumber}: [${lat}, ${lng}]`,
        );
        shipmentsWithoutCoordinates.push({
          shipment: shipment,
          reason: "Invalid coordinates",
        });
      }
    } else {
      console.warn(
        `No coordinates available for shipment: ${shipment.shipmentNumber || shipment.id}`,
      );
      shipmentsWithoutCoordinates.push({
        shipment: shipment,
        reason: "No coordinates data",
      });
    }
  });

  // Auto fit map to markers if we have valid coordinates
  if (validCoordinates.length > 0) {
    if (validCoordinates.length === 1) {
      // Single marker - center and zoom to a reasonable level
      map.setView(validCoordinates[0], 8);
    } else {
      // Multiple markers - fit bounds to show all
      const bounds = L.latLngBounds(validCoordinates);
      map.fitBounds(bounds, { padding: [20, 20] });
    }
    console.log(`Map fitted to ${validCoordinates.length} shipment locations`);
  } else {
    console.log("No valid coordinates found for visible shipments");

    // Show notification if all shipments lack coordinates
    if (visibleShipments.length > 0) {
      showNoCoordinatesNotification(shipmentsWithoutCoordinates.length);
    }
  }

  // Log summary of coordinate availability
  if (shipmentsWithoutCoordinates.length > 0) {
    console.group("Shipments without coordinates:");
    shipmentsWithoutCoordinates.forEach((item) => {
      console.log(
        `${item.shipment.shipmentNumber || item.shipment.id}: ${item.reason}`,
      );
    });
    console.groupEnd();
  }

  // Add legend to map
  addMapLegend(map);

  // Add map controls
  addMapControls(map, gridApi);

  // Store markers on map object for potential cleanup
  map._shipmentMarkers = markers;

  return map;
}

// Add legend to the map
function addMapLegend(map) {
  const legend = L.control({ position: "topright" });

  legend.onAdd = function (map) {
    const div = L.DomUtil.create("div", "map-legend");

    div.innerHTML = `
      <div class="legend-title">Shipment Status</div>
      <div class="legend-item">
        <div class="legend-marker" style="background-color: #DBEAFE; border-color: #3B82F6;">
          <div style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); width: 6px; height: 6px; border-radius: 50%; background-color: #3B82F6;"></div>
        </div>
        <span>In Transit</span>
      </div>
      <div class="legend-item">
        <div class="legend-marker" style="background-color: #D1FAE5; border-color: #10B981;">
          <div style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); width: 6px; height: 6px; border-radius: 50%; background-color: #10B981;"></div>
        </div>
        <span>Delivered</span>
      </div>
      <div class="legend-item">
        <div class="legend-marker" style="background-color: #FEF3C7; border-color: #F59E0B;">
          <div style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); width: 6px; height: 6px; border-radius: 50%; background-color: #F59E0B;"></div>
        </div>
        <span>Planned</span>
      </div>
      <div class="legend-item">
        <div class="legend-marker" style="background-color: #F3F4F6; border-color: #6B7280;">
          <div style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); width: 6px; height: 6px; border-radius: 50%; background-color: #6B7280;"></div>
        </div>
        <span>Unknown</span>
      </div>
    `;

    return div;
  };

  legend.addTo(map);
}

// Add custom map controls
function addMapControls(map, gridApi) {
  const mapControls = L.control({ position: "topleft" });

  mapControls.onAdd = function (map) {
    const div = L.DomUtil.create("div", "map-controls");

    div.innerHTML = `
      <div style="background: rgba(255, 255, 255, 0.9); border-radius: 6px; padding: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.2); margin-bottom: 8px;">
        <button id="refreshMapBtn" class="map-control-btn" title="Refresh Map">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 2v6h-6M3 12a9 9 0 0 1 15-6.7L21 8M3 22v-6h6M21 12a9 9 0 0 1-15 6.7L3 16"/>
          </svg>
        </button>
        <button id="fitMapBtn" class="map-control-btn" title="Fit All Markers">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"/>
          </svg>
        </button>
        <button id="toggleMarkersBtn" class="map-control-btn" title="Toggle Markers">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/>
            <circle cx="12" cy="10" r="3"/>
          </svg>
        </button>
      </div>
    `;

    // Prevent map events when clicking controls
    L.DomEvent.disableClickPropagation(div);

    // Add event listeners
    div.querySelector("#refreshMapBtn").addEventListener("click", (e) => {
      e.preventDefault();
      console.log("Refreshing map...");
      updateMapMarkers(map, gridApi);
    });

    div.querySelector("#fitMapBtn").addEventListener("click", (e) => {
      e.preventDefault();
      fitMapToMarkers(map);
    });

    div.querySelector("#toggleMarkersBtn").addEventListener("click", (e) => {
      e.preventDefault();
      toggleMarkers(map);
    });

    return div;
  };

  mapControls.addTo(map);
}

// Function to update map markers when grid data changes (filters, pagination, etc.)
export function updateMapMarkers(map, gridApi) {
  if (!map || !gridApi) {
    console.warn("updateMapMarkers: map or gridApi is not available");
    return;
  }

  console.log("Updating map markers based on current grid state...");

  try {
    // Clear existing markers
    if (map._shipmentMarkers && map._shipmentMarkers.length > 0) {
      map._shipmentMarkers.forEach((marker) => {
        try {
          map.removeLayer(marker);
        } catch (error) {
          console.warn("Error removing marker:", error);
        }
      });
      map._shipmentMarkers = [];
    }

    // Get current visible shipments
    const visibleShipments = getVisibleShipments(gridApi);
    console.log(
      `Updating map with ${visibleShipments.length} visible shipments`,
    );

    const markers = [];
    const validCoordinates = [];
    const shipmentsWithoutCoordinates = [];

    // Add markers for current visible shipments
    visibleShipments.forEach((shipment) => {
      if (
        shipment.routeData &&
        shipment.routeData.coordinates &&
        shipment.routeData.coordinates.latitude &&
        shipment.routeData.coordinates.longitude
      ) {
        const coords = shipment.routeData.coordinates;
        const lat = coords.latitude;
        const lng = coords.longitude;

        if (lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180) {
          validCoordinates.push([lat, lng]);

          try {
            const customIcon = createStatusIcon(shipment.shippingStatus);
            const marker = L.marker([lat, lng], { icon: customIcon }).addTo(
              map,
            );

            const popupContent = createShipmentPopup(shipment);
            marker.bindPopup(popupContent);

            markers.push(marker);
          } catch (error) {
            console.error(
              `Error creating marker for ${shipment.shipmentNumber}:`,
              error,
            );
          }
        } else {
          shipmentsWithoutCoordinates.push({
            shipment: shipment,
            reason: "Invalid coordinates",
          });
        }
      } else {
        shipmentsWithoutCoordinates.push({
          shipment: shipment,
          reason: "No coordinates data",
        });
      }
    });

    // Update map view if we have markers
    if (validCoordinates.length > 0) {
      try {
        if (validCoordinates.length === 1) {
          map.setView(validCoordinates[0], 8);
        } else {
          const bounds = L.latLngBounds(validCoordinates);
          map.fitBounds(bounds, { padding: [20, 20] });
        }
      } catch (error) {
        console.error("Error updating map view:", error);
      }
    }

    // Store updated markers
    map._shipmentMarkers = markers;
    console.log(`Map updated with ${markers.length} markers`);

    // Log shipments without coordinates
    if (shipmentsWithoutCoordinates.length > 0) {
      console.warn(
        `${shipmentsWithoutCoordinates.length} shipments don't have valid coordinates`,
      );
    }
  } catch (error) {
    console.error("Error in updateMapMarkers:", error);
  }
}

// Fit map to show all current markers
function fitMapToMarkers(map) {
  if (!map._shipmentMarkers || map._shipmentMarkers.length === 0) {
    console.log("No markers to fit");
    return;
  }

  const coordinates = map._shipmentMarkers.map((marker) => marker.getLatLng());

  if (coordinates.length === 1) {
    map.setView([coordinates[0].lat, coordinates[0].lng], 8);
  } else {
    const bounds = L.latLngBounds(coordinates);
    map.fitBounds(bounds, { padding: [20, 20] });
  }

  console.log(`Map fitted to ${coordinates.length} markers`);
}

// Toggle marker visibility
function toggleMarkers(map) {
  if (!map._shipmentMarkers) return;

  const markersVisible = map._shipmentMarkersVisible !== false;

  map._shipmentMarkers.forEach((marker) => {
    if (markersVisible) {
      map.removeLayer(marker);
    } else {
      map.addLayer(marker);
    }
  });

  map._shipmentMarkersVisible = !markersVisible;
  console.log(`Markers ${markersVisible ? "hidden" : "shown"}`);
}

// Helper function to create popup content for shipments
function createShipmentPopup(shipment) {
  try {
    const formatDate = (dateStr) => {
      if (!dateStr) return "N/A";
      try {
        const date = new Date(dateStr);
        if (isNaN(date.getTime())) return "N/A";
        return date.toLocaleDateString("en-US", {
          month: "short",
          day: "numeric",
          year: "numeric",
        });
      } catch (error) {
        console.warn("Error formatting date:", dateStr, error);
        return "N/A";
      }
    };

    const getRouteInfo = () => {
      if (!shipment.route) return { origin: "N/A", destination: "N/A" };

      const origin =
        shipment.route.pol?.location?.name ||
        shipment.route.prepol?.location?.name ||
        "N/A";
      const destination =
        shipment.route.pod?.location?.name ||
        shipment.route.postpod?.location?.name ||
        "N/A";

      return { origin, destination };
    };

    const getVesselInfo = () => {
      try {
        if (!shipment.vessels || shipment.vessels.length === 0) return "N/A";
        const vessel = shipment.vessels[0];
        return vessel?.name || "Unknown Vessel";
      } catch (error) {
        console.warn("Error getting vessel info:", error);
        return "N/A";
      }
    };

    const getStatusBadge = (status) => {
      const statusConfig = {
        IN_TRANSIT: { label: "In Transit", color: "blue" },
        DELIVERED: { label: "Delivered", color: "green" },
        PLANNED: { label: "Planned", color: "yellow" },
        UNKNOWN: { label: "Unknown", color: "gray" },
      };

      const config = statusConfig[status] || statusConfig.UNKNOWN;
      return `<span style="background: ${config.color}; color: white; padding: 2px 6px; border-radius: 4px; font-size: 11px;">${config.label}</span>`;
    };

    const routeInfo = getRouteInfo();
    const lastUpdate = formatDate(
      shipment.routeData?.coordinates?.updatedAt || shipment.updatedAt,
    );

    return `
    <div style="min-width: 200px;">
      <div style="font-weight: bold; margin-bottom: 8px; color: #2563eb;">
        ${shipment.shipmentNumber || "N/A"}
      </div>

      <div style="margin-bottom: 4px;">
        <strong>Status:</strong> ${getStatusBadge(shipment.shippingStatus)}
      </div>

      <div style="margin-bottom: 4px;">
        <strong>Route:</strong><br>
        <small>${routeInfo.origin} → ${routeInfo.destination}</small>
      </div>

      <div style="margin-bottom: 4px;">
        <strong>Vessel:</strong> ${getVesselInfo()}
      </div>

      <div style="margin-bottom: 4px;">
        <strong>Sealine:</strong> ${shipment.sealineName || shipment.sealineCode || "N/A"}
      </div>

      <div style="margin-bottom: 8px;">
        <strong>Containers:</strong> ${shipment.containers?.length || 0}
      </div>

      <div style="font-size: 11px; color: #666; border-top: 1px solid #eee; padding-top: 4px;">
        Last updated: ${lastUpdate}
      </div>

      <div style="margin-top: 8px;">
        <button onclick="openModalFetchDetails('${shipment.id}')"
                style="background: #2563eb; color: white; border: none; padding: 4px 8px; border-radius: 4px; cursor: pointer; font-size: 11px;">
          View Details
        </button>
      </div>
    </div>
  `;
  } catch (error) {
    console.error("Error creating popup content:", error);
    return `
      <div style="color: red; padding: 8px;">
        <strong>Error loading shipment details</strong><br>
        ID: ${shipment?.id || "Unknown"}
      </div>
    `;
  }
}

// Function to show notification for shipments without coordinates
function showNoCoordinatesNotification(count) {
  // Remove existing notification
  const existingNotification = document.getElementById(
    "no-coordinates-notification",
  );
  if (existingNotification) {
    existingNotification.remove();
  }

  // Create notification
  const notification = document.createElement("div");
  notification.id = "no-coordinates-notification";
  notification.innerHTML = `
    <div style="
      position: fixed;
      top: 20px;
      right: 20px;
      background: #FEF3C7;
      border: 1px solid #F59E0B;
      color: #92400E;
      padding: 12px;
      border-radius: 8px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.1);
      z-index: 1000;
      max-width: 300px;
      font-size: 14px;
    ">
      <strong>⚠️ Location Data Missing</strong><br>
      ${count} shipment${count > 1 ? "s" : ""} ${count > 1 ? "don't" : "doesn't"} have location coordinates and won't appear on the map.
      <button onclick="this.parentElement.parentElement.remove()" style="
        float: right;
        background: none;
        border: none;
        font-size: 16px;
        cursor: pointer;
        margin-left: 8px;
      ">×</button>
    </div>
  `;

  document.body.appendChild(notification);

  // Auto-hide after 8 seconds
  setTimeout(() => {
    if (document.getElementById("no-coordinates-notification")) {
      notification.remove();
    }
  }, 8000);
}
