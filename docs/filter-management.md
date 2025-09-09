# AG Grid State Management System

## Overview

The Grid State Management System provides comprehensive functionality for saving, loading, and managing complete grid states in the AG Grid shipments view. Users can save frequently used filter combinations, column layouts, sorting configurations, share them with others, and quickly switch between different grid configurations.

## Features

### 1. Save Current Grid State
- **Button**: "Save Filter" (blue button with save icon)
- **Description**: Saves the current grid state including filters, column order, column widths, sorting, and visibility settings
- **Usage**: 
  1. Configure the grid as desired (apply filters, reorder columns, set sorting, etc.)
  2. Click "Save Filter"
  3. Enter a descriptive name when prompted
  4. Complete grid state is saved to local storage

### 2. Load Saved Grid States
- **Control**: "Select Saved Filter..." dropdown
- **Description**: Quickly apply previously saved grid configurations including filters, column layout, and sorting
- **Usage**: 
  1. Click the dropdown to see available saved configurations
  2. Select a configuration to apply it immediately
  3. The grid will update with the selected complete state (filters, columns, sorting)

### 3. Clear All Filters
- **Button**: "Clear Filters" (gray button with cross icon)
- **Description**: Removes all active filters from the grid (does not affect column order or sorting)
- **Usage**: Click the button to clear filters while maintaining column layout and sorting

### 4. Grid State Management Panel
- **Button**: "Manage" (purple button with gear icon)
- **Description**: Advanced management interface for saved grid configurations
- **Features**:
  - View all saved configurations with details (filters, columns, sorting)
  - Apply, duplicate, or delete individual configurations
  - Import/export configuration collections
  - Clear all saved configurations

## Filter Management Panel

### Configuration List
Each saved configuration displays:
- **Name**: User-defined configuration name
- **State Summary**: Number of active filters, column changes, and sorting
- **Save Date**: When the configuration was saved
- **Description**: Auto-generated description of filters, sorting, and column changes
- **Actions**:
  - **Apply**: Load and apply the complete configuration
  - **Copy**: Duplicate the configuration with a new name
  - **Delete**: Remove the configuration (with confirmation)

### Import/Export Functionality

#### Export Configurations
- **Button**: "Export Filters"
- **Function**: Downloads all saved grid configurations as a JSON file
- **File Format**: `ag-grid-filters-YYYY-MM-DD.json`
- **Use Cases**:
  - Backup complete grid configurations
  - Share grid setups with team members
  - Transfer configurations between environments

#### Import Configurations
- **Button**: "Import Filters" (file upload)
- **Function**: Loads grid configurations from a JSON file
- **Supported Format**: JSON files exported by this system
- **Behavior**: 
  - Merges imported configurations with existing ones
  - Overwrites configurations with identical names (with confirmation)
  - Shows summary of imported configurations

### Danger Zone
- **Clear All Saved Filters**: Permanently deletes all saved configurations (with confirmation)

## Technical Details

### Storage
- **Method**: Browser localStorage
- **Key**: `ag-grid-saved-filters`
- **Format**: JSON object with configuration names as keys

### Grid State Components
Each saved configuration includes:
- **Filters**: AG Grid filter model (all column filters)
- **Columns**: Complete column state (width, order, visibility, pinning)
- **Sorting**: Sort model (multi-column sorting state)
- **Metadata**: Save date, filter count, column changes, description, version

### Browser Compatibility
- Requires localStorage support
- Compatible with modern browsers (Chrome 4+, Firefox 3.5+, Safari 4+, IE 8+)

## Usage Examples

### Example 1: Save "In Transit Dashboard" Configuration
1. Configure the grid:
   - Filters: Status: "In Transit", Next ETA: Last 7 days
   - Column order: Move "Status" and "Next ETA" to front
   - Sorting: Sort by "Next ETA" ascending
2. Click "Save Filter"
3. Enter name: "In Transit Dashboard"
4. Complete configuration is saved and appears in dropdown

### Example 2: Team Configuration Sharing
1. Team member creates useful grid configurations (filters, columns, sorting)
2. Exports configurations using "Export Filters" button
3. Shares the JSON file with team
4. Other team members import using "Import Filters"
5. Everyone has access to the same grid layouts and filter presets

### Example 3: Role-Based Grid Views
1. Save role-specific grid configurations:
   - "Manager View" (All columns, sorted by priority, filtered for exceptions)
   - "Operator View" (Essential columns only, sorted by ETA)
   - "Customer Service" (Customer-facing columns, filtered by active shipments)
2. Use dropdown to quickly switch between complete role-based layouts

## Keyboard Shortcuts

Currently not implemented, but could be added:
- `Ctrl+S`: Save current grid configuration
- `Ctrl+Shift+S`: Open configuration management panel
- `Escape`: Close configuration management panel

## Troubleshooting

### Common Issues

#### Configuration Not Saving
- **Cause**: No changes made to grid (no filters, column changes, or sorting)
- **Solution**: Apply filters, reorder columns, or set sorting before saving
- **Error Message**: "No filters, column changes, or sorting are currently applied. Please configure the grid before saving."

#### Storage Quota Exceeded
- **Cause**: Too many saved filters (localStorage limit ~5-10MB)
- **Solution**: Delete old filters or export and clear storage
- **Error Message**: "Storage quota exceeded. Please delete some old filters and try again."

#### Filter Not Loading
- **Cause**: Corrupted filter data or incompatible format
- **Solution**: Delete the problematic filter and recreate
- **Prevention**: Use export/import for backup

#### Import Fails
- **Cause**: Invalid JSON format or incompatible file structure
- **Solution**: Ensure file was exported from this system
- **Error Message**: "Error importing filters. Please check the file format."

### Data Recovery
If filters are lost:
1. Check browser localStorage for `ag-grid-saved-filters`
2. Use browser developer tools to inspect storage
3. Restore from exported JSON backup if available

### Performance Considerations
- Large numbers of saved filters (>50) may slow down dropdown loading
- Complex filter states may take longer to apply
- Consider periodic cleanup of unused filters

## Advanced Features

### Filter Validation
- Validates filter structure before saving
- Checks for required fields (timestamp, filters object)
- Prevents saving of empty or invalid filter states

### Auto-Generated Descriptions
- Creates human-readable descriptions of filter conditions
- Shows column names and filter types
- Helps identify filters quickly

### Version Tracking
- Each filter includes version information
- Enables future compatibility and migration
- Tracks save/import timestamps

## Best Practices

### Naming Conventions
- Use descriptive names: "High Priority Shipments" vs "Filter1"
- Include date ranges: "Q1 2024 Exports"
- Specify criteria: "US Domestic - In Transit"

### Organization
- Regularly review and clean up unused filters
- Export filters periodically for backup
- Share common filters with team members

### Performance
- Limit saved filters to essential combinations
- Delete outdated filters regularly
- Use clear filters button to reset before applying new criteria

## Future Enhancements

Potential improvements:
- Filter categories/folders
- User permissions and sharing
- Server-side filter storage
- Filter templates
- Scheduled filter application
- Integration with user profiles
- Real-time filter collaboration

## API Reference

### FilterManager Class Methods

```javascript
// Initialize filter manager
const filterManager = new FilterManager(gridApi);

// Save current filter state
filterManager.handleSaveFilter();

// Load specific filter
filterManager.loadFilter(filterName);

// Delete filter
filterManager.deleteFilter(filterName);

// Clear all filters from grid
filterManager.clearAllFilters();

// Export filters as JSON
const jsonString = filterManager.exportFilters();

// Import filters from JSON
filterManager.importFilters(jsonString);

// Get filter statistics
const stats = filterManager.getFilterStats();
```

### Grid State Structure

```javascript
{
  "configurationName": {
    "filters": {}, // AG Grid filter model
    "columns": [], // Complete column state array (order, width, visibility)
    "sorting": [], // Sort model array (multi-column sorting)
    "timestamp": "2024-01-01T00:00:00.000Z",
    "savedAt": "2024-01-01T00:00:00.000Z",
    "activeFilterCount": 3,
    "columnChanges": 2,
    "description": "Filters: status: 2 selected | Sorted by: nextETA (asc) | 8 of 10 columns visible",
    "version": "1.0",
    "name": "configurationName"
  }
}
```

## Support

For issues or feature requests:
1. Check this documentation first
2. Review browser console for error messages
3. Try clearing browser cache and localStorage
4. Contact development team with specific error details