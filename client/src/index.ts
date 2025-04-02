import Alpine from "alpinejs";
import "./css/style.css";

import htmx from "htmx.org";

declare global {
  interface Window {
    htmx: typeof htmx;
    Alpine: typeof Alpine;
  }
}

window.htmx = htmx;
window.Alpine = Alpine;

Alpine.start();
