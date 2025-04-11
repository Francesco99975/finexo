package boot

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/internal/tools"
	"github.com/labstack/gommon/log"
	"golang.org/x/sync/semaphore"
)

const maxWorkers = 7

func SeedDatabase(load int, suffix string) error {
	seeds, err := tools.ReadAllSeeds()
	if err != nil {
		return err
	}

	helpers.Shuffle(seeds)

	if len(suffix) > 0 {
		if suffix == "." {
			seeds = helpers.FilteredSlice(seeds, func(s string) bool {
				return !strings.Contains(s, ".")
			})
		} else {
			seeds = helpers.FilteredSlice(seeds, func(s string) bool {
				return strings.Contains(s, "."+suffix)
			})
		}
	}

	if load > 0 && load < len(seeds) {
		seeds = seeds[:load]
	}

	reportFilename := fmt.Sprintf("ScrapingReport-%d.log", time.Now().Unix())
	ScraperReporter, err := helpers.NewReporter(reportFilename)
	if err != nil {
		log.Errorf("failed to create reporter: %v", err)
	}
	defer func() {
		err := ScraperReporter.Close()
		if err != nil {
			log.Errorf("failed to close reporter: %v", err)
		}
	}()

	var d *tools.Discoverer
	if Environment.GoEnv == "development" {
		d, err = tools.NewDiscoverer()
		if err != nil {
			log.Errorf("Failed to create discoverer: %v", err)
		}
		defer func() {
			err := d.Close()
			if err != nil {
				log.Errorf("failed to close discoverer: %v", err)
			}
		}()
	}

	manager := models.NewBrowserManager(100)
	go manager.MonitorMemory()
	go manager.MonitorBrowserHealth()

	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(maxWorkers) // Control concurrency
	var progress atomic.Uint32
	var failed atomic.Uint32

	for _, seed := range seeds {
		wg.Add(1)

		go func(seed string) {

			defer func() {
				if r := recover(); r != nil {
					failed.Add(1)
					log.Errorf("Panic occurred while scraping seed (%s): %v", seed, r)
					err := ScraperReporter.Report(helpers.SeverityLevels.PANIC, fmt.Sprintf("was scraping seed (%s) -> %v", seed, r))
					if err != nil {
						log.Errorf("failed to report panic: %v", err)
					}
				}

				currentProgress := progress.Add(1)
				// Log Progress
				log.Infof("<<< Processed seed %d of %d: [%s] >>>", currentProgress, len(seeds), seed)

			}()

			err := tools.Scrape(seed, nil, manager, sem, &wg, d)
			if err != nil {
				failed.Add(1)
				log.Errorf("Could not Scrape <- %v", err)
				err := ScraperReporter.Report(helpers.SeverityLevels.ERROR, fmt.Sprintf("was scraping seed (%s) -> %v", seed, err))
				if err != nil {
					log.Errorf("failed to report error: %v", err)
				}
			}

		}(seed)
	}

	wg.Wait()

	failedProgress := failed.Load()
	if failedProgress > 0 {
		log.Errorf("Failed to scrape %d seeds", failedProgress)
	}
	log.Infof("All seeds have been scraped. Successfully scraped %d seeds", uint32(len(seeds))-failedProgress)

	return nil
}
