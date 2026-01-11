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

  initialize() {
    try {
      this.channel = new BroadcastChannel(this.channelName);
      this.channel.addEventListener("message", this.handleMessage.bind(this));
      this.isConnected = true;
      this.requestInitialData();
    } catch (error) {
      console.error("Failed to initialize MapDataService:", error);
      this.isConnected = false;
    }
  }

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

  requestInitialData() {
    if (!this.isConnected) return;

    const message = {
      type: "REQUEST_INITIAL_DATA",
      data: { requesterId: Date.now() },
      timestamp: new Date().toISOString(),
    };

    this.channel.postMessage(message);
  }

  handleShipmentsUpdate(data) {
    this.currentData.visibleShipments = data.visibleShipments || [];
    this.currentData.selectedShipments = data.selectedShipments || [];
    this.notifyListeners("shipmentsUpdate", data);
  }

  handleGridStateUpdate(data) {
    this.currentData.gridState = { ...this.currentData.gridState, ...data };
    this.notifyListeners("gridStateUpdate", data);
  }

  handleSelectionUpdate(data) {
    this.currentData.selectedShipments = data.selectedShipments || [];
    this.notifyListeners("selectionUpdate", data.selectedShipments);
  }

  handleInitialDataRequest() {
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
  }

  handleInitialDataResponse(data) {
    this.currentData = { ...this.currentData, ...data };
    this.notifyListeners("initialDataReceived", data);
  }

  handleMapViewChange(data) {
    this.notifyListeners("mapViewChange", data);
  }

  addEventListener(eventType, callback) {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, []);
    }
    this.listeners.get(eventType).push(callback);

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

  removeEventListener(eventType, callback) {
    if (this.listeners.has(eventType)) {
      const callbacks = this.listeners.get(eventType);
      const index = callbacks.indexOf(callback);
      if (index > -1) {
        callbacks.splice(index, 1);
      }
    }
  }

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

  getCurrentData() {
    return { ...this.currentData };
  }

  isServiceConnected() {
    return this.isConnected;
  }

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

  destroy() {
    if (this.channel) {
      this.channel.close();
    }
    this.listeners.clear();
    this.isConnected = false;
  }
}

const mapDataService = new MapDataService();

export { MapDataService, mapDataService };

if (typeof window !== "undefined") {
  window.mapDataService = mapDataService;
}