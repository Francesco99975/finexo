package boot

import (
	"sync"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/internal/tools"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/labstack/gommon/log"
	"golang.org/x/sync/semaphore"
)

const maxWorkers = 6

func SeedDatabase() error {
	isDBEmpty, err := models.IsSecuritiesTableEmpty(database.DB)
	if err != nil {
		return err
	}

	if isDBEmpty {
		seeds, err := tools.ReadAllSeeds()
		if err != nil {
			return err
		}

		// Run Rod in headless mode
		u := launcher.New().Headless(true).MustLaunch()
		browser := rod.New().ControlURL(u).MustConnect()
		defer browser.MustClose()

		var wg sync.WaitGroup
		sem := semaphore.NewWeighted(maxWorkers) // Control concurrency

		for _, seed := range seeds {
			wg.Add(1)
			go func(seed string) {
				defer func() {
					if r := recover(); r != nil {
						log.Errorf("Panic occurred while scraping %s: %v", seed, r)
					}
				}()

				err := tools.Scrape(seed, nil, browser, sem, &wg)
				if err != nil {
					log.Errorf("Error scraping: %v", err)
				}
			}(seed)
		}

		wg.Wait()
		log.Info("All seeds have been scraped.")
	}
	return nil
}
