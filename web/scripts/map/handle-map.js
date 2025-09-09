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

  // Store markers and routes for potential cleanup
  const markers = [];
  const routes = [];
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

  // --- Draw Routes for Visible Shipments ---
  visibleShipments.forEach((shipment) => {
    if (
      shipment.routeData &&
      shipment.routeData.routeSegments &&
      shipment.routeData.routeSegments.length > 0
    ) {
      const shipmentRoutes = drawShipmentRoutes(map, shipment);
      routes.push(...shipmentRoutes);
    }
  });

  // Store markers and routes on map object for potential cleanup
  map._shipmentMarkers = markers;
  map._shipmentRoutes = routes;

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

      <div class="legend-title" style="margin-top: 16px; padding-top: 12px; border-top: 1px solid #e5e7eb;">Route Types</div>
      <div class="legend-item">
        <div style="width: 20px; height: 2px; background: repeating-linear-gradient(to right, #06B6D4 0px, #06B6D4 3px, transparent 3px, transparent 9px); margin-right: 6px; border-radius: 1px;"></div>
        <span style="font-size: 11px;">Sea</span>
      </div>
      <div class="legend-item">
        <div style="width: 20px; height: 2px; background: repeating-linear-gradient(to right, #10B981 0px, #10B981 2px, transparent 2px, transparent 6px); margin-right: 6px; border-radius: 1px;"></div>
        <span style="font-size: 11px;">Land</span>
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
        <button id="toggleMarkersBtn" class="map-control-btn" title="Toggle Markers & Routes">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/>
            <circle cx="12" cy="10" r="3"/>
          </svg>
        </button>
        <button id="toggleRoutesBtn" class="map-control-btn" title="Toggle Routes Only">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="4,18 8.5,11.5 13.5,15.5 22,4"/>
          </svg>
        </button>
        <button id="optimizeRoutesBtn" class="map-control-btn" title="Optimize Route Display">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2L2 7L12 12L22 7L12 2"/>
            <path d="M2 17L12 22L22 17"/>
            <path d="M2 12L12 17L22 12"/>
          </svg>
        </button>
        <button id="showRouteStatsBtn" class="map-control-btn" title="Show Route Statistics">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M16 4v12l-4-2-4 2V4"/>
            <rect x="6" y="2" width="12" height="18" rx="2" ry="2"/>
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

    div.querySelector("#toggleRoutesBtn").addEventListener("click", (e) => {
      e.preventDefault();
      toggleRoutes(map);
    });

    div.querySelector("#optimizeRoutesBtn").addEventListener("click", (e) => {
      e.preventDefault();
      optimizeRouteDisplay(map);
    });

    div.querySelector("#showRouteStatsBtn").addEventListener("click", (e) => {
      e.preventDefault();
      showRouteStatistics(map, gridApi);
    });

    return div;
  };

  mapControls.addTo(map);
}

// Function to draw routes for a single shipment
function drawShipmentRoutes(map, shipment) {
  const routes = [];

  if (!shipment.routeData || !shipment.routeData.routeSegments) {
    return routes;
  }

  // Define colors for different route types
  const routeColors = {
    SEA: "#06B6D4", // Cyan
    LAND: "#10B981", // Green
    UNKNOWN: "#6B7280", // Gray
  };

  // Create route group for this shipment
  const routeGroup = L.layerGroup().addTo(map);

  // Store route metadata
  if (!map._routeGroups) {
    map._routeGroups = new Map();
  }
  map._routeGroups.set(shipment.id, routeGroup);

  // Sort segments by order
  const sortedSegments = [...shipment.routeData.routeSegments].sort(
    (a, b) => (a.segmentOrder || 0) - (b.segmentOrder || 0),
  );

  sortedSegments.forEach((segment, index) => {
    if (!segment.path || segment.path.length < 2) {
      return; // Skip segments without enough points
    }

    // Sort path points by order
    const sortedPath = [...segment.path].sort(
      (a, b) => (a.pointOrder || 0) - (b.pointOrder || 0),
    );

    // Convert path to LatLng array
    const latlngs = sortedPath.map((point) => [
      point.latitude,
      point.longitude,
    ]);

    // Validate coordinates
    const validLatLngs = latlngs.filter(
      ([lat, lng]) => lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180,
    );

    if (validLatLngs.length < 2) {
      console.warn(
        `Insufficient valid coordinates for route segment ${index} of shipment ${shipment.shipmentNumber}`,
      );
      return;
    }

    // Get route color based on type
    const routeType = segment.routeType || "UNKNOWN";
    const color = routeColors[routeType] || routeColors["UNKNOWN"];

    // Check for overlapping routes and adjust weight - using thinner dotted lines
    const baseWeight = routeType === "SEA" ? 2 : 2;
    const routeWeight = getOptimalRouteWeight(map, validLatLngs, baseWeight);

    // Enhanced visual styling for SEA vs LAND routes - both use dotted thin lines
    const routeStyle = {
      color: color,
      weight: routeWeight,
      opacity: routeType === "SEA" ? 0.7 : 0.6,
      interactive: true,
      className: `route-${shipment.id} route-type-${routeType.toLowerCase()}`,
      lineCap: "round",
      lineJoin: "round",
    };

    // Different dotted patterns for SEA vs LAND
    if (routeType === "SEA") {
      routeStyle.dashArray = "3, 6"; // Larger dots with more spacing for sea routes
    } else if (routeType === "LAND") {
      routeStyle.dashArray = "2, 4"; // Smaller dots with less spacing for land routes
    }

    // Create polyline with enhanced styling
    const polyline = L.polyline(validLatLngs, routeStyle).addTo(routeGroup);

    // Add hover effects optimized for thin dotted lines
    polyline.on("mouseover", function (e) {
      this.setStyle({
        weight: Math.min(routeWeight + 1, 4), // More subtle weight increase for thin lines
        opacity: 0.9, // Slightly less intense opacity change
        color: routeType === "SEA" ? "#0891b2" : "#059669", // Slightly darker color on hover
      });
      this.bringToFront();
    });

    polyline.on("mouseout", function (e) {
      this.setStyle({
        weight: routeWeight,
        opacity: routeType === "SEA" ? 0.7 : 0.6,
        color: color, // Reset to original color
      });
    });

    // Add popup to route segment with enhanced information
    const popupContent = createRoutePopupContent(
      shipment,
      routeType,
      index,
      sortedSegments,
      validLatLngs,
      color,
    );
    polyline.bindPopup(popupContent);

    // Store route metadata
    polyline._shipmentId = shipment.id;
    polyline._routeType = routeType;
    polyline._segmentIndex = index;

    routes.push(polyline);
  });

  return routes;
}

// Function to get optimal route weight based on overlapping routes
function getOptimalRouteWeight(map, routeLatLngs, baseWeight) {
  if (!map._shipmentRoutes || map._shipmentRoutes.length === 0) {
    return baseWeight;
  }

  let overlappingCount = 0;

  // Simple overlap detection - check if routes share similar coordinates
  map._shipmentRoutes.forEach((existingRoute) => {
    if (existingRoute.getLatLngs) {
      const existingLatLngs = existingRoute.getLatLngs();
      const hasOverlap = routeLatLngs.some((newPoint) =>
        existingLatLngs.some(
          (existingPoint) =>
            Math.abs(newPoint[0] - existingPoint.lat) < 0.1 &&
            Math.abs(newPoint[1] - existingPoint.lng) < 0.1,
        ),
      );

      if (hasOverlap) {
        overlappingCount++;
      }
    }
  });

  // Adjust weight based on overlap - reduce weight for overlapping routes
  return Math.max(baseWeight - Math.floor(overlappingCount / 2), 2);
}

// Function to create enhanced route popup content
function createRoutePopupContent(
  shipment,
  routeType,
  segmentIndex,
  allSegments,
  validLatLngs,
  color,
) {
  const distanceKm = calculateRouteDistance(validLatLngs);

  return `
    <div style="font-size: 12px; min-width: 200px;">
      <div style="font-weight: bold; color: #2563eb; margin-bottom: 8px;">
        ${shipment.shipmentNumber}
      </div>

      <div style="margin: 4px 0;">
        <strong>Route Type:</strong> <span style="color: ${color};">${routeType}</span><br>
        <strong>Segment:</strong> ${segmentIndex + 1} of ${allSegments.length}<br>
        <strong>Distance:</strong> ~${distanceKm} km<br>
        <strong>Points:</strong> ${validLatLngs.length}
      </div>

      <div style="margin: 8px 0; padding: 6px; background: #f8fafc; border-radius: 4px;">
        <div style="font-size: 11px; color: #64748b;">
          <strong>Vessel:</strong> ${shipment.vessels?.[0]?.name || "N/A"}<br>
          <strong>Status:</strong> ${shipment.shippingStatus || "Unknown"}
        </div>
      </div>

      <div style="margin-top: 8px; display: flex; align-items: center;">
        <span style="display: inline-block; width: 20px; height: 2px; background: ${routeType === "SEA" ? `repeating-linear-gradient(to right, ${color} 0px, ${color} 3px, transparent 3px, transparent 9px)` : `repeating-linear-gradient(to right, ${color} 0px, ${color} 2px, transparent 2px, transparent 6px)`}; margin-right: 8px; border-radius: 1px;"></span>
        <span style="font-size: 11px; color: #64748b;">${routeType} Route (dotted)</span>
      </div>

      <button onclick="highlightShipmentRoute('${shipment.id}')"
              style="margin-top: 8px; background: #3b82f6; color: white; border: none; padding: 4px 8px; border-radius: 4px; cursor: pointer; font-size: 11px;">
        Highlight All Routes
      </button>
    </div>
  `;
}

// Function to calculate approximate route distance
function calculateRouteDistance(latLngs) {
  if (latLngs.length < 2) return 0;

  let totalDistance = 0;
  for (let i = 1; i < latLngs.length; i++) {
    const lat1 = latLngs[i - 1][0];
    const lon1 = latLngs[i - 1][1];
    const lat2 = latLngs[i][0];
    const lon2 = latLngs[i][1];

    // Haversine formula for approximate distance
    const R = 6371; // Earth's radius in km
    const dLat = ((lat2 - lat1) * Math.PI) / 180;
    const dLon = ((lon2 - lon1) * Math.PI) / 180;
    const a =
      Math.sin(dLat / 2) * Math.sin(dLat / 2) +
      Math.cos((lat1 * Math.PI) / 180) *
        Math.cos((lat2 * Math.PI) / 180) *
        Math.sin(dLon / 2) *
        Math.sin(dLon / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    totalDistance += R * c;
  }

  return Math.round(totalDistance);
}

// Global function to highlight specific shipment routes
window.highlightShipmentRoute = function (shipmentId) {
  const map = window.currentMap;
  if (!map || !map._routeGroups) return;

  // Reset all routes to normal style
  if (map._shipmentRoutes) {
    map._shipmentRoutes.forEach((route) => {
      if (route.setStyle) {
        route.setStyle({
          opacity: 0.7,
          weight: route.options.weight || 3,
        });
      }
    });
  }

  // Highlight routes for specific shipment
  const routeGroup = map._routeGroups.get(shipmentId);
  if (routeGroup) {
    routeGroup.eachLayer((layer) => {
      if (layer.setStyle) {
        layer.setStyle({
          opacity: 1.0,
          weight: (layer.options.weight || 3) + 3,
        });
        layer.bringToFront();
      }
    });

    setTimeout(() => {
      // Reset after 3 seconds
      routeGroup.eachLayer((layer) => {
        if (layer.setStyle) {
          layer.setStyle({
            opacity: 0.7,
            weight: layer.options.weight || 3,
          });
        }
      });
    }, 3000);
  }
};

// Function to update map markers and routes when grid data changes (filters, pagination, etc.)
export function updateMapMarkers(map, gridApi) {
  if (!map || !gridApi) {
    console.warn("updateMapMarkers: map or gridApi is not available");
    return;
  }

  console.log("Updating map markers and routes based on current grid state...");

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

    // Clear existing routes and route groups
    if (map._shipmentRoutes && map._shipmentRoutes.length > 0) {
      map._shipmentRoutes.forEach((route) => {
        try {
          map.removeLayer(route);
        } catch (error) {
          console.warn("Error removing route:", error);
        }
      });
      map._shipmentRoutes = [];
    }

    // Clear route groups
    if (map._routeGroups) {
      map._routeGroups.forEach((routeGroup, shipmentId) => {
        try {
          map.removeLayer(routeGroup);
        } catch (error) {
          console.warn("Error removing route group:", error);
        }
      });
      map._routeGroups.clear();
    }

    // Get current visible shipments
    const visibleShipments = getVisibleShipments(gridApi);

    const markers = [];
    const routes = [];
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

    // Add routes for current visible shipments
    visibleShipments.forEach((shipment) => {
      if (
        shipment.routeData &&
        shipment.routeData.routeSegments &&
        shipment.routeData.routeSegments.length > 0
      ) {
        try {
          const shipmentRoutes = drawShipmentRoutes(map, shipment);
          routes.push(...shipmentRoutes);
        } catch (error) {
          console.error(
            `Error drawing routes for ${shipment.shipmentNumber}:`,
            error,
          );
        }
      }
    });

    // Store updated markers and routes
    map._shipmentMarkers = markers;
    map._shipmentRoutes = routes;
    console.log(
      `Map updated with ${markers.length} markers and ${routes.length} route segments`,
    );

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

// Toggle marker and route visibility
function toggleMarkers(map) {
  if (!map._shipmentMarkers && !map._shipmentRoutes) return;

  const markersVisible = map._shipmentMarkersVisible !== false;

  // Toggle markers
  if (map._shipmentMarkers) {
    map._shipmentMarkers.forEach((marker) => {
      if (markersVisible) {
        map.removeLayer(marker);
      } else {
        map.addLayer(marker);
      }
    });
  }

  // Toggle routes
  if (map._shipmentRoutes) {
    map._shipmentRoutes.forEach((route) => {
      if (markersVisible) {
        map.removeLayer(route);
      } else {
        map.addLayer(route);
      }
    });
  }

  map._shipmentMarkersVisible = !markersVisible;
  console.log(`Markers and routes ${markersVisible ? "hidden" : "shown"}`);
}

// Toggle route visibility only
function toggleRoutes(map) {
  if (!map._shipmentRoutes) return;

  const routesVisible = map._shipmentRoutesVisible !== false;

  map._shipmentRoutes.forEach((route) => {
    if (routesVisible) {
      map.removeLayer(route);
    } else {
      map.addLayer(route);
    }
  });

  map._shipmentRoutesVisible = !routesVisible;
  console.log(`Routes ${routesVisible ? "hidden" : "shown"}`);
}

// Function to optimize route display by reducing overlapping routes
function optimizeRouteDisplay(map) {
  if (!map._shipmentRoutes || map._shipmentRoutes.length === 0) {
    console.log("No routes to optimize");
    return;
  }

  console.log("Optimizing route display...");

  // Group routes by type and apply different opacities
  const routesByType = {
    SEA: [],
    LAND: [],
    UNKNOWN: [],
  };

  map._shipmentRoutes.forEach((route) => {
    const routeType = route._routeType || "UNKNOWN";
    if (routesByType[routeType]) {
      routesByType[routeType].push(route);
    }
  });

  // Apply different styles based on route type and count
  Object.keys(routesByType).forEach((routeType) => {
    const routes = routesByType[routeType];
    const routeCount = routes.length;

    if (routeCount > 0) {
      const opacity =
        routeType === "SEA" ? 0.8 : Math.max(0.3, 1.0 - routeCount * 0.1);
      const weight =
        routeType === "SEA" ? 2 : Math.max(1, 3 - Math.floor(routeCount / 2));

      routes.forEach((route, index) => {
        if (route.setStyle) {
          route.setStyle({
            opacity: opacity,
            weight: weight,
            zIndex: routeType === "SEA" ? 1000 + index : 500 + index,
          });
        }
      });
    }
  });

  console.log(`Optimized display for ${map._shipmentRoutes.length} routes`);

  // Show optimization notification
  showOptimizationNotification(
    Object.keys(routesByType).filter((type) => routesByType[type].length > 0),
  );
}

// Function to show route optimization notification
function showOptimizationNotification(optimizedTypes) {
  const existingNotification = document.getElementById(
    "route-optimization-notification",
  );
  if (existingNotification) {
    existingNotification.remove();
  }

  const notification = document.createElement("div");
  notification.id = "route-optimization-notification";
  notification.innerHTML = `
    <div style="
      position: fixed;
      top: 20px;
      left: 20px;
      background: #DBEAFE;
      border: 1px solid #3B82F6;
      color: #1E40AF;
      padding: 12px;
      border-radius: 8px;
      box-shadow: 0 4px 12px rgba(0,0,0,0.1);
      z-index: 1000;
      max-width: 300px;
      font-size: 14px;
    ">
      <strong>🛣️ Routes Optimized</strong><br>
      Adjusted display for ${optimizedTypes.join(", ")} routes to reduce overlap.
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

  setTimeout(() => {
    if (document.getElementById("route-optimization-notification")) {
      notification.remove();
    }
  }, 5000);
}

// Function to show route statistics
function showRouteStatistics(map, gridApi) {
  if (!map._shipmentRoutes || map._shipmentRoutes.length === 0) {
    showNoRoutesNotification();
    return;
  }

  const visibleShipments = getVisibleShipments(gridApi);

  // Calculate route statistics
  const stats = {
    totalShipments: visibleShipments.length,
    totalRouteSegments: map._shipmentRoutes.length,
    seaRoutes: 0,
    landRoutes: 0,
    totalDistance: 0,
    shipmentsWithRoutes: 0,
  };

  // Count shipments with route data
  visibleShipments.forEach((shipment) => {
    if (
      shipment.routeData &&
      shipment.routeData.routeSegments &&
      shipment.routeData.routeSegments.length > 0
    ) {
      stats.shipmentsWithRoutes++;
    }
  });

  // Analyze route segments
  map._shipmentRoutes.forEach((route) => {
    if (route._routeType === "SEA") {
      stats.seaRoutes++;
    } else if (route._routeType === "LAND") {
      stats.landRoutes++;
    }

    // Calculate approximate distance if available
    if (route.getLatLngs) {
      const latLngs = route.getLatLngs();
      const distance = calculateRouteDistance(
        latLngs.map((ll) => [ll.lat, ll.lng]),
      );
      stats.totalDistance += distance;
    }
  });

  // Show statistics popup
  const existingStats = document.getElementById("route-statistics-popup");
  if (existingStats) {
    existingStats.remove();
  }

  const statsPopup = document.createElement("div");
  statsPopup.id = "route-statistics-popup";
  statsPopup.innerHTML = `
    <div style="
      position: fixed;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      background: white;
      border-radius: 12px;
      box-shadow: 0 8px 24px rgba(0,0,0,0.2);
      z-index: 1001;
      padding: 20px;
      min-width: 300px;
      max-width: 400px;
    ">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <h3 style="margin: 0; color: #1f2937; font-size: 18px;">📊 Route Statistics</h3>
        <button onclick="this.closest('#route-statistics-popup').remove()" style="
          background: none;
          border: none;
          font-size: 20px;
          cursor: pointer;
          color: #6b7280;
        ">×</button>
      </div>

      <div style="space-y: 12px;">
        <div style="display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #e5e7eb;">
          <span style="color: #6b7280;">Total Shipments:</span>
          <span style="font-weight: 600; color: #1f2937;">${stats.totalShipments}</span>
        </div>

        <div style="display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #e5e7eb;">
          <span style="color: #6b7280;">With Route Data:</span>
          <span style="font-weight: 600; color: #059669;">${stats.shipmentsWithRoutes}</span>
        </div>

        <div style="display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #e5e7eb;">
          <span style="color: #6b7280;">Total Route Segments:</span>
          <span style="font-weight: 600; color: #1f2937;">${stats.totalRouteSegments}</span>
        </div>

        <div style="display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #e5e7eb;">
          <span style="color: #6b7280; display: flex; align-items: center;">
            <span style="display: inline-block; width: 16px; height: 2px; background: repeating-linear-gradient(to right, #06B6D4 0px, #06B6D4 3px, transparent 3px, transparent 9px); margin-right: 6px; border-radius: 1px;"></span>
            Sea Routes:
          </span>
          <span style="font-weight: 600; color: #0891b2;">${stats.seaRoutes}</span>
        </div>

        <div style="display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #e5e7eb;">
          <span style="color: #6b7280; display: flex; align-items: center;">
            <span style="display: inline-block; width: 16px; height: 2px; background: repeating-linear-gradient(to right, #10B981 0px, #10B981 2px, transparent 2px, transparent 6px); margin-right: 6px; border-radius: 1px;"></span>
            Land Routes:
          </span>
          <span style="font-weight: 600; color: #059669;">${stats.landRoutes}</span>
        </div>

        <div style="display: flex; justify-content: space-between; padding: 8px 0;">
          <span style="color: #6b7280;">Total Distance:</span>
          <span style="font-weight: 600; color: #1f2937;">~${stats.totalDistance.toLocaleString()} km</span>
        </div>
      </div>

      <div style="margin-top: 16px; padding-top: 16px; border-top: 1px solid #e5e7eb;">
        <div style="font-size: 12px; color: #6b7280; text-align: center;">
          Statistics based on currently visible shipments in the grid
        </div>
      </div>
    </div>
  `;

  // Add backdrop
  const backdrop = document.createElement("div");
  backdrop.style.cssText = `
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0,0,0,0.3);
    z-index: 1000;
  `;
  backdrop.onclick = () => {
    statsPopup.remove();
    backdrop.remove();
  };

  document.body.appendChild(backdrop);
  document.body.appendChild(statsPopup);

  console.log("Route Statistics:", stats);
}

// Function to show notification when no routes are available
function showNoRoutesNotification() {
  const notification = document.createElement("div");
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
      <strong>📊 No Route Data</strong><br>
      No route segments available for visible shipments.
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

  setTimeout(() => {
    if (notification.parentElement) {
      notification.remove();
    }
  }, 4000);
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
