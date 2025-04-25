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

document.body.addEventListener('htmx:responseError', function(event: any) {
  const target = event.detail.target; // The element specified by hx-target
  const response = event.detail.xhr.responseText; // The HTML response
  const contentType = event.detail.xhr.getResponseHeader('Content-Type');

  // Check if the response is HTML before swapping
  if (target && contentType.includes('text/html')) {
    target.innerHTML = response;
  }
});
