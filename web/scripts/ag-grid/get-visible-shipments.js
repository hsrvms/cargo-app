/**
 * Get all visible shipments after filters and sorting are applied
 * @param {Object} gridApi - AG Grid API instance
 * @param {Object} options - Configuration options
 * @returns {Array} Array of shipment data objects
 */
export function getVisibleShipments(gridApi, options = {}) {
  const { includeRenderedOnly = false, debug = false } = options; // Enable debug by default

  if (!gridApi) {
    console.error("getVisibleShipments: gridApi is required");
    return [];
  }

  // Validate gridApi has required methods
  if (typeof gridApi.forEachNodeAfterFilterAndSort !== "function") {
    console.error("getVisibleShipments: gridApi missing required methods");
    return [];
  }

  const shipments = [];

  try {
    // Get grid state information for debugging
    let totalRowCount = 0;
    let filteredRowCount = 0;
    let renderedRowCount = 0;

    // Count total rows
    if (typeof gridApi.forEachNode === "function") {
      gridApi.forEachNode(() => totalRowCount++);
    }

    // Count filtered rows
    gridApi.forEachNodeAfterFilter(() => filteredRowCount++);

    // Count rendered rows
    const renderedNodes = gridApi.getRenderedNodes();
    renderedRowCount = renderedNodes.length;

    if (debug) {
      console.group("Grid State Debug Info:");
      console.log("Total rows in grid:", totalRowCount);
      console.log("Rows after filter:", filteredRowCount);
      console.log("Rendered rows (viewport):", renderedRowCount);
      console.log("Include rendered only:", includeRenderedOnly);

      // Check if grid has data
      if (totalRowCount === 0) {
        console.warn("⚠️ Grid has no data loaded");
      } else if (filteredRowCount === 0) {
        console.warn("⚠️ All rows are filtered out");
      } else if (renderedRowCount === 0 && includeRenderedOnly) {
        console.warn("⚠️ No rows are currently rendered in viewport");
      }
    }

    if (includeRenderedOnly) {
      // Get only nodes that are currently rendered on screen (visible in viewport)
      renderedNodes.forEach((node) => {
        if (node && node.data) {
          shipments.push(node.data);
          if (debug && shipments.length <= 3) {
            // Log first 3 for brevity
            console.log("Adding rendered shipment:", node.data);
          }
        }
      });
    } else {
      // Collect all filtered and sorted nodes (visible after filters are applied)
      gridApi.forEachNodeAfterFilterAndSort((rowNode) => {
        if (rowNode && rowNode.data) {
          shipments.push(rowNode.data);
          if (debug && shipments.length <= 3) {
            // Log first 3 for brevity
            console.log("Adding filtered shipment:", rowNode.data);
          }
        }
      });
    }

    if (debug) {
      console.log("✅ Total visible shipments collected:", shipments.length);
      if (shipments.length > 3) {
        console.log("... (showing first 3 shipments only in logs)");
      }
      console.groupEnd();
    }

    return shipments;
  } catch (error) {
    console.error("Error getting visible shipments:", error);
    console.error(
      "GridApi methods available:",
      Object.getOwnPropertyNames(gridApi),
    );
    return [];
  }
}

/**
 * Get only shipments that are currently rendered in the viewport
 * @param {Object} gridApi - AG Grid API instance
 * @returns {Array} Array of rendered shipment data objects
 */
export function getRenderedShipments(gridApi) {
  return getVisibleShipments(gridApi, { includeRenderedOnly: true });
}

/**
 * Get all shipment nodes (not just data) after filters and sorting
 * @param {Object} gridApi - AG Grid API instance
 * @returns {Array} Array of AG Grid row nodes
 */
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

/**
 * Get selected shipments from the grid
 * @param {Object} gridApi - AG Grid API instance
 * @returns {Array} Array of selected shipment data objects
 */
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
