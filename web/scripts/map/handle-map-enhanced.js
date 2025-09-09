/**
 * Enhanced Map Handler with Cross-Page Communication
 * Integrates with MapDataService for real-time synchronization between grid and map pages
 */

import { mapDataService } from "./map-data-service.js";
import { getVisibleShipments } from "../ag-grid/get-visible-shipments.js";

// Global variables for map management
let currentMapInstance = null;
let markers = [];
let routes = [];
let isStandaloneMap = false;

/**
 * Enhanced handleMap function that works with BroadcastChannel
 * @param {Object} gridApi - AG Grid API instance (optional for standalone map)
 * @param {Object} options - Configuration options
 */
export function handleMapEnhanced(gridApi = null, options = {}) {
  const {
    mapContainerId = "miniMap",
    standalone = false,
    initialView = [20, 0],
    initialZoom = 2,
  } = options;

  isStandaloneMap = standalone;

  // Initialize map container
  const mapContainer = document.getElementById(mapContainerId);
  if (!mapContainer) {
    console.error(`Map container '${mapContainerId}' not found`);
    return null;
  }

  // Clean up existing map instance
  if (mapContainer._leaflet_id) {
    mapContainer._leaflet_id = null;
    mapContainer.innerHTML = "";
  }

  // Create new map instance
  currentMapInstance = L.map(mapContainerId).setView(initialView, initialZoom);

  // Add tile layer with proper zoom limits
  L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
    minZoom: 2,
    maxZoom: 19,
    attribution:
      '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>',
  }).addTo(currentMapInstance);

  // Set map bounds to prevent over-zooming out
  currentMapInstance.setMaxBounds([
    [-85, -180],
    [85, 180],
  ]);

  // Initialize markers and routes arrays
  markers = [];
  routes = [];

  // Add map controls
  addMapControls(currentMapInstance, gridApi);

  // Set up cross-page communication
  setupMapDataServiceListeners();

  // If we have a gridApi, get initial data and set up grid listeners
  if (gridApi) {
    const visibleShipments = getVisibleShipments(gridApi, { debug: false });
    updateMapWithShipments(visibleShipments);

    // Broadcast initial data to other pages
    mapDataService.broadcastShipments(visibleShipments);

    // Set up grid event listeners for real-time updates
    setupGridEventListeners(gridApi);
  } else if (isStandaloneMap) {
    console.log(
      "🗺️ Initializing standalone map, requesting data from other pages",
    );
    // Request initial data from other pages
    mapDataService.requestInitialData();
  }

  // Set up map event listeners for broadcasting view changes
  setupMapEventListeners();

  // Store map globally for access from other components
  window.currentMapEnhanced = currentMapInstance;
  currentMapInstance._shipmentMarkers = markers;
  currentMapInstance._shipmentRoutes = routes;

  return currentMapInstance;
}

/**
 * Set up listeners for MapDataService events
 */
function setupMapDataServiceListeners() {
  // Listen for shipment updates from other pages
  mapDataService.addEventListener("shipmentsUpdate", (data) => {
    updateMapWithShipments(data.visibleShipments, data.selectedShipments);
  });

  // Listen for initial data responses
  mapDataService.addEventListener("initialDataReceived", (data) => {
    if (data.visibleShipments && data.visibleShipments.length > 0) {
      updateMapWithShipments(data.visibleShipments, data.selectedShipments);
    }
  });

  // Listen for selection updates
  mapDataService.addEventListener("selectionUpdate", (selectedShipments) => {
    console.log("📨 Received selection update:", selectedShipments);
    highlightSelectedShipments(selectedShipments);
  });
}

/**
 * Set up grid event listeners for real-time map updates
 */
function setupGridEventListeners(gridApi) {
  if (!gridApi) return;

  const updateMapFromGrid = () => {
    setTimeout(() => {
      const visibleShipments = getVisibleShipments(gridApi, { debug: false });
      const selectedShipments = getSelectedShipments(gridApi);

      // Update local map
      updateMapWithShipments(visibleShipments, selectedShipments);

      // Broadcast to other pages
      mapDataService.broadcastShipments(visibleShipments, selectedShipments);

      // Also broadcast grid state (with null checks)
      const gridState = {};
      try {
        if (typeof gridApi.getFilterModel === "function") {
          gridState.filters = gridApi.getFilterModel();
        }
        if (typeof gridApi.getSortModel === "function") {
          gridState.sorting = gridApi.getSortModel();
        }
        if (
          typeof gridApi.paginationGetCurrentPage === "function" &&
          typeof gridApi.paginationGetPageSize === "function"
        ) {
          gridState.pagination = {
            currentPage: gridApi.paginationGetCurrentPage() + 1,
            pageSize: gridApi.paginationGetPageSize(),
          };
        }
        mapDataService.broadcastGridState(gridState);
      } catch (error) {
        console.warn("Error getting grid state:", error);
      }
    }, 100);
  };

  // Set up grid event listeners
  const gridEvents = [
    "filterChanged",
    "paginationChanged",
    "sortChanged",
    "selectionChanged",
  ];

  gridEvents.forEach((eventName) => {
    gridApi.addEventListener(eventName, updateMapFromGrid);
  });
}

/**
 * Set up map event listeners for broadcasting view changes
 */
function setupMapEventListeners() {
  if (!currentMapInstance) return;

  // Broadcast map view changes to other pages
  const broadcastViewChange = () => {
    const center = currentMapInstance.getCenter();
    const zoom = currentMapInstance.getZoom();
    const bounds = currentMapInstance.getBounds();

    mapDataService.broadcastMapView({
      center: { lat: center.lat, lng: center.lng },
      zoom: zoom,
      bounds: {
        north: bounds.getNorth(),
        south: bounds.getSouth(),
        east: bounds.getEast(),
        west: bounds.getWest(),
      },
    });
  };

  currentMapInstance.on("moveend", broadcastViewChange);
  currentMapInstance.on("zoomend", broadcastViewChange);
}

/**
 * Update map with shipment data
 */
function updateMapWithShipments(visibleShipments, selectedShipments = []) {
  if (!currentMapInstance) return;

  // Clear existing markers and routes
  clearMapElements();

  const validCoordinates = [];
  const shipmentsWithoutCoordinates = [];

  // Add markers for visible shipments
  visibleShipments.forEach((shipment) => {
    if (hasValidCoordinates(shipment)) {
      const coords = shipment.routeData.coordinates;
      const lat = coords.latitude;
      const lng = coords.longitude;

      validCoordinates.push([lat, lng]);

      // Create marker with custom status-based icon
      const customIcon = createStatusIcon(shipment.shippingStatus);
      const marker = L.marker([lat, lng], { icon: customIcon }).addTo(
        currentMapInstance,
      );

      // Store shipment data with marker for selection highlighting
      marker._shipmentData = shipment;

      // Create popup content
      const popupContent = createShipmentPopup(shipment);
      marker.bindPopup(popupContent);

      markers.push(marker);
    } else {
      shipmentsWithoutCoordinates.push(shipment);
    }
  });

  // Draw routes for shipments with route data
  drawShipmentRoutes(visibleShipments);

  // Fit map to show all markers if we have valid coordinates
  if (validCoordinates.length > 0) {
    fitMapToMarkers(currentMapInstance, validCoordinates);
  }

  // Highlight selected shipments
  if (selectedShipments.length > 0) {
    highlightSelectedShipments(selectedShipments);
  }

  // Show notifications
  if (shipmentsWithoutCoordinates.length > 0) {
    showNoCoordinatesNotification(shipmentsWithoutCoordinates.length);
  }

  if (validCoordinates.length === 0) {
    showNoRoutesNotification();
  }
}

/**
 * Highlight selected shipments on the map
 */
function highlightSelectedShipments(selectedShipments) {
  if (!selectedShipments || selectedShipments.length === 0) {
    // Remove highlighting from all markers
    markers.forEach((marker) => {
      if (marker._originalIcon) {
        marker.setIcon(marker._originalIcon);
        delete marker._originalIcon;
      }
    });
    return;
  }

  const selectedIds = selectedShipments.map((s) => s.id);

  markers.forEach((marker) => {
    const shipment = marker._shipmentData;
    if (shipment && selectedIds.includes(shipment.id)) {
      // Highlight selected marker
      if (!marker._originalIcon) {
        marker._originalIcon = marker.getIcon();
        // Create highlighted version of the icon
        const highlightedIcon = createHighlightedIcon(shipment.shippingStatus);
        marker.setIcon(highlightedIcon);
      }
    } else {
      // Remove highlighting
      if (marker._originalIcon) {
        marker.setIcon(marker._originalIcon);
        delete marker._originalIcon;
      }
    }
  });
}

/**
 * Clear all markers and routes from the map
 */
function clearMapElements() {
  // Remove markers
  markers.forEach((marker) => {
    currentMapInstance.removeLayer(marker);
  });
  markers = [];

  // Remove routes
  routes.forEach((route) => {
    currentMapInstance.removeLayer(route);
  });
  routes = [];
}

/**
 * Check if shipment has valid coordinates
 */
function hasValidCoordinates(shipment) {
  if (!shipment.routeData || !shipment.routeData.coordinates) return false;

  const coords = shipment.routeData.coordinates;
  const lat = coords.latitude;
  const lng = coords.longitude;

  return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180;
}

/**
 * Get selected shipments from grid API
 */
function getSelectedShipments(gridApi) {
  if (!gridApi) return [];

  try {
    const selectedNodes = gridApi.getSelectedNodes();
    return selectedNodes
      .filter((node) => node && node.data)
      .map((node) => node.data);
  } catch (error) {
    console.error("Error getting selected shipments:", error);
    return [];
  }
}

// Re-export original functions with enhanced versions
export { updateMapWithShipments as updateMapMarkers };

// Import and re-export necessary functions from original handle-map.js
// These would need to be copied or imported from the original file
export function createStatusIcon(status) {
  const iconColors = {
    IN_TRANSIT: "#3B82F6", // Blue
    DELIVERED: "#10B981", // Green
    PLANNED: "#F59E0B", // Yellow
    UNKNOWN: "#6B7280", // Gray
  };

  const color = iconColors[status] || iconColors["UNKNOWN"];

  return L.divIcon({
    className: "custom-marker",
    html: `
      <div style="
        background-color: ${color};
        width: 20px;
        height: 20px;
        border-radius: 50%;
        border: 3px solid white;
        box-shadow: 0 2px 5px rgba(0,0,0,0.3);
        display: flex;
        align-items: center;
        justify-content: center;
      ">
        <div style="
          width: 8px;
          height: 8px;
          background-color: white;
          border-radius: 50%;
        "></div>
      </div>
    `,
    iconSize: [26, 26],
    iconAnchor: [13, 13],
  });
}

export function createHighlightedIcon(status) {
  const iconColors = {
    IN_TRANSIT: "#3B82F6",
    DELIVERED: "#10B981",
    PLANNED: "#F59E0B",
    UNKNOWN: "#6B7280",
  };

  const color = iconColors[status] || iconColors["UNKNOWN"];

  return L.divIcon({
    className: "custom-marker highlighted",
    html: `
      <div style="
        background-color: ${color};
        width: 24px;
        height: 24px;
        border-radius: 50%;
        border: 4px solid #FFD700;
        box-shadow: 0 0 15px rgba(255,215,0,0.8);
        display: flex;
        align-items: center;
        justify-content: center;
        animation: pulse 2s infinite;
      ">
        <div style="
          width: 10px;
          height: 10px;
          background-color: white;
          border-radius: 50%;
        "></div>
      </div>
      <style>
        @keyframes pulse {
          0% { transform: scale(1); }
          50% { transform: scale(1.1); }
          100% { transform: scale(1); }
        }
      </style>
    `,
    iconSize: [32, 32],
    iconAnchor: [16, 16],
  });
}

// Complete implementations from original handle-map.js

export function addMapControls(map, gridApi) {
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
      if (gridApi) {
        updateMapWithShipments(getVisibleShipments(gridApi));
      } else {
        mapDataService.requestInitialData();
      }
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

// Add missing functions for the additional controls
export function optimizeRouteDisplay(map) {
  console.log("Optimizing route display...");
  // Implementation for route optimization
}

export function showRouteStatistics(map, gridApi) {
  console.log("Showing route statistics...");
  // Implementation for route statistics
}

export function drawShipmentRoutes(shipments) {
  if (!currentMapInstance || !shipments.length) return;

  // Use original route drawing logic for curved routes
  shipments.forEach((shipment) => {
    const shipmentRoutes = drawShipmentRoutesForShipment(
      currentMapInstance,
      shipment,
    );
    routes.push(...shipmentRoutes);
  });
}

function drawShipmentRoutesForShipment(map, shipment) {
  const shipmentRoutes = [];

  if (!shipment.routeData || !shipment.routeData.routeSegments) {
    return shipmentRoutes;
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

    shipmentRoutes.push(polyline);
  });

  return shipmentRoutes;
}

function getOptimalRouteWeight(map, coordinates, baseWeight) {
  // Simple implementation - can be enhanced based on zoom level and overlaps
  const zoom = map.getZoom();
  if (zoom < 4) return Math.max(1, baseWeight - 1);
  if (zoom < 8) return baseWeight;
  return baseWeight + 1;
}

function createRoutePopupContent(
  shipment,
  routeType,
  segmentIndex,
  totalSegments,
  coordinates,
  color,
) {
  const distance = calculateRouteDistance(coordinates);
  return `
    <div class="route-popup">
      <h4>${shipment.shipmentNumber}</h4>
      <p><strong>Route Type:</strong> ${routeType}</p>
      <p><strong>Segment:</strong> ${segmentIndex + 1} of ${totalSegments.length}</p>
      <p><strong>Distance:</strong> ~${distance.toFixed(0)} km</p>
      <p><strong>Status:</strong> ${shipment.shippingStatus}</p>
    </div>
  `;
}

function calculateRouteDistance(coordinates) {
  if (!coordinates || coordinates.length < 2) return 0;

  let totalDistance = 0;
  for (let i = 1; i < coordinates.length; i++) {
    const [lat1, lng1] = coordinates[i - 1];
    const [lat2, lng2] = coordinates[i];
    totalDistance += getDistanceBetweenPoints(lat1, lng1, lat2, lng2);
  }

  return totalDistance;
}

function getDistanceBetweenPoints(lat1, lng1, lat2, lng2) {
  const R = 6371; // Radius of Earth in kilometers
  const dLat = ((lat2 - lat1) * Math.PI) / 180;
  const dLng = ((lng2 - lng1) * Math.PI) / 180;
  const a =
    Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos((lat1 * Math.PI) / 180) *
      Math.cos((lat2 * Math.PI) / 180) *
      Math.sin(dLng / 2) *
      Math.sin(dLng / 2);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  return R * c;
}

export function fitMapToMarkers(map, coordinates) {
  if (coordinates.length === 0) return;

  if (coordinates.length === 1) {
    map.setView(coordinates[0], 10);
  } else {
    const group = new L.FeatureGroup(
      coordinates.map((coord) => L.marker(coord)),
    );
    map.fitBounds(group.getBounds().pad(0.1));
  }
}

export function createShipmentPopup(shipment) {
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

export function showNoCoordinatesNotification(count) {
  console.warn(`${count} shipments don't have coordinate data`);
}

export function showNoRoutesNotification() {
  console.warn("No shipments with valid coordinates found");
}

/**
 * Initialize enhanced map functionality
 * Can be called from main.js or standalone map page
 */
export function initEnhancedMap(gridApi = null, options = {}) {
  // Check if MapDataService is available
  if (!mapDataService.isServiceConnected()) {
    console.error(
      "❌ MapDataService not available. Enhanced features disabled.",
    );
    return null;
  }

  return handleMapEnhanced(gridApi, options);
}

// Debug function
/**
 * Toggle markers visibility on the map
 */
export function toggleMarkers(map) {
  if (!map || !markers.length) return;

  const firstMarker = markers[0];
  const areMarkersVisible = map.hasLayer(firstMarker);

  markers.forEach((marker) => {
    if (areMarkersVisible) {
      map.removeLayer(marker);
    } else {
      map.addLayer(marker);
    }
  });

  console.log(`🔄 Markers ${areMarkersVisible ? "hidden" : "shown"}`);
}

/**
 * Toggle routes visibility on the map
 */
export function toggleRoutes(map) {
  if (!map || !routes.length) return;

  const firstRoute = routes[0];
  const areRoutesVisible = map.hasLayer(firstRoute);

  routes.forEach((route) => {
    if (areRoutesVisible) {
      map.removeLayer(route);
    } else {
      map.addLayer(route);
    }
  });

  console.log(`🔄 Routes ${areRoutesVisible ? "hidden" : "shown"}`);
}

export function getMapStatus() {
  return {
    mapInitialized: !!currentMapInstance,
    isStandalone: isStandaloneMap,
    markerCount: markers.length,
    routeCount: routes.length,
    serviceStatus: mapDataService.getStatus(),
  };
}

// Export updateMapWithShipments for standalone map usage
export { updateMapWithShipments };

// Make functions globally available for debugging and standalone usage
if (typeof window !== "undefined") {
  window.updateMapWithShipments = updateMapWithShipments;
  window.getMapStatus = getMapStatus;
}
