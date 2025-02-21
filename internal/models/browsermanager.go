package models

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/shirou/gopsutil/mem"
)

type BrowserManager struct {
	mu             sync.Mutex
	activeBrowser  *rod.Browser
	oldBrowser     *rod.Browser
	backupBrowser  *rod.Browser
	activeScrapers int
	requests       int
	maxRequests    int
	restarting     bool
}

func NewBrowserManager(maxRequests int) *BrowserManager {
	manager := &BrowserManager{
		maxRequests: maxRequests,
	}
	manager.activeBrowser = manager.launchNewBrowser()
	manager.backupBrowser = manager.launchNewBrowser() // Pre-launch a backup browser
	return manager
}

// 🚀 Launches a new browser
func (bm *BrowserManager) launchNewBrowser() *rod.Browser {
	fmt.Println("Starting new Chrome instance...")

	u := launcher.New().NoSandbox(true).Headless(true).Devtools(false).
		Set("disable-dev-shm-usage").
		Set("disable-extensions").MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	return browser
}

// 📌 Get the latest browser dynamically (waits if restarting)
func (bm *BrowserManager) GetBrowser() *rod.Browser {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	// 🚀 If a restart is in progress, wait for the new browser
	for bm.restarting {
		bm.mu.Unlock() // Unlock while waiting (to prevent blocking everything)
		// fmt.Println("🛑 Waiting for Chrome restart...")
		bm.mu.Lock() // Relock to check condition again
	}

	bm.requests++
	bm.activeScrapers++

	// 🚀 If request limit is reached, trigger a safe browser restart
	if bm.requests >= bm.maxRequests {
		bm.requests = 0
		go bm.RestartBrowserSafely()
	}

	return bm.activeBrowser
}

// 📌 Mark scrapers as done and close old Chrome when safe
func (bm *BrowserManager) ReleaseBrowser() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bm.activeScrapers--
	fmt.Printf("Active scrapers remaining: %d\n", bm.activeScrapers)

	// If the old browser is still around and no scrapers are using it, close it
	if bm.oldBrowser != nil && bm.activeScrapers == 0 {
		fmt.Println("All scrapers done. Closing old Chrome...")
		bm.oldBrowser.MustClose()
		bm.oldBrowser = nil
		fmt.Println("Old Chrome successfully closed.")
	}
}

// 🚀 Safe Restart with Backup Browser
func (bm *BrowserManager) RestartBrowserSafely() {
	bm.mu.Lock()

	// 🛑 If restart is already in progress, don't start another
	if bm.restarting {
		bm.mu.Unlock()
		return
	}
	bm.restarting = true // Mark that we're restarting
	bm.mu.Unlock()

	fmt.Println("🚀 Switching to backup browser before restart...")

	// 1️⃣ Immediately switch to backup browser to avoid blocking scrapers
	bm.mu.Lock()
	bm.oldBrowser = bm.activeBrowser    // Store old browser for safe shutdown
	bm.activeBrowser = bm.backupBrowser // Use backup browser immediately
	bm.mu.Unlock()

	// 2️⃣ Now, safely start a new Chrome instance
	newBrowser := bm.launchNewBrowser()

	// 3️⃣ Assign the new browser as the backup (so it's ready for the next restart)
	bm.mu.Lock()
	bm.backupBrowser = newBrowser
	bm.restarting = false // Mark restart as done
	bm.mu.Unlock()

	fmt.Println("✅ New Chrome is ready, using backup browser meanwhile.")
	// Old browser will be closed when scrapers finish (in ReleaseBrowser)
}

// 🚀 Restart Chrome if memory usage exceeds 70%
func (bm *BrowserManager) MonitorMemory() {
	for {
		time.Sleep(30 * time.Second) // Check memory every 30s

		vm, _ := mem.VirtualMemory()
		if vm.UsedPercent > 95 {
			fmt.Println("🚨 High memory usage detected! Restarting Chrome...")
			go bm.RestartBrowserSafely()
		}
	}
}
