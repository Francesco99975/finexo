package models

import (
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/labstack/gommon/log"
	"github.com/shirou/gopsutil/mem"
)

type BrowserManager struct {
	mu                 sync.Mutex
	activeBrowser      *rod.Browser
	oldBrowser         *rod.Browser
	backupBrowser      *rod.Browser
	activeScrapers     int
	oldBrowserScrapers int // New: Track scrapers using the old browser
	maxRequests        int
	requests           int
}

func NewBrowserManager(maxRequests int) *BrowserManager {
	manager := &BrowserManager{
		maxRequests: maxRequests,
	}
	manager.activeBrowser = manager.launchNewBrowser()
	return manager
}

// 🚀 Launches a new browser with your settings
func (bm *BrowserManager) launchNewBrowser() *rod.Browser {
	log.Info("Starting new Chrome instance...")

	u := launcher.New().NoSandbox(true).Headless(true).Devtools(false).
		Set("disable-dev-shm-usage").
		Set("disable-extensions").MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	return browser
}

// 📌 Get a browser instance for scraping
func (bm *BrowserManager) GetBrowser() *rod.Browser {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.requests++
	bm.activeScrapers++

	// 🚀 If request limit is reached, start a new browser (with backup)
	if bm.requests >= bm.maxRequests {
		bm.requests = 0
		go bm.RestartBrowserSafely()
	}

	return bm.activeBrowser
}

// 📌 Called when a scraper finishes using the browser
func (bm *BrowserManager) ReleaseBrowser() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.activeScrapers--
	log.Debugf("Active scrapers remaining: %d\n", bm.activeScrapers)

	// ✅ If the scraper was using the old browser, decrease the counter
	if bm.oldBrowser != nil {
		bm.oldBrowserScrapers--
		log.Debug("Old browser scrapers remaining:", bm.oldBrowserScrapers)
	}

	// ✅ Now correctly close the old browser only when it's no longer needed
	if bm.oldBrowser != nil && bm.oldBrowserScrapers == 0 {
		log.Info("All old browser scrapers are done. Closing old Chrome...")
		bm.oldBrowser.MustClose()
		bm.oldBrowser = nil
		log.Info("Old Chrome successfully closed.")
	}
}

// 🚀 Safe Restart with Backup Browser
func (bm *BrowserManager) RestartBrowserSafely() {
	bm.mu.Lock()

	// 🛑 If a restart is already in progress, don't start another one
	if bm.oldBrowser != nil {
		bm.mu.Unlock()
		return
	}

	log.Info("🚀 Creating backup browser before restart...")

	// 1️⃣ Start a backup browser first (as a fallback in case the new one fails)
	bm.backupBrowser = bm.launchNewBrowser()

	// 2️⃣ Move the current active browser to "oldBrowser" (so scrapers finish safely)
	bm.oldBrowser = bm.activeBrowser
	bm.oldBrowserScrapers = bm.activeScrapers

	// 3️⃣ Make the backup browser the new active browser
	bm.activeBrowser = bm.backupBrowser
	bm.backupBrowser = nil // Reset backup

	bm.mu.Unlock()

	log.Info("✅ New Chrome is now active. Waiting for old scrapers to finish...")

	// 4️⃣ Old Chrome closes once scrapers are done (handled in ReleaseBrowser)
}

// 🚀 Restart Chrome if memory usage exceeds 70%
func (bm *BrowserManager) MonitorMemory() {
	for {
		time.Sleep(30 * time.Second) // Check memory every 30s

		vm, _ := mem.VirtualMemory()
		if vm.UsedPercent > 70 {
			log.Warn("🚨 High memory usage detected! Restarting Chrome...")
			go bm.RestartBrowserSafely()
		}
	}
}

func (bm *BrowserManager) MonitorBrowserHealth() {
	for {
		time.Sleep(30 * time.Second) // 🕒 Check every 30 seconds

		bm.mu.Lock()

		// ✅ Check if the browser is still responding
		_, err := bm.activeBrowser.GetCookies() // Rod Health Check
		if err == nil {
			bm.mu.Unlock()
			continue // If healthy, no need to restart
		}

		log.Warn("🚨 Browser is unresponsive! Restarting Chrome...")

		// 🚀 Restart the browser safely
		bm.requests = 0
		go bm.RestartBrowserSafely()

		bm.mu.Unlock()
	}
}
