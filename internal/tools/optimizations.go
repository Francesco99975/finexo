package tools

import (
	"github.com/go-rod/rod"
	"github.com/labstack/gommon/log"
)

func disableWebRTC(page *rod.Page) {
	log.Debug("Disabling WebRTC")
	_, err := page.Eval(`(function() {
    if (navigator.mediaDevices) {
        Object.defineProperty(navigator.mediaDevices, "enumerateDevices", {
            get: function() {
                return () => Promise.resolve([]);
            }
        });
    } else {
        console.warn("navigator.mediaDevices is undefined");
    }
})();`, nil)
	if err != nil {
		log.Debugf("failed to disable WebRTC: %v", err)
	}
}
