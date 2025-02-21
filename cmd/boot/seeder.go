package boot

import (
	"sync"

	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/internal/tools"
	"github.com/labstack/gommon/log"
	"golang.org/x/sync/semaphore"
)

const maxWorkers = 5

func SeedDatabase() error {
	seeds, err := tools.ReadAllSeeds()
	if err != nil {
		return err
	}

	manager := models.NewBrowserManager(100)
	go manager.MonitorMemory()

	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(maxWorkers) // Control concurrency

	for _, seed := range seeds {
		wg.Add(1)
		go func(seed string) {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("Panic occurred while scraping seed (%s): %v", seed, r)
				}
			}()

			err := tools.Scrape(seed, nil, manager, sem, &wg)
			if err != nil {
				log.Errorf("Could not Scrape <- %v", err)
			}
		}(seed)
	}

	wg.Wait()
	log.Info("All seeds have been scraped.")

	return nil
}
