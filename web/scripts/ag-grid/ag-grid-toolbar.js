export function deleteSelectedBtnEvent(gridApi) {
  const selectedRows = gridApi.getSelectedRows();
  if (selectedRows.length === 0) {
    alert("No rows selected!");
    return;
  }

  if (!confirm(`Delete ${selectedRows.length} shipment(s)?`)) return;

  const ids = selectedRows.map((row) => row.id);

  fetch("/api/shipments/bulk-delete", {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ shipmentIDs: ids }),
  })
    .then((res) => res.json())
    .then(() => {
      gridApi.applyTransaction({ remove: selectedRows });
    })
    .catch((err) => {
      console.log("Delete failed:", err);
      alert("Delete failed");
    });
}
