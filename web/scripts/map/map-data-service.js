/**
 * Map Data Service - Handles cross-page communication for map functionality
 * Uses BroadcastChannel API for real-time synchronization between grid and map pages
 */

class MapDataService {
  constructor() {
    this.channelName = "shipment-map-data";
    this.channel = null;
    this.listeners = new Map();
    this.currentData = {
      visibleShipments: [],
      selectedShipments: [],
      gridState: {
        filters: {},
        sorting: [],
        pagination: { currentPage: 1, pageSize: 15 },
      },
      lastUpdate: null,
    };
    this.isConnected = false;
    this.initialize();
  }

  /**
   * Initialize the BroadcastChannel and set up event listeners
   */
  initialize() {
    try {
      this.channel = new BroadcastChannel(this.channelName);
      this.channel.addEventListener("message", this.handleMessage.bind(this));
      this.isConnected = true;
      console.log("✅ MapDataService initialized successfully");

      // Request initial data when a new page connects
      this.requestInitialData();
    } catch (error) {
      console.error("❌ Failed to initialize MapDataService:", error);
      this.isConnected = false;
    }
  }

  /**
   * Handle incoming messages from other pages
   */
  handleMessage(event) {
    const { type, data, timestamp } = event.data;

    switch (type) {
      case "SHIPMENTS_UPDATE":
        this.handleShipmentsUpdate(data);
        break;
      case "GRID_STATE_UPDATE":
        this.handleGridStateUpdate(data);
        break;
      case "SELECTION_UPDATE":
        this.handleSelectionUpdate(data);
        break;
      case "REQUEST_INITIAL_DATA":
        this.handleInitialDataRequest();
        break;
      case "INITIAL_DATA_RESPONSE":
        this.handleInitialDataResponse(data);
        break;
      case "MAP_VIEW_CHANGE":
        this.handleMapViewChange(data);
        break;
      default:
        console.warn("Unknown message type:", type);
    }
  }

  /**
   * Send shipments data to all connected pages
   */
  broadcastShipments(visibleShipments, selectedShipments = []) {
    if (!this.isConnected) {
      console.warn("MapDataService not connected");
      return;
    }

    this.currentData.visibleShipments = visibleShipments;
    this.currentData.selectedShipments = selectedShipments;
    this.currentData.lastUpdate = new Date().toISOString();

    const message = {
      type: "SHIPMENTS_UPDATE",
      data: {
        visibleShipments,
        selectedShipments,
        shipmentCount: visibleShipments.length,
        selectedCount: selectedShipments.length,
      },
      timestamp: this.currentData.lastUpdate,
    };

    this.channel.postMessage(message);
    this.notifyListeners("shipmentsUpdate", message.data);
  }

  /**
   * Send grid state changes to all connected pages
   */
  broadcastGridState(gridState) {
    if (!this.isConnected) return;

    this.currentData.gridState = gridState;
    this.currentData.lastUpdate = new Date().toISOString();

    const message = {
      type: "GRID_STATE_UPDATE",
      data: gridState,
      timestamp: this.currentData.lastUpdate,
    };

    this.channel.postMessage(message);
    this.notifyListeners("gridStateUpdate", gridState);
  }

  /**
   * Send selection changes to all connected pages
   */
  broadcastSelection(selectedShipments) {
    if (!this.isConnected) return;

    this.currentData.selectedShipments = selectedShipments;
    this.currentData.lastUpdate = new Date().toISOString();

    const message = {
      type: "SELECTION_UPDATE",
      data: { selectedShipments },
      timestamp: this.currentData.lastUpdate,
    };

    this.channel.postMessage(message);
    this.notifyListeners("selectionUpdate", selectedShipments);
  }

  /**
   * Send map view changes (bounds, zoom, center) to other pages
   */
  broadcastMapView(viewData) {
    if (!this.isConnected) return;

    const message = {
      type: "MAP_VIEW_CHANGE",
      data: viewData,
      timestamp: new Date().toISOString(),
    };

    this.channel.postMessage(message);
    this.notifyListeners("mapViewChange", viewData);
  }

  /**
   * Request initial data from other pages (used when a new page loads)
   */
  requestInitialData() {
    if (!this.isConnected) return;

    const message = {
      type: "REQUEST_INITIAL_DATA",
      data: { requesterId: Date.now() },
      timestamp: new Date().toISOString(),
    };

    this.channel.postMessage(message);
  }

  /**
   * Handle shipments data update
   */
  handleShipmentsUpdate(data) {
    this.currentData.visibleShipments = data.visibleShipments || [];
    this.currentData.selectedShipments = data.selectedShipments || [];
    this.notifyListeners("shipmentsUpdate", data);
  }

  /**
   * Handle grid state update
   */
  handleGridStateUpdate(data) {
    this.currentData.gridState = { ...this.currentData.gridState, ...data };
    this.notifyListeners("gridStateUpdate", data);
  }

  /**
   * Handle selection update
   */
  handleSelectionUpdate(data) {
    this.currentData.selectedShipments = data.selectedShipments || [];
    this.notifyListeners("selectionUpdate", data.selectedShipments);
  }

  /**
   * Handle request for initial data
   */
  handleInitialDataRequest() {
    // Always respond to let requester know this page exists
    const message = {
      type: "INITIAL_DATA_RESPONSE",
      data: {
        ...this.currentData,
        hasData: this.currentData.visibleShipments.length > 0,
        isGridPage: true,
      },
      timestamp: new Date().toISOString(),
    };
    this.channel.postMessage(message);
    console.log(
      `📡 Responding to initial data request with ${this.currentData.visibleShipments.length} shipments`,
    );
  }

  /**
   * Handle initial data response
   */
  handleInitialDataResponse(data) {
    this.currentData = { ...this.currentData, ...data };
    this.notifyListeners("initialDataReceived", data);
  }

  /**
   * Handle map view change
   */
  handleMapViewChange(data) {
    this.notifyListeners("mapViewChange", data);
  }

  /**
   * Add event listener for specific events
   */
  addEventListener(eventType, callback) {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, []);
    }
    this.listeners.get(eventType).push(callback);

    // If we already have data and this is a data-related event, call immediately
    if (
      eventType === "shipmentsUpdate" &&
      this.currentData.visibleShipments.length > 0
    ) {
      setTimeout(
        () =>
          callback({
            visibleShipments: this.currentData.visibleShipments,
            selectedShipments: this.currentData.selectedShipments,
            shipmentCount: this.currentData.visibleShipments.length,
            selectedCount: this.currentData.selectedShipments.length,
          }),
        0,
      );
    }
  }

  /**
   * Remove event listener
   */
  removeEventListener(eventType, callback) {
    if (this.listeners.has(eventType)) {
      const callbacks = this.listeners.get(eventType);
      const index = callbacks.indexOf(callback);
      if (index > -1) {
        callbacks.splice(index, 1);
      }
    }
  }

  /**
   * Notify all listeners of a specific event type
   */
  notifyListeners(eventType, data) {
    if (this.listeners.has(eventType)) {
      this.listeners.get(eventType).forEach((callback) => {
        try {
          callback(data);
        } catch (error) {
          console.error(`Error in ${eventType} listener:`, error);
        }
      });
    }
  }

  /**
   * Get current data
   */
  getCurrentData() {
    return { ...this.currentData };
  }

  /**
   * Check if service is connected
   */
  isServiceConnected() {
    return this.isConnected;
  }

  /**
   * Get connection status and data summary
   */
  getStatus() {
    return {
      connected: this.isConnected,
      hasData: this.currentData.visibleShipments.length > 0,
      shipmentCount: this.currentData.visibleShipments.length,
      selectedCount: this.currentData.selectedShipments.length,
      lastUpdate: this.currentData.lastUpdate,
      listenerCount: Array.from(this.listeners.values()).reduce(
        (sum, arr) => sum + arr.length,
        0,
      ),
    };
  }

  /**
   * Cleanup resources
   */
  destroy() {
    if (this.channel) {
      this.channel.close();
    }
    this.listeners.clear();
    this.isConnected = false;
    console.log("🧹 MapDataService destroyed");
  }
}

// Create singleton instance
const mapDataService = new MapDataService();

// Export both the class and the singleton instance
export { MapDataService, mapDataService };

// Also make it available globally for debugging
if (typeof window !== "undefined") {
  window.mapDataService = mapDataService;
}
