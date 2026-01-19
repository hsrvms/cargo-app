/**
 * Custom Renderer for Container Dropdowns
 * Smart positioning: Detects screen edge and opens UP if near the bottom.
 */
export class ContainerDropdownRenderer {
    init(params) {
        this.params = params;
        this.containers = params.data?.containers || [];

        this.eGui = document.createElement("div");
        this.eGui.className = "relative inline-block";

        if (!this.containers || this.containers.length === 0) {
            this.eGui.innerHTML = '<span class="text-gray-400 italic">N/A</span>';
            return;
        }

        const button = document.createElement("button");
        button.className = "flex items-center gap-1 text-cyan-200 hover:text-blue-800 font-medium focus:outline-none";
        button.innerHTML = `
      <span>${this.containers.length} Container${this.containers.length > 1 ? "s" : ""}</span>
      <svg class="w-3 h-3 transition-transform duration-200" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>
    `;

        const dropdown = document.createElement("div");
        dropdown.className = "hidden absolute left-0 w-48 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded shadow-lg z-[9999]";

        const ul = document.createElement("ul");
        ul.className = "py-1 text-sm text-gray-700 dark:text-gray-200 max-h-48 overflow-y-auto";

        this.containers.forEach((container) => {
            const containerNum = container.number || container.name || container.id || "Unknown";

            const li = document.createElement("li");
            li.className = "group flex items-center justify-between px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-700 cursor-pointer transition-colors";

            const textSpan = document.createElement("span");
            textSpan.textContent = containerNum;

            const iconSpan = document.createElement("span");
            iconSpan.className = "text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity text-xs font-semibold";
            iconSpan.textContent = "Copy";

            li.appendChild(textSpan);
            li.appendChild(iconSpan);

            li.addEventListener("click", async (e) => {
                e.stopPropagation();
                try {
                    await navigator.clipboard.writeText(containerNum);
                    iconSpan.textContent = "Copied!";
                    iconSpan.className = "text-green-600 font-bold text-xs opacity-100";
                    setTimeout(() => {
                        iconSpan.textContent = "Copy";
                        iconSpan.className = "text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity text-xs";
                    }, 1500);
                } catch (err) {
                    console.error("Failed to copy text: ", err);
                }
            });

            ul.appendChild(li);
        });

        dropdown.appendChild(ul);
        this.eGui.appendChild(button);
        this.eGui.appendChild(dropdown);

        button.addEventListener("click", (e) => {
            e.stopPropagation();

            const isHidden = dropdown.classList.contains("hidden");

            document.querySelectorAll(".container-dropdown-open").forEach((el) => {
                if (el !== dropdown) {
                    el.classList.add("hidden");
                    el.classList.remove("container-dropdown-open");
                    const otherRow = el.closest(".ag-row");
                    if (otherRow) otherRow.style.zIndex = "";
                }
            });

            if (isHidden) {
                const rect = button.getBoundingClientRect();
                const spaceBelow = window.innerHeight - rect.bottom;
                const dropdownHeight = 200;

                dropdown.classList.remove("top-full", "mt-1", "bottom-full", "mb-1");

                if (spaceBelow < dropdownHeight) {
                    dropdown.classList.add("bottom-full", "mb-1");
                } else {
                    dropdown.classList.add("top-full", "mt-1");
                }

                dropdown.classList.remove("hidden");
                dropdown.classList.add("container-dropdown-open");

                const currentRow = this.eGui.closest(".ag-row");
                if (currentRow) currentRow.style.zIndex = "50";

                const svg = button.querySelector("svg");
                if (svg) svg.style.transform = "rotate(180deg)";
            } else {
                dropdown.classList.add("hidden");
                dropdown.classList.remove("container-dropdown-open");

                const currentRow = this.eGui.closest(".ag-row");
                if (currentRow) currentRow.style.zIndex = "";

                const svg = button.querySelector("svg");
                if (svg) svg.style.transform = "rotate(0deg)";
            }
        });

        this.clickOutsideHandler = (event) => {
            if (!this.eGui.contains(event.target)) {
                dropdown.classList.add("hidden");
                dropdown.classList.remove("container-dropdown-open");

                const row = this.eGui.closest(".ag-row");
                if (row) row.style.zIndex = "";

                const svg = button.querySelector("svg");
                if (svg) svg.style.transform = "rotate(0deg)";
            }
        };
        document.addEventListener("click", this.clickOutsideHandler);
    }

    getGui() {
        return this.eGui;
    }

    destroy() {
        document.removeEventListener("click", this.clickOutsideHandler);
    }
}
