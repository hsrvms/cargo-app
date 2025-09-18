/**
 * Filter Manager for AG Grid - API-based version that stores filters in database
 */
class FilterManager {
  constructor(gridApi) {
    this.gridApi = gridApi;

    // Check specific methods we need
    const requiredMethods = [
      "getFilterModel",
      "setFilterModel",
      "getSortModel",
      "setSortModel",
      "getColumnState",
      "applyColumnState",
    ];
    requiredMethods.forEach((method) => {
      const available = typeof gridApi[method] === "function";
    });

    this.init();
  }

  /**
   * Initialize the filter manager
   */
  init() {
    this.setupEventListeners();
    this.loadSavedFiltersToDropdown();
  }

  /**
   * Setup event listeners for filter management
   */
  setupEventListeners() {
    // Save filter button
    const saveBtn = document.getElementById("saveFilterBtn");
    if (saveBtn) {
      saveBtn.addEventListener("click", () => this.handleSaveFilter());
    }

    // Load filter dropdown
    const dropdown = document.getElementById("savedFiltersDropdown");
    if (dropdown) {
      dropdown.addEventListener("change", (e) => {
        if (e.target.value) {
          this.loadFilter(e.target.value);
        }
      });
    }

    // Clear all filters button
    const clearBtn = document.getElementById("clearFiltersBtn");
    if (clearBtn) {
      clearBtn.addEventListener("click", () => this.clearAllFilters());
    }

    // Clear all saved filters button
    const clearSavedBtn = document.getElementById("clearAllFiltersBtn");
    if (clearSavedBtn) {
      clearSavedBtn.addEventListener("click", () =>
        this.clearAllSavedFilters(),
      );
    }

    // Manage filters button
    const manageBtn = document.getElementById("manageFiltersBtn");
    if (manageBtn) {
      manageBtn.addEventListener("click", () =>
        this.showFilterManagementPanel(),
      );
    }

    // Export filters button
    const exportBtn = document.getElementById("exportFiltersBtn");
    if (exportBtn) {
      exportBtn.addEventListener("click", () => this.handleExportFilters());
    }

    // Import filters input
    const importInput = document.getElementById("importFiltersInput");
    if (importInput) {
      importInput.addEventListener("change", (e) =>
        this.handleImportFilters(e),
      );
    }
  }

  /**
   * Get the current filter state from the grid
   * @returns {Object} Current filter state
   */
  getCurrentFilterState() {
    try {
      const filterState = {
        filters: {},
        columns: [],
        sorting: [],
        activeFilterCount: 0,
      };

      // Get filter model
      if (typeof this.gridApi.getFilterModel === "function") {
        filterState.filters = this.gridApi.getFilterModel() || {};
        filterState.activeFilterCount = Object.keys(filterState.filters).length;
      }

      // Get column state
      if (typeof this.gridApi.getColumnState === "function") {
        filterState.columns = this.gridApi.getColumnState() || [];
      }

      // Get sort model
      if (typeof this.gridApi.getSortModel === "function") {
        filterState.sorting = this.gridApi.getSortModel() || [];
      }

      return filterState;
    } catch (error) {
      console.error("Error getting current filter state:", error);
      return {
        filters: {},
        columns: [],
        sorting: [],
        activeFilterCount: 0,
      };
    }
  }

  /**
   * Handle save filter button click
   */
  async handleSaveFilter() {
    const filterState = this.getCurrentFilterState();

    if (filterState.activeFilterCount === 0) {
      this.showToast("No filters to save. Apply some filters first.", "info");
      return;
    }

    const name = await this.promptForFilterName();
    if (!name) return;

    await this.saveFilter(name, filterState);
  }

  /**
   * Prompt user for filter name
   * @returns {Promise<string|null>} Filter name or null if cancelled
   */
  async promptForFilterName() {
    return new Promise((resolve) => {
      const name = prompt("Enter a name for this filter:");
      if (name && name.trim()) {
        resolve(name.trim());
      } else {
        resolve(null);
      }
    });
  }

  /**
   * Save filter to database via API
   * @param {string} name - Filter name
   * @param {Object} filterState - Filter state object
   */
  async saveFilter(name, filterState) {
    try {
      // Validate filter state before saving
      if (!this.validateFilterState(filterState)) {
        this.showToast("Invalid filter state. Cannot save filter.", "error");
        return;
      }

      const filterData = {
        ...filterState,
        version: "1.0",
        userAgent: navigator.userAgent.split(" ")[0], // Simple browser detection
      };

      const response = await fetch("/api/filters", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          name: name,
          filter_data: filterData,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Failed to save filter");
      }

      const result = await response.json();
      await this.loadSavedFiltersToDropdown();

      // Show detailed success message
      const filterCount = filterState.activeFilterCount || 0;
      const hasColumns = (filterState.columns || []).length > 0;
      const hasSorting = (filterState.sorting || []).length > 0;

      let message = `"${name}" saved with ${filterCount} filter${filterCount !== 1 ? "s" : ""}`;
      if (hasColumns) message += ", column layout";
      if (hasSorting) message += ", sorting";
      message += "!";

      this.showToast(message, "success");
    } catch (error) {
      console.error("Error saving filter:", error);
      this.showToast("Error saving filter: " + error.message, "error");
    }
  }

  /**
   * Load filter and apply it to the grid
   * @param {string} name - Filter name to load
   */
  async loadFilter(name) {
    try {
      const response = await fetch(
        `/api/filters/by-name/${encodeURIComponent(name)}`,
        {
          method: "GET",
          credentials: "include",
        },
      );

      if (!response.ok) {
        if (response.status === 404) {
          this.showToast(`Filter "${name}" not found.`, "error");
          return;
        }
        const errorData = await response.json();
        throw new Error(errorData.error || "Failed to load filter");
      }

      const result = await response.json();
      const filterState = result.filter_data;

      // Apply filters
      if (
        filterState.filters &&
        typeof this.gridApi.setFilterModel === "function"
      ) {
        this.gridApi.setFilterModel(filterState.filters);
      }

      // Apply column state first (includes order, width, visibility)
      if (
        filterState.columns &&
        typeof this.gridApi.applyColumnState === "function"
      ) {
        this.gridApi.applyColumnState({
          state: filterState.columns,
          applyOrder: true,
        });
      }

      // Apply sorting after columns
      if (
        filterState.sorting &&
        typeof this.gridApi.setSortModel === "function"
      ) {
        this.gridApi.setSortModel(filterState.sorting);
      }

      this.showToast(`Filter "${name}" applied successfully!`, "success");
    } catch (error) {
      console.error("Error loading filter:", error);
      this.showToast("Error loading filter: " + error.message, "error");
    }
  }

  /**
   * Delete a saved filter
   * @param {string} name - Filter name to delete
   */
  async deleteFilter(name) {
    try {
      const confirmDelete = confirm(
        `Are you sure you want to delete the filter "${name}"?`,
      );
      if (!confirmDelete) return;

      const response = await fetch(
        `/api/filters/by-name/${encodeURIComponent(name)}`,
        {
          method: "DELETE",
          credentials: "include",
        },
      );

      if (!response.ok) {
        if (response.status === 404) {
          this.showToast(`Filter "${name}" not found.`, "error");
          return;
        }
        const errorData = await response.json();
        throw new Error(errorData.error || "Failed to delete filter");
      }

      await this.loadSavedFiltersToDropdown();

      // Refresh management panel if it's open
      if (
        document.getElementById("filter-management-panel") &&
        !document
          .getElementById("filter-management-panel")
          .classList.contains("hidden")
      ) {
        this.renderFilterList();
      }

      // Reset dropdown selection
      const dropdown = document.getElementById("savedFiltersDropdown");
      if (dropdown) {
        dropdown.value = "";
      }

      this.showToast(`Filter "${name}" deleted successfully!`, "success");
    } catch (error) {
      console.error("Error deleting filter:", error);
      this.showToast("Error deleting filter: " + error.message, "error");
    }
  }

  /**
   * Get all saved filters from API
   * @returns {Object} Saved filters object
   */
  async getSavedFilters() {
    try {
      const response = await fetch("/api/filters", {
        method: "GET",
        credentials: "include",
      });

      if (!response.ok) {
        if (response.status === 401) {
          console.log("User not authenticated, returning empty filters");
          return {};
        }
        throw new Error("Failed to fetch filters");
      }

      const result = await response.json();
      const filters = {};

      // Convert API response to the expected format
      result.filters.forEach((filter) => {
        filters[filter.name] = {
          ...filter.filter_data,
          savedAt: filter.created_at,
          name: filter.name,
        };
      });

      return filters;
    } catch (error) {
      console.error("Error reading saved filters:", error);
      return {};
    }
  }

  /**
   * Load saved filters into the dropdown
   */
  async loadSavedFiltersToDropdown() {
    const dropdown = document.getElementById("savedFiltersDropdown");
    if (!dropdown) return;

    try {
      const savedFilters = await this.getSavedFilters();
      const filterNames = Object.keys(savedFilters).sort();

      // Clear existing options except the first one
      dropdown.innerHTML = '<option value="">Select Saved Filter...</option>';

      if (filterNames.length === 0) {
        const option = document.createElement("option");
        option.disabled = true;
        option.textContent = "No saved filters";
        dropdown.appendChild(option);
        return;
      }

      // Add filter options with metadata
      filterNames.forEach((name) => {
        const filter = savedFilters[name];
        const option = document.createElement("option");
        option.value = name;

        // Create descriptive text
        let text = name;
        if (filter.activeFilterCount > 0) {
          text += ` (${filter.activeFilterCount} filter${filter.activeFilterCount !== 1 ? "s" : ""})`;
        }

        option.textContent = text;

        // Add tooltip with additional info
        const savedDate = new Date(filter.savedAt).toLocaleDateString();
        option.title = `Saved: ${savedDate}`;
        if (filter.version) option.title += ` | Version: ${filter.version}`;
        if (filter.userAgent) option.title += ` | Browser: ${filter.userAgent}`;

        dropdown.appendChild(option);

        // Add delete button for each filter (inline)
        const deleteBtn = document.createElement("button");
        deleteBtn.textContent = "×";
        deleteBtn.className = "filter-delete-btn";
        deleteBtn.style.cssText =
          "margin-left:5px;font-size:12px;color:red;background:none;border:none;cursor:pointer;";
        deleteBtn.title = `Delete filter: ${name}`;
        deleteBtn.onclick = (e) => {
          e.stopPropagation();
          this.deleteFilter(name);
        };

        // Note: Adding buttons to select options is not standard HTML
        // This is for demonstration. In practice, you'd have separate delete buttons.
      });

      // Update management panel if it exists
      this.renderFilterList();
    } catch (error) {
      console.error("Error loading saved filters to dropdown:", error);
      dropdown.innerHTML = '<option value="">Error loading filters</option>';
    }
  }

  /**
   * Clear all active filters from the grid
   */
  clearAllFilters() {
    try {
      if (typeof this.gridApi.setFilterModel === "function") {
        this.gridApi.setFilterModel(null);
      }
      if (typeof this.gridApi.setSortModel === "function") {
        this.gridApi.setSortModel(null);
      }

      // Reset dropdown selection
      const dropdown = document.getElementById("savedFiltersDropdown");
      if (dropdown) {
        dropdown.value = "";
      }

      this.showToast("All filters and sorting cleared!", "info");
    } catch (error) {
      console.error("Error clearing filters:", error);
      this.showToast("Error clearing filters. Please try again.", "error");
    }
  }

  /**
   * Export filters to JSON file
   * @returns {void}
   */
  exportFilters() {
    // This method is kept for backward compatibility but doesn't do anything
    // since filters are now stored in the database
    this.showToast(
      "Export feature not needed - filters are now stored in your account!",
      "info",
    );
  }

  /**
   * Import filters from JSON file
   * @param {Event} event - File input change event
   * @returns {Promise<void>}
   */
  async importFilters(event) {
    const file = event.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = async (e) => {
      try {
        const importData = JSON.parse(e.target.result);

        // Validate import data structure
        if (!importData.filters || typeof importData.filters !== "object") {
          throw new Error("Invalid filter data format");
        }

        const importedFilters = importData.filters;
        const importedCount = Object.keys(importedFilters).length;

        if (importedCount === 0) {
          this.showToast("No filters found in the imported file.", "info");
          return;
        }

        // Import each filter via API
        let successCount = 0;
        let errorCount = 0;

        for (const [name, filterData] of Object.entries(importedFilters)) {
          try {
            await this.saveFilter(name, filterData);
            successCount++;
          } catch (error) {
            console.error(`Error importing filter ${name}:`, error);
            errorCount++;
          }
        }

        await this.loadSavedFiltersToDropdown();

        if (errorCount === 0) {
          this.showToast(
            `Successfully imported ${successCount} filter${successCount !== 1 ? "s" : ""}!`,
            "success",
          );
        } else {
          this.showToast(
            `Imported ${successCount} filter${successCount !== 1 ? "s" : ""} with ${errorCount} error${errorCount !== 1 ? "s" : ""}.`,
            "warning",
          );
        }
      } catch (error) {
        console.error("Error importing filters:", error);
        this.showToast("Error importing filters: " + error.message, "error");
      }
    };

    reader.readAsText(file);
    // Clear the input so the same file can be selected again
    event.target.value = "";
  }

  /**
   * Validate filter state object
   * @param {Object} filterState - Filter state to validate
   * @returns {boolean} True if valid
   */
  validateFilterState(filterState) {
    if (!filterState || typeof filterState !== "object") return false;

    // Check required properties
    const requiredProps = ["filters", "columns", "sorting"];
    return requiredProps.every((prop) => filterState.hasOwnProperty(prop));
  }

  /**
   * Generate a human-readable description of the filter
   * @param {Object} filterState - Filter state object
   * @returns {string} Filter description
   */
  generateFilterDescription(filterState) {
    const parts = [];

    if (filterState.activeFilterCount > 0) {
      parts.push(
        `${filterState.activeFilterCount} filter${filterState.activeFilterCount !== 1 ? "s" : ""}`,
      );
    }

    if (filterState.sorting && filterState.sorting.length > 0) {
      parts.push(
        `${filterState.sorting.length} sort${filterState.sorting.length !== 1 ? "s" : ""}`,
      );
    }

    if (filterState.columns && filterState.columns.length > 0) {
      parts.push("custom column layout");
    }

    return parts.length > 0 ? parts.join(", ") : "no filters";
  }

  /**
   * Clear all saved filters (with confirmation)
   */
  async clearAllSavedFilters() {
    try {
      const savedFilters = await this.getSavedFilters();
      const filterCount = Object.keys(savedFilters).length;

      if (filterCount === 0) {
        this.showToast("No saved filters to clear.", "info");
        return;
      }

      const confirmClear = confirm(
        `Are you sure you want to delete all ${filterCount} saved filter${filterCount !== 1 ? "s" : ""}? This action cannot be undone.`,
      );

      if (!confirmClear) return;

      // Delete each filter individually since we don't have a bulk delete endpoint
      const deletePromises = Object.keys(savedFilters).map((name) =>
        fetch(`/api/filters/by-name/${encodeURIComponent(name)}`, {
          method: "DELETE",
          credentials: "include",
        }),
      );

      await Promise.all(deletePromises);
      await this.loadSavedFiltersToDropdown();

      // Refresh management panel if it's open
      if (
        document.getElementById("filter-management-panel") &&
        !document
          .getElementById("filter-management-panel")
          .classList.contains("hidden")
      ) {
        this.renderFilterList();
      }

      // Reset dropdown selection
      const dropdown = document.getElementById("savedFiltersDropdown");
      if (dropdown) {
        dropdown.value = "";
      }

      this.showToast(`All ${filterCount} saved filters cleared!`, "success");
    } catch (error) {
      console.error("Error clearing all saved filters:", error);
      this.showToast(
        "Error clearing saved filters. Please try again.",
        "error",
      );
    }
  }

  /**
   * Duplicate a filter with a new name
   * @param {string} originalName - Original filter name
   * @param {string} newName - New filter name
   */
  async duplicateFilter(originalName, newName) {
    try {
      const response = await fetch(
        `/api/filters/by-name/${encodeURIComponent(originalName)}`,
        {
          method: "GET",
          credentials: "include",
        },
      );

      if (!response.ok) {
        throw new Error("Failed to load original filter");
      }

      const result = await response.json();
      await this.saveFilter(newName, result.filter_data);

      this.showToast(
        `Filter duplicated as "${newName}" successfully!`,
        "success",
      );
    } catch (error) {
      console.error("Error duplicating filter:", error);
      this.showToast("Error duplicating filter: " + error.message, "error");
    }
  }

  /**
   * Show toast notification
   * @param {string} message - Message to show
   * @param {string} type - Toast type (success, error, warning, info)
   */
  showToast(message, type = "info") {
    // Create toast element
    const toast = document.createElement("div");
    toast.className = `filter-toast filter-toast-${type}`;
    toast.textContent = message;

    // Style the toast
    toast.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      padding: 12px 20px;
      border-radius: 4px;
      color: white;
      font-weight: bold;
      z-index: 10000;
      opacity: 0;
      transition: opacity 0.3s ease;
      max-width: 400px;
      word-wrap: break-word;
    `;

    // Set colors based on type
    const colors = {
      success: "#28a745",
      error: "#dc3545",
      warning: "#ffc107",
      info: "#17a2b8",
    };
    toast.style.backgroundColor = colors[type] || colors.info;

    if (type === "warning") {
      toast.style.color = "#212529"; // Dark text for warning
    }

    // Add to page
    document.body.appendChild(toast);

    // Animate in
    setTimeout(() => (toast.style.opacity = "1"), 100);

    // Remove after delay
    setTimeout(() => {
      toast.style.opacity = "0";
      setTimeout(() => {
        if (document.body.contains(toast)) {
          document.body.removeChild(toast);
        }
      }, 300);
    }, 4000);
  }

  /**
   * Get filter statistics
   * @returns {Promise<Object>} Filter statistics
   */
  async getFilterStats() {
    try {
      const response = await fetch("/api/filters/stats", {
        method: "GET",
        credentials: "include",
      });

      if (!response.ok) {
        throw new Error("Failed to fetch filter stats");
      }

      return await response.json();
    } catch (error) {
      console.error("Error getting filter stats:", error);
      return { total_filters: 0 };
    }
  }

  /**
   * Show filter management panel
   */
  showFilterManagementPanel() {
    // Use the existing modal function if available
    if (typeof window.openFilterManagementPanel === "function") {
      window.openFilterManagementPanel();
    } else {
      // Fallback implementation
      const panel = document.getElementById("filter-management-panel");
      const panelContent = document.getElementById("filter-panel-content");

      if (panel && panelContent) {
        // Show the panel
        panel.classList.remove("hidden");

        // Trigger animation
        setTimeout(() => {
          panel.classList.remove("opacity-0");
          panelContent.classList.remove("scale-95", "opacity-0");
          panelContent.classList.add("scale-100", "opacity-100");
        }, 10);

        // Load current filters into the management panel
        this.renderFilterList();
      } else {
        this.showToast("Filter management panel not found!", "error");
      }
    }
  }

  /**
   * Setup management panel events
   */
  setupManagementPanelEvents() {
    // Placeholder for management panel event setup
  }

  /**
   * Render filter list in management panel
   */
  async renderFilterList() {
    const filterList = document.getElementById("filter-list");
    if (!filterList) return;

    try {
      const savedFilters = await this.getSavedFilters();
      const filterNames = Object.keys(savedFilters).sort();

      if (filterNames.length === 0) {
        filterList.innerHTML =
          '<p class="text-gray-500 dark:text-gray-400">No saved filters</p>';
        return;
      }

      const filterItems = filterNames
        .map((name) => {
          const filter = savedFilters[name];
          const savedDate = new Date(filter.savedAt).toLocaleDateString();

          return `
          <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
            <div class="flex-1">
              <h4 class="font-medium text-gray-900 dark:text-white">${name}</h4>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                ${filter.activeFilterCount || 0} filters • Saved ${savedDate}
              </p>
            </div>
            <div class="flex space-x-2">
              <button onclick="filterManager.loadFilter('${name}').then(() => filterManager.renderFilterList())" class="text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"></path>
                </svg>
              </button>
              <button onclick="filterManager.deleteFilter('${name}').then(() => filterManager.renderFilterList())" class="text-red-600 hover:text-red-800 dark:text-red-400 dark:hover:text-red-300">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                </svg>
              </button>
            </div>
          </div>
        `;
        })
        .join("");

      filterList.innerHTML = filterItems;
    } catch (error) {
      console.error("Error rendering filter list:", error);
      filterList.innerHTML =
        '<p class="text-red-500">Error loading filters</p>';
    }
  }

  /**
   * Handle export filters button click
   */
  async handleExportFilters() {
    try {
      const savedFilters = await this.getSavedFilters();
      const filterCount = Object.keys(savedFilters).length;

      if (filterCount === 0) {
        this.showToast("No saved filters to export.", "info");
        return;
      }

      const exportData = {
        version: "1.0",
        exportDate: new Date().toISOString(),
        filterCount: filterCount,
        filters: savedFilters,
      };

      const dataStr = JSON.stringify(exportData, null, 2);
      const dataBlob = new Blob([dataStr], { type: "application/json" });

      const link = document.createElement("a");
      link.href = URL.createObjectURL(dataBlob);
      link.download = `ag-grid-filters-${new Date().toISOString().split("T")[0]}.json`;

      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);

      URL.revokeObjectURL(link.href);

      this.showToast(
        `${filterCount} filter${filterCount !== 1 ? "s" : ""} exported successfully!`,
        "success",
      );
    } catch (error) {
      console.error("Error exporting filters:", error);
      this.showToast("Error exporting filters. Please try again.", "error");
    }
  }

  /**
   * Handle import filters from file input
   * @param {Event} event - File input change event
   */
  async handleImportFilters(event) {
    await this.importFilters(event);
  }
}

// For backward compatibility, alias to the window object
window.FilterManager = FilterManager;

// ES6 module export
export { FilterManager };
