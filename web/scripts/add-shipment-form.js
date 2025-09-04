export function handleShipmentForm(gridApi) {
  document.addEventListener("alpine:init", () => {
    Alpine.data("submitForm", () => ({
      async handleSubmit() {
        const form = document.getElementById("addShipmentForm");
        const formMessages = document.getElementById("form-messages");

        this.loading = true;

        try {
          const formData = new FormData(form);
          const payload = Object.fromEntries(formData.entries());
          const res = await fetch("/api/shipments", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload),
          });

          if (!res.ok) {
            throw new Error(`Failed: ${res.statusText}`);
          }

          const newShipment = await res.json();

          gridApi.applyTransaction({ add: [newShipment] });

          form.reset();
          formMessages.innerHTML = `<div class="text-green-600 mt-2">Shipment added successfully!</div>`;
        } catch (err) {
          console.error(err);
          formMessages.innerHTML = `<div class="text-red-600 mt-2">Error: ${err.message}</div>`;
        } finally {
          this.loading = false;
        }
      },
    }));
  });
}
