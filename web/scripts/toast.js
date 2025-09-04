export function showToast(message, type = "success") {
  const colorTypes = ["success", "error", "info"];
  if (!colorTypes.includes(type)) {
    console.error("Wrong toast color type. Available types:", colorTypes);
  }
  const container = document.getElementById("toast-container");
  const colors = {
    success: "bg-green-600",
    error: "bg-red-600",
    info: "bg-blue-600",
  };

  const toast = document.createElement("div");
  toast.className = `text-white px-4 py-2 rounded shadow-md ${colors[type]} animate-fade-in`;
  toast.innerText = message;

  container.appendChild(toast);

  setTimeout(() => {
    toast.classList.add("animate-fade-out");
    toast.addEventListener("animationend", () => toast.remove());
  }, 3000);
}
