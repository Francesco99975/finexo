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

func spoofWebGLFingerPrint(page *rod.Page) {
	_, err := page.Eval(`WebGLRenderingContext.prototype.getParameter = function () { return "spoofed"; };`)
	if err != nil {
		log.Warnf("failed to spoof WebGL fingerprinting: %v", err)
	}
}

func spoofCanvasFingerPrint(page *rod.Page) {
	_, err := page.Eval(`
		HTMLCanvasElement.prototype.getContext = function () {
			return { getImageData: () => new Uint8ClampedArray([faker.Number(0,255)]) };
		};
	`)
	if err != nil {
		log.Warnf("failed to spoof Canvas fingerprinting: %v", err)
	}
}
