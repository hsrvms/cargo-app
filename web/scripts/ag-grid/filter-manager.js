/**
 * Filter Manager for AG Grid - Handles saving, loading, and managing filter states
 */
class FilterManager {
  constructor(gridApi) {
    this.gridApi = gridApi;
    this.storageKey = "ag-grid-saved-filters";

    // Debug: Log available grid API methods
    console.log("FilterManager: Grid API object:", gridApi);
    console.log(
      "FilterManager: Available methods on gridApi:",
      Object.getOwnPropertyNames(gridApi),
    );
    console.log(
      "FilterManager: Grid API prototype methods:",
      Object.getOwnPropertyNames(Object.getPrototypeOf(gridApi)),
    );

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
      console.log(`FilterManager: ${method} available: ${available}`);
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
   * Setup event listeners for filter-related buttons and dropdowns
   */
  setupEventListeners() {
    // Save filter button
    const saveFilterBtn = document.getElementById("saveFilterBtn");
    if (saveFilterBtn) {
      saveFilterBtn.addEventListener("click", () => this.handleSaveFilter());
    }

    // Clear filters button
    const clearFiltersBtn = document.getElementById("clearFiltersBtn");
    if (clearFiltersBtn) {
      clearFiltersBtn.addEventListener("click", () => this.clearAllFilters());
    }

    // Manage filters button
    const manageFiltersBtn = document.getElementById("manageFiltersBtn");
    if (manageFiltersBtn) {
      manageFiltersBtn.addEventListener("click", () =>
        this.showFilterManagementPanel(),
      );
    }

    // Saved filters dropdown
    const savedFiltersDropdown = document.getElementById(
      "savedFiltersDropdown",
    );
    if (savedFiltersDropdown) {
      savedFiltersDropdown.addEventListener("change", (e) => {
        if (e.target.value) {
          this.loadFilter(e.target.value);
        }
      });
    }
  }

  /**
   * Get current complete grid state including filters, column order, and sorting
   * @returns {Object} Complete grid state with filters, columns, and sorting
   */
  getCurrentFilterState() {
    if (!this.gridApi) {
      console.error("Grid API not available");
      return {};
    }

    try {
      // Safely get filter model
      const filterModel =
        typeof this.gridApi.getFilterModel === "function"
          ? this.gridApi.getFilterModel()
          : {};

      // Safely get column state
      const columnState =
        typeof this.gridApi.getColumnState === "function"
          ? this.gridApi.getColumnState()
          : [];

      // Get current sorting state (may not be available in all AG Grid versions)
      const sortModel =
        typeof this.gridApi.getSortModel === "function"
          ? this.gridApi.getSortModel()
          : [];

      // Count active filters
      const activeFilters = Object.keys(filterModel || {}).length;

      // Count columns that have been reordered from default
      const reorderedColumns = columnState.filter(
        (col, index) => col.colId !== columnState[0].colId || col.sort !== null,
      ).length;

      return {
        filters: filterModel,
        columns: columnState,
        sorting: sortModel,
        timestamp: new Date().toISOString(),
        activeFilterCount: activeFilters,
        columnChanges: reorderedColumns,
        description: this.generateFilterDescription(
          filterModel,
          columnState,
          sortModel,
        ),
      };
    } catch (error) {
      console.error("Error getting current filter state:", error);
      return {};
    }
  }

  /**
   * Save current filter state with a name
   */
  handleSaveFilter() {
    const filterName = this.promptForFilterName();
    if (!filterName) return;

    const currentState = this.getCurrentFilterState();
    const hasFilters = Object.keys(currentState.filters || {}).length > 0;
    const hasColumnChanges = (currentState.columns || []).some(
      (col) => col.sort !== null,
    );
    const hasSorting = (currentState.sorting || []).length > 0;

    if (!hasFilters && !hasColumnChanges && !hasSorting) {
      alert(
        "No filters, column changes, or sorting are currently applied. Please configure the grid before saving.",
      );
      return;
    }

    this.saveFilter(filterName, currentState);
  }

  /**
   * Prompt user for filter name
   * @returns {string|null} Filter name or null if cancelled
   */
  promptForFilterName() {
    let filterName = prompt("Enter a name for this filter:");

    if (!filterName) return null;

    filterName = filterName.trim();
    if (!filterName) {
      this.showToast("Please enter a valid filter name.", "error");
      return null;
    }

    // Validate filter name length and characters
    if (filterName.length > 50) {
      this.showToast("Filter name must be 50 characters or less.", "error");
      return null;
    }

    // Check for invalid characters
    const invalidChars = /[<>:"/\\|?*]/g;
    if (invalidChars.test(filterName)) {
      this.showToast(
        "Filter name contains invalid characters. Please use letters, numbers, spaces, and basic punctuation only.",
        "error",
      );
      return null;
    }

    // Check if name already exists
    const savedFilters = this.getSavedFilters();
    if (savedFilters[filterName]) {
      const overwrite = confirm(
        `A filter named "${filterName}" already exists. Do you want to overwrite it?`,
      );
      if (!overwrite) return null;
    }

    return filterName;
  }

  /**
   * Save filter to localStorage
   * @param {string} name - Filter name
   * @param {Object} filterState - Filter state object
   */
  saveFilter(name, filterState) {
    try {
      // Validate filter state before saving
      if (!this.validateFilterState(filterState)) {
        this.showToast("Invalid filter state. Cannot save filter.", "error");
        return;
      }

      const savedFilters = this.getSavedFilters();
      savedFilters[name] = {
        ...filterState,
        savedAt: new Date().toISOString(),
        name: name,
        version: "1.0",
        userAgent: navigator.userAgent.split(" ")[0], // Simple browser detection
      };

      localStorage.setItem(this.storageKey, JSON.stringify(savedFilters));
      this.loadSavedFiltersToDropdown();

      // Show detailed success message
      const filterCount = filterState.activeFilterCount || 0;
      const hasColumns = (filterState.columns || []).length > 0;
      const hasSorting = (filterState.sorting || []).length > 0;

      let message = `"${name}" saved with ${filterCount} filter${filterCount !== 1 ? "s" : ""}`;
      if (hasColumns) message += ", column layout";
      if (hasSorting) message += ", sorting";
      message += "!";

      this.showToast(message, "success");

      console.log("Filter saved:", name, filterState);
    } catch (error) {
      console.error("Error saving filter:", error);
      if (error.name === "QuotaExceededError") {
        this.showToast(
          "Storage quota exceeded. Please delete some old filters and try again.",
          "error",
        );
      } else {
        this.showToast("Error saving filter. Please try again.", "error");
      }
    }
  }

  /**
   * Load filter and apply it to the grid
   * @param {string} name - Filter name to load
   */
  loadFilter(name) {
    try {
      const savedFilters = this.getSavedFilters();
      const filterState = savedFilters[name];

      if (!filterState) {
        console.error("Filter not found:", name);
        this.showToast(`Filter "${name}" not found.`, "error");
        return;
      }

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
      console.log("Filter loaded:", name, filterState);
    } catch (error) {
      console.error("Error loading filter:", error);
      this.showToast("Error loading filter. Please try again.", "error");
    }
  }

  /**
   * Delete a saved filter
   * @param {string} name - Filter name to delete
   */
  deleteFilter(name) {
    try {
      const savedFilters = this.getSavedFilters();

      if (!savedFilters[name]) {
        console.error("Filter not found:", name);
        return;
      }

      const confirmDelete = confirm(
        `Are you sure you want to delete the filter "${name}"?`,
      );
      if (!confirmDelete) return;

      delete savedFilters[name];
      localStorage.setItem(this.storageKey, JSON.stringify(savedFilters));
      this.loadSavedFiltersToDropdown();

      // Reset dropdown selection
      const dropdown = document.getElementById("savedFiltersDropdown");
      if (dropdown) {
        dropdown.value = "";
      }

      this.showToast(`Filter "${name}" deleted successfully!`, "success");
      console.log("Filter deleted:", name);
    } catch (error) {
      console.error("Error deleting filter:", error);
      this.showToast("Error deleting filter. Please try again.", "error");
    }
  }

  /**
   * Get all saved filters from localStorage
   * @returns {Object} Saved filters object
   */
  getSavedFilters() {
    try {
      const saved = localStorage.getItem(this.storageKey);
      return saved ? JSON.parse(saved) : {};
    } catch (error) {
      console.error("Error reading saved filters:", error);
      return {};
    }
  }

  /**
   * Load saved filters into the dropdown
   */
  loadSavedFiltersToDropdown() {
    const dropdown = document.getElementById("savedFiltersDropdown");
    if (!dropdown) return;

    const savedFilters = this.getSavedFilters();
    const filterNames = Object.keys(savedFilters).sort();

    // Clear existing options except the first one
    dropdown.innerHTML = '<option value="">Select Saved Filter...</option>';

    // Add saved filters with descriptions
    filterNames.forEach((name) => {
      const filterData = savedFilters[name];
      const option = document.createElement("option");
      option.value = name;

      // Create descriptive text
      const filterCount = filterData.activeFilterCount || 0;
      const savedDate = new Date(filterData.savedAt).toLocaleDateString();
      const hasColumns = (filterData.columns || []).length > 0;
      const hasSorting = (filterData.sorting || []).length > 0;

      let description = `${filterCount} filter${filterCount !== 1 ? "s" : ""}`;
      if (hasColumns || hasSorting) {
        const extras = [];
        if (hasColumns) extras.push("columns");
        if (hasSorting) extras.push("sorting");
        description += ` + ${extras.join(" + ")}`;
      }

      option.textContent = `${name} (${description} - ${savedDate})`;

      // Add tooltip with description if available
      if (filterData.description) {
        option.title = filterData.description;
      }

      dropdown.appendChild(option);
    });

    // Add delete options if there are saved filters
    if (filterNames.length > 0) {
      const separator = document.createElement("option");
      separator.disabled = true;
      separator.textContent = "──────────────";
      dropdown.appendChild(separator);

      filterNames.forEach((name) => {
        const deleteOption = document.createElement("option");
        deleteOption.value = `delete:${name}`;
        deleteOption.textContent = `🗑️ Delete: ${name}`;
        deleteOption.style.color = "#dc2626";
        dropdown.appendChild(deleteOption);
      });
    }

    // Update dropdown change handler to handle delete options
    dropdown.removeEventListener("change", this.dropdownChangeHandler);
    this.dropdownChangeHandler = (e) => {
      const value = e.target.value;
      if (value.startsWith("delete:")) {
        const filterName = value.substring(7);
        this.deleteFilter(filterName);
      } else if (value) {
        this.loadFilter(value);
      }
    };
    dropdown.addEventListener("change", this.dropdownChangeHandler);

    console.log("Loaded filters to dropdown:", filterNames);
  }

  /**
   * Clear all filters from the grid
   */
  clearAllFilters() {
    try {
      if (typeof this.gridApi.setFilterModel === "function") {
        this.gridApi.setFilterModel({});
        this.showToast("All filters cleared!", "success");
      } else {
        console.warn("setFilterModel method not available");
        this.showToast(
          "Clear filters not supported in this grid version.",
          "error",
        );
      }

      // Reset dropdown selection
      const dropdown = document.getElementById("savedFiltersDropdown");
      if (dropdown) {
        dropdown.value = "";
      }
    } catch (error) {
      console.error("Error clearing filters:", error);
      this.showToast("Error clearing filters.", "error");
    }
  }

  /**
   * Export saved filters as JSON
   * @returns {string} JSON string of saved filters
   */
  exportFilters() {
    const savedFilters = this.getSavedFilters();
    return JSON.stringify(savedFilters, null, 2);
  }

  /**
   * Import filters from JSON string
   * @param {string} jsonString - JSON string of filters to import
   */
  importFilters(jsonString) {
    try {
      const importedFilters = JSON.parse(jsonString);
      const currentFilters = this.getSavedFilters();

      // Merge imported filters with existing ones
      const mergedFilters = { ...currentFilters, ...importedFilters };

      localStorage.setItem(this.storageKey, JSON.stringify(mergedFilters));
      this.loadSavedFiltersToDropdown();

      const importedCount = Object.keys(importedFilters).length;
      this.showToast(
        `${importedCount} filters imported successfully!`,
        "success",
      );
    } catch (error) {
      console.error("Error importing filters:", error);
      this.showToast(
        "Error importing filters. Please check the JSON format.",
        "error",
      );
    }
  }

  /**
   * Validate filter state before saving
   * @param {Object} filterState - Filter state to validate
   * @returns {boolean} Whether the filter state is valid
   */
  validateFilterState(filterState) {
    if (!filterState || typeof filterState !== "object") {
      return false;
    }

    // Check if filters object exists and is valid
    if (!filterState.filters || typeof filterState.filters !== "object") {
      return false;
    }

    // Check if timestamp is valid
    if (!filterState.timestamp || isNaN(new Date(filterState.timestamp))) {
      return false;
    }

    return true;
  }

  /**
   * Generate human-readable description of complete grid state
   * @param {Object} filterModel - AG Grid filter model
   * @param {Array} columnState - AG Grid column state
   * @param {Array} sortModel - AG Grid sort model
   * @returns {string} Complete state description
   */
  generateFilterDescription(filterModel, columnState = [], sortModel = []) {
    const parts = [];

    // Describe filters
    if (filterModel && Object.keys(filterModel).length > 0) {
      const filterDescriptions = [];
      Object.entries(filterModel).forEach(([column, filter]) => {
        if (filter.filterType === "text") {
          filterDescriptions.push(`${column}: "${filter.filter}"`);
        } else if (filter.filterType === "set") {
          const values = filter.values || [];
          filterDescriptions.push(`${column}: ${values.length} selected`);
        } else if (filter.filterType === "number") {
          filterDescriptions.push(`${column}: ${filter.type} ${filter.filter}`);
        } else {
          filterDescriptions.push(`${column}: filtered`);
        }
      });
      parts.push(`Filters: ${filterDescriptions.join(", ")}`);
    }

    // Describe sorting
    if (sortModel && sortModel.length > 0) {
      const sortDescriptions = sortModel.map(
        (sort) => `${sort.colId} (${sort.sort})`,
      );
      parts.push(`Sorted by: ${sortDescriptions.join(", ")}`);
    }

    // Describe column changes (simplified)
    if (columnState && columnState.length > 0) {
      const hiddenCols = columnState.filter((col) => !col.hide).length;
      if (hiddenCols < columnState.length) {
        parts.push(
          `${columnState.length - hiddenCols} of ${columnState.length} columns visible`,
        );
      }
    }

    return parts.length > 0 ? parts.join(" | ") : "No active configuration";
  }

  /**
   * Clear all saved filters (with confirmation)
   */
  clearAllSavedFilters() {
    const savedFilters = this.getSavedFilters();
    const filterCount = Object.keys(savedFilters).length;

    if (filterCount === 0) {
      this.showToast("No saved filters to clear.", "info");
      return;
    }

    const confirmClear = confirm(
      `Are you sure you want to delete all ${filterCount} saved filter${filterCount !== 1 ? "s" : ""}? This action cannot be undone.`,
    );

    if (!confirmClear) return;

    try {
      localStorage.removeItem(this.storageKey);
      this.loadSavedFiltersToDropdown();

      // Reset dropdown selection
      const dropdown = document.getElementById("savedFiltersDropdown");
      if (dropdown) {
        dropdown.value = "";
      }

      this.showToast(`All ${filterCount} saved filters cleared!`, "success");
      console.log("All saved filters cleared");
    } catch (error) {
      console.error("Error clearing all saved filters:", error);
      this.showToast(
        "Error clearing saved filters. Please try again.",
        "error",
      );
    }
  }

  /**
   * Duplicate an existing saved filter with a new name
   * @param {string} originalName - Name of filter to duplicate
   */
  duplicateFilter(originalName) {
    const savedFilters = this.getSavedFilters();
    const originalFilter = savedFilters[originalName];

    if (!originalFilter) {
      this.showToast(`Filter "${originalName}" not found.`, "error");
      return;
    }

    const newName = prompt(
      `Enter name for duplicated filter:`,
      `${originalName} (Copy)`,
    );
    if (!newName || newName.trim() === "") return;

    const trimmedName = newName.trim();
    if (savedFilters[trimmedName]) {
      this.showToast(
        `A filter named "${trimmedName}" already exists.`,
        "error",
      );
      return;
    }

    try {
      savedFilters[trimmedName] = {
        ...originalFilter,
        name: trimmedName,
        savedAt: new Date().toISOString(),
      };

      localStorage.setItem(this.storageKey, JSON.stringify(savedFilters));
      this.loadSavedFiltersToDropdown();

      this.showToast(
        `Filter "${originalName}" duplicated as "${trimmedName}"!`,
        "success",
      );
      console.log("Filter duplicated:", originalName, "->", trimmedName);
    } catch (error) {
      console.error("Error duplicating filter:", error);
      this.showToast("Error duplicating filter. Please try again.", "error");
    }
  }

  /**
   * Show toast notification
   * @param {string} message - Message to show
   * @param {string} type - Toast type ('success', 'error', 'info')
   */
  showToast(message, type = "info") {
    // Create toast element
    const toast = document.createElement("div");
    toast.className = `fixed top-4 right-4 z-50 p-4 rounded-lg shadow-lg transition-all duration-300 transform translate-x-full`;

    // Set colors based on type
    const typeClasses = {
      success: "bg-green-500 text-white",
      error: "bg-red-500 text-white",
      info: "bg-blue-500 text-white",
    };

    toast.className += ` ${typeClasses[type] || typeClasses.info}`;
    toast.textContent = message;

    // Add to DOM
    document.body.appendChild(toast);

    // Animate in
    setTimeout(() => {
      toast.classList.remove("translate-x-full");
    }, 100);

    // Remove after 3 seconds
    setTimeout(() => {
      toast.classList.add("translate-x-full");
      setTimeout(() => {
        if (toast.parentNode) {
          toast.parentNode.removeChild(toast);
        }
      }, 300);
    }, 3000);
  }

  /**
   * Get filter statistics
   * @returns {Object} Statistics about saved filters
   */
  getFilterStats() {
    const savedFilters = this.getSavedFilters();
    const filterNames = Object.keys(savedFilters);

    return {
      totalFilters: filterNames.length,
      filterNames: filterNames,
      lastSaved:
        filterNames.length > 0
          ? Math.max(
              ...Object.values(savedFilters).map((f) =>
                new Date(f.savedAt).getTime(),
              ),
            )
          : null,
    };
  }

  /**
   * Show the filter management panel
   */
  showFilterManagementPanel() {
    // Setup panel events if not already done
    this.setupManagementPanelEvents();

    // Use global function to open panel (handles animation)
    if (typeof window.openFilterManagementPanel === "function") {
      window.openFilterManagementPanel();
    } else {
      console.error("Global openFilterManagementPanel function not available");
    }
  }

  /**
   * Setup event listeners for the management panel
   */
  setupManagementPanelEvents() {
    // Export filters button
    const exportBtn = document.getElementById("exportFiltersBtn");
    if (exportBtn && !exportBtn.hasAttribute("data-listener-added")) {
      exportBtn.addEventListener("click", () => this.handleExportFilters());
      exportBtn.setAttribute("data-listener-added", "true");
    }

    // Import filters input
    const importInput = document.getElementById("importFiltersInput");
    if (importInput && !importInput.hasAttribute("data-listener-added")) {
      importInput.addEventListener("change", (e) =>
        this.handleImportFilters(e),
      );
      importInput.setAttribute("data-listener-added", "true");
    }

    // Clear all saved filters button
    const clearAllBtn = document.getElementById("clearAllFiltersBtn");
    if (clearAllBtn && !clearAllBtn.hasAttribute("data-listener-added")) {
      clearAllBtn.addEventListener("click", () => this.clearAllSavedFilters());
      clearAllBtn.setAttribute("data-listener-added", "true");
    }
  }

  /**
   * Render the list of saved filters in the management panel
   */
  renderFilterList() {
    const filterList = document.getElementById("filter-list");
    if (!filterList) return;

    const savedFilters = this.getSavedFilters();
    const filterNames = Object.keys(savedFilters).sort();

    if (filterNames.length === 0) {
      filterList.innerHTML = `
        <div class="text-center py-8 text-gray-500 dark:text-gray-400">
          <svg class="w-12 h-12 mx-auto mb-4 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"></path>
          </svg>
          <p class="text-lg font-medium">No saved filters</p>
          <p class="text-sm">Apply some filters to the grid and save them to get started.</p>
        </div>
      `;
      return;
    }

    const filterItems = filterNames
      .map((name) => {
        const filter = savedFilters[name];
        const date = new Date(filter.savedAt).toLocaleDateString();
        const time = new Date(filter.savedAt).toLocaleTimeString();
        const filterCount = filter.activeFilterCount || 0;

        return `
        <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
          <div class="flex items-center justify-between">
            <div class="flex-1">
              <h4 class="font-medium text-gray-900 dark:text-white">${name}</h4>
              <p class="text-sm text-gray-500 dark:text-gray-400">
                ${filterCount} filter${filterCount !== 1 ? "s" : ""} • Saved ${date} at ${time}
              </p>
              ${filter.description ? `<p class="text-xs text-gray-400 dark:text-gray-500 mt-1" title="${filter.description}">${filter.description.length > 80 ? filter.description.substring(0, 77) + "..." : filter.description}</p>` : ""}
            </div>
            <div class="flex space-x-2 ml-4">
              <button
                onclick="window.filterManager?.loadFilter('${name}')"
                class="px-3 py-1 text-xs bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 rounded hover:bg-blue-200 dark:hover:bg-blue-800"
                title="Apply this filter"
              >
                Apply
              </button>
              <button
                onclick="window.filterManager?.duplicateFilter('${name}')"
                class="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
                title="Duplicate this filter"
              >
                Copy
              </button>
              <button
                onclick="window.filterManager?.deleteFilter('${name}')"
                class="px-3 py-1 text-xs bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-300 rounded hover:bg-red-200 dark:hover:bg-red-800"
                title="Delete this filter"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      `;
      })
      .join("");

    filterList.innerHTML = filterItems;
  }

  /**
   * Handle export filters button click
   */
  handleExportFilters() {
    try {
      const savedFilters = this.getSavedFilters();
      const filterCount = Object.keys(savedFilters).length;

      if (filterCount === 0) {
        this.showToast("No saved filters to export.", "info");
        return;
      }

      const exportData = {
        filters: savedFilters,
        exportDate: new Date().toISOString(),
        version: "1.0",
        totalFilters: filterCount,
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
      console.log("Filters exported:", exportData);
    } catch (error) {
      console.error("Error exporting filters:", error);
      this.showToast("Error exporting filters. Please try again.", "error");
    }
  }

  /**
   * Handle import filters from file input
   * @param {Event} event - File input change event
   */
  handleImportFilters(event) {
    const file = event.target.files[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
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

        // Merge with existing filters
        const currentFilters = this.getSavedFilters();
        const mergedFilters = { ...currentFilters };
        let newCount = 0;
        let overwriteCount = 0;

        Object.entries(importedFilters).forEach(([name, filter]) => {
          if (currentFilters[name]) {
            overwriteCount++;
          } else {
            newCount++;
          }
          mergedFilters[name] = {
            ...filter,
            importedAt: new Date().toISOString(),
          };
        });

        localStorage.setItem(this.storageKey, JSON.stringify(mergedFilters));
        this.loadSavedFiltersToDropdown();
        this.renderFilterList();

        let message = `Imported ${importedCount} filter${importedCount !== 1 ? "s" : ""}`;
        if (newCount > 0 && overwriteCount > 0) {
          message += ` (${newCount} new, ${overwriteCount} overwritten)`;
        } else if (overwriteCount > 0) {
          message += ` (${overwriteCount} overwritten)`;
        }

        this.showToast(`${message}!`, "success");
        console.log("Filters imported:", {
          newCount,
          overwriteCount,
          importData,
        });
      } catch (error) {
        console.error("Error importing filters:", error);
        this.showToast(
          "Error importing filters. Please check the file format.",
          "error",
        );
      } finally {
        // Clear the file input
        event.target.value = "";
      }
    };

    reader.onerror = () => {
      this.showToast("Error reading file. Please try again.", "error");
      event.target.value = "";
    };

    reader.readAsText(file);
  }
}

// Export for use in other modules
export { FilterManager };
