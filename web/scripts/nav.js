document.addEventListener("DOMContentLoaded", function () {
    // Get user email from token
    const authCookie = document.cookie.split("; ").find((row) => row.startsWith("auth_token="));
    if (authCookie) {
        try {
            const token = authCookie.split("=")[1];
            const payload = JSON.parse(atob(token.split(".")[1]));
            const userEmailEl = document.getElementById("user-email");
            if (userEmailEl && payload.email) {
                userEmailEl.textContent = payload.email;
            }
        } catch (e) {
            console.error("Could not decode user info from token");
        }
    }

    // Update header based on current page
    const path = window.location.pathname;
    const activeClass = "text-white bg-blue-600 hover:bg-blue-700";
    const inactiveClass = "text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600";

    const viewBoth = document.getElementById("view-both");
    const viewGrid = document.getElementById("view-grid");
    const viewMap = document.getElementById("view-map");
    const pageTitle = document.getElementById("page-title");
    const pageSubtitle = document.getElementById("page-subtitle");
    const connectionStatus = document.getElementById("connectionStatus");
    const dataStatus = document.getElementById("dataStatus");

    // Reset all buttons to inactive state
    viewBoth.className = `inline-flex items-center px-3 py-1.5 rounded text-sm font-medium transition-colors ${inactiveClass}`;
    viewGrid.className = `inline-flex items-center px-3 py-1.5 rounded text-sm font-medium transition-colors ${inactiveClass}`;
    viewMap.className = `inline-flex items-center px-3 py-1.5 rounded text-sm font-medium transition-colors ${inactiveClass}`;

    // Set active state and page info based on path
    if (path === "/map") {
        viewMap.className = `inline-flex items-center px-3 py-1.5 rounded text-sm font-medium transition-colors ${activeClass}`;
        pageTitle.textContent = "Shipment Map";
        pageSubtitle.textContent = "Real-time Vessel Tracking";
        // Show connection status on map page
        connectionStatus.classList.remove("hidden");
        connectionStatus.classList.add("flex");
        dataStatus.classList.remove("hidden");
    } else if (path === "/shipments/grid") {
        viewGrid.className = `inline-flex items-center px-3 py-1.5 rounded text-sm font-medium transition-colors ${activeClass}`;
        pageTitle.textContent = "Shipments Grid";
        pageSubtitle.textContent = "Table view";
    } else {
        viewBoth.className = `inline-flex items-center px-3 py-1.5 rounded text-sm font-medium transition-colors ${activeClass}`;
        pageTitle.textContent = "Shipment Management";
        pageSubtitle.textContent = "Manage and track your shipments";
    }
});
