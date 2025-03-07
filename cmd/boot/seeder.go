package boot

import (
	"fmt"
	"sync"

	"github.com/Francesco99975/finexo/internal/helpers"
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

	ScraperReporter, err := helpers.NewReporter("scraper.log")
	if err != nil {
		log.Errorf("failed to create reporter: %v", err)
	}

	defer func() {
		err := ScraperReporter.Close()
		if err != nil {
			log.Errorf("failed to close reporter: %v", err)
		}
	}()

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
					err := ScraperReporter.Report(helpers.SeverityLevels.PANIC, fmt.Sprintf("was scraping seed (%s) -> %v", seed, r))
					if err != nil {
						log.Errorf("failed to report panic: %v", err)
					}
				}
			}()

			err := tools.Scrape(seed, nil, manager, sem, &wg)
			if err != nil {
				log.Errorf("Could not Scrape <- %v", err)
				err := ScraperReporter.Report(helpers.SeverityLevels.ERROR, fmt.Sprintf("was scraping seed (%s) -> %v", seed, err))
				if err != nil {
					log.Errorf("failed to report error: %v", err)
				}
			}
		}(seed)
	}

	wg.Wait()
	log.Info("All seeds have been scraped.")

	return nil
}
