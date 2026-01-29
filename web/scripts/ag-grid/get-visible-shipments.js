export function getVisibleShipments(gridApi, options = {}) {
  const { includeRenderedOnly = false } = options;

  if (!gridApi) {
    console.error("getVisibleShipments: gridApi is required");
    return [];
  }

  if (typeof gridApi.forEachNodeAfterFilterAndSort !== "function") {
    console.error("getVisibleShipments: gridApi missing required methods");
    return [];
  }

  const shipments = [];

  try {
    if (includeRenderedOnly) {
      const renderedNodes = gridApi.getRenderedNodes();
      renderedNodes.forEach((node) => {
        if (node && node.data) {
          shipments.push(node.data);
        }
      });
    } else {
      gridApi.forEachNodeAfterFilterAndSort((rowNode) => {
        if (rowNode && rowNode.data) {
          shipments.push(rowNode.data);
        }
      });
    }

    return shipments;
  } catch (error) {
    console.error("Error getting visible shipments:", error);
    return [];
  }
}

export function getRenderedShipments(gridApi) {
  return getVisibleShipments(gridApi, { includeRenderedOnly: true });
}

export function getVisibleShipmentNodes(gridApi) {
  if (!gridApi) {
    console.error("getVisibleShipmentNodes: gridApi is required");
    return [];
  }

  const nodes = [];

  try {
    gridApi.forEachNodeAfterFilterAndSort((rowNode) => {
      if (rowNode && rowNode.data) {
        nodes.push(rowNode);
      }
    });
    return nodes;
  } catch (error) {
    console.error("Error getting visible shipment nodes:", error);
    return [];
  }
}

export function getSelectedShipments(gridApi) {
  if (!gridApi) {
    console.error("getSelectedShipments: gridApi is required");
    return [];
  }

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