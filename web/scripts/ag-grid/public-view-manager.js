export class PublicViewManager {
    constructor(gridApi) {
        this.gridApi = gridApi;
    }

    async createFromSelection() {
        const selectedRows = this.gridApi.getSelectedRows();

        if (!selectedRows || selectedRows.length === 0) {
            this.showToast("Please select at least one shipment to share.", "warning");
            return;
        }

        const name = prompt(`Create public link for ${selectedRows.length} shipment(s)?\n\nEnter a name for this view (e.g., "Client X - Jan Import"):`);

        if (!name || name.trim() === "") return;


        const payload = {
            name: name.trim(),
            shipmentIDs: selectedRows.map((row) => row.id),
        };


        try {

            const response = await fetch("/api/shipments/public-views", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });

            console.log(response)

            const data = await response.json();

            console.log("data:",data)
            if (!response.ok) {
                throw new Error(data.error || "Failed to create public view");
            }


            this.showSuccessModal(data.name, data.shareLink, data.expiresAt);


            this.gridApi.deselectAll();
        } catch (error) {
            console.error("Public View Error:", error);
            this.showToast(error.message, "error");
        }
    }

    /**
     * Displays the generated link.
     * Using a prompt() here is a simple hack because it pre-selects text for easy copying.
     * You can replace this with a nice custom Modal if you prefer.
     */
    showSuccessModal(name, link, expiresAt) {
        const msg = `✅ View "${name}" Created!\n\nLink expires: ${new Date(expiresAt).toLocaleDateString()}\n\nCopy this link:`;
        window.prompt(msg, link);
    }

    /**
     * Helper to reuse your toast styles from FilterManager
     */
    showToast(message, type = "info") {
        // Reuse the toast logic if it's globally available, or simple alert
        if (window.filterManager && typeof window.filterManager.showToast === "function") {
            window.filterManager.showToast(message, type);
        } else {
            alert(message);
        }
    }
}
