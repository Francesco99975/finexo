package tools

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/labstack/gommon/log"
)

const BASE_YAHOO_URL = "https://finance.yahoo.com/quote/"              // TICKER.EXCHANGE_SUFFIX
const BASE_MARKETBEAT_URL = "https://www.marketbeat.com/stocks/"       // EXCHANGE_PREFIX/TICKER
const BASE_DIVIDENDHISTORY_URL = "https://dividendhistory.org/payout/" // EXCHANGE_TITLE?uk!/TICKER

// List of User Agents
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
}

// Randomly select a User-Agent
func getRandomUserAgent() string {
	log.Debug("Selecting a random User-Agent")
	rand.NewSource(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}

// func scrollLikeHuman(page *rod.Page) {
// 	// Simulate human-like scrolling
// 	for range 3 {
// 		page.MustEval(`window.scrollBy(0, 100)`)
// 		time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)+300))
// 	}
// }

// Simulates human-like mouse movements
func moveMouseLikeHuman(page *rod.Page) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Panic occurred while moving mouse: %v", r)
		}
	}()
	mouse := page.Mouse
	width, height := 800, 600 // Adjust viewport size

	// Move randomly within the viewport
	for range 5 {
		x := rand.Intn(width)
		y := rand.Intn(height)
		mouse.MustMoveTo(float64(x), float64(y))
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)+200))
	}

	// Click in a natural position (e.g., middle of the page)
	mouse.MustMoveTo(400, 300)
}

// Background behavior routine
func randomUserBehavior(ctx context.Context, page *rod.Page, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Debugf("Starting random behavior.")

	for {
		select {
		case <-ctx.Done():
			log.Debugf("Stopping random behavior.")
			return // Stop the Goroutine
		default:
			if page != nil {
				// Simulate human-like mouse movements
				moveMouseLikeHuman(page)
			}

			// Wait for a random time before next action
			time.Sleep(time.Second * time.Duration(rand.Intn(5)+3))
		}
	}
}
