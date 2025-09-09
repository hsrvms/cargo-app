# AG Grid Filter Management System

## Overview

The Filter Management System provides comprehensive functionality for saving, loading, and managing filter states in the AG Grid shipments view. Users can save frequently used filter combinations, share them with others, and quickly switch between different filtering scenarios.

## Features

### 1. Save Current Filters
- **Button**: "Save Filter" (blue button with save icon)
- **Description**: Saves the currently applied filters with a custom name
- **Usage**: 
  1. Apply desired filters to the grid
  2. Click "Save Filter"
  3. Enter a descriptive name when prompted
  4. Filter is saved to local storage

### 2. Load Saved Filters
- **Control**: "Select Saved Filter..." dropdown
- **Description**: Quickly apply previously saved filter configurations
- **Usage**: 
  1. Click the dropdown to see available saved filters
  2. Select a filter to apply it immediately
  3. The grid will update with the selected filter state

### 3. Clear All Filters
- **Button**: "Clear Filters" (gray button with cross icon)
- **Description**: Removes all active filters from the grid
- **Usage**: Click the button to reset the grid to show all data

### 4. Filter Management Panel
- **Button**: "Manage" (purple button with gear icon)
- **Description**: Advanced management interface for saved filters
- **Features**:
  - View all saved filters with details
  - Apply, duplicate, or delete individual filters
  - Import/export filter collections
  - Clear all saved filters

## Filter Management Panel

### Filter List
Each saved filter displays:
- **Name**: User-defined filter name
- **Filter Count**: Number of active filters
- **Save Date**: When the filter was saved
- **Description**: Auto-generated description of filter conditions
- **Actions**:
  - **Apply**: Load and apply the filter
  - **Copy**: Duplicate the filter with a new name
  - **Delete**: Remove the filter (with confirmation)

### Import/Export Functionality

#### Export Filters
- **Button**: "Export Filters"
- **Function**: Downloads all saved filters as a JSON file
- **File Format**: `ag-grid-filters-YYYY-MM-DD.json`
- **Use Cases**:
  - Backup filter configurations
  - Share filters with team members
  - Transfer filters between environments

#### Import Filters
- **Button**: "Import Filters" (file upload)
- **Function**: Loads filters from a JSON file
- **Supported Format**: JSON files exported by this system
- **Behavior**: 
  - Merges imported filters with existing ones
  - Overwrites filters with identical names (with confirmation)
  - Shows summary of imported filters

### Danger Zone
- **Clear All Saved Filters**: Permanently deletes all saved filters (with confirmation)

## Technical Details

### Storage
- **Method**: Browser localStorage
- **Key**: `ag-grid-saved-filters`
- **Format**: JSON object with filter names as keys

### Filter State Components
Each saved filter includes:
- **Filters**: AG Grid filter model (column filters)
- **Columns**: Column state (width, order, visibility)
- **Sorting**: Sort model (column sorting state)
- **Metadata**: Save date, filter count, description, version

### Browser Compatibility
- Requires localStorage support
- Compatible with modern browsers (Chrome 4+, Firefox 3.5+, Safari 4+, IE 8+)

## Usage Examples

### Example 1: Save "In Transit Shipments" Filter
1. Apply filters:
   - Status: "In Transit"
   - Next ETA: Last 7 days
2. Click "Save Filter"
3. Enter name: "In Transit This Week"
4. Filter is saved and appears in dropdown

### Example 2: Team Filter Sharing
1. Team member creates useful filter combinations
2. Exports filters using "Export Filters" button
3. Shares the JSON file with team
4. Other team members import using "Import Filters"
5. Everyone has access to the same filter presets

### Example 3: Quick Status Filtering
1. Save common status combinations:
   - "Active Shipments" (In Transit + Planned)
   - "Completed Shipments" (Delivered)
   - "Problem Shipments" (custom criteria)
2. Use dropdown to quickly switch between views

## Keyboard Shortcuts

Currently not implemented, but could be added:
- `Ctrl+S`: Save current filter
- `Ctrl+Shift+S`: Open filter management panel
- `Escape`: Close filter management panel

## Troubleshooting

### Common Issues

#### Filter Not Saving
- **Cause**: No active filters applied
- **Solution**: Apply at least one filter before saving
- **Error Message**: "No filters are currently applied. Please apply some filters before saving."

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

### Filter State Structure

```javascript
{
  "filterName": {
    "filters": {}, // AG Grid filter model
    "columns": [], // Column state array
    "sorting": [], // Sort model array
    "timestamp": "2024-01-01T00:00:00.000Z",
    "savedAt": "2024-01-01T00:00:00.000Z",
    "activeFilterCount": 3,
    "description": "status: 2 selected, recipient: contains text",
    "version": "1.0",
    "name": "filterName"
  }
}
```

## Support

For issues or feature requests:
1. Check this documentation first
2. Review browser console for error messages
3. Try clearing browser cache and localStorage
4. Contact development team with specific error details