package boot

import (
	"fmt"
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

func SyncDatabase(seeds []string) error {
	reportFilename := fmt.Sprintf("reports/sync/scraping-report_%s.log", time.Now().Format("2006-01-02_15-04-05"))
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

	manager := models.NewBrowserManager(100)
	defer manager.Close()
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
					helpers.RecordBusinessEvent("sync_panic")
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

			err := tools.Scrape(seed, nil, manager, sem, &wg, nil)
			if err != nil {
				helpers.RecordBusinessEvent("sync_failed")
				failed.Add(1)
				log.Errorf("Could not Scrape <- %v", err)
				err := ScraperReporter.Report(helpers.SeverityLevels.ERROR, fmt.Sprintf("was scraping seed (%s) -> %v", seed, err))
				if err != nil {
					log.Errorf("failed to report error: %v", err)
				}
			}
			helpers.RecordBusinessEvent("sync_successful")

		}(seed)
	}

	wg.Wait()

	failedProgress := failed.Load()

	log.Infof("<REPORT> All seeds have been scraped. Successfully scraped %d seeds", uint32(len(seeds))-failedProgress)
	if failedProgress > 0 {
		log.Errorf("<REPORT> Failed to scrape %d seeds (%.2f%%)", failedProgress, (float64(failedProgress)/float64(len(seeds)))*100)
	}

	return nil
}

func SeedDatabase(load int) error {
	seeds, err := tools.ReadAllSeeds()
	if err != nil {
		return err
	}

	helpers.Shuffle(seeds)

	if load > 0 && load < len(seeds) {
		seeds = seeds[:load]
	}

	reportFilename := fmt.Sprintf("reports/discovery/scraping-report-%s.log", time.Now().Format("2006-01-02_15-04-05"))
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

	d, err := tools.NewDiscoverer()
	if err != nil {
		log.Errorf("Failed to create discoverer: %v", err)
	}
	defer func() {
		err := d.Close()
		if err != nil {
			log.Errorf("failed to close discoverer: %v", err)
		}
	}()

	manager := models.NewBrowserManager(100)
	defer manager.Close()
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
					helpers.RecordBusinessEvent("scrape_panic_occurred")
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
				helpers.RecordBusinessEvent("scrape_failed")
				failed.Add(1)
				log.Errorf("Could not Scrape <- %v", err)
				err := ScraperReporter.Report(helpers.SeverityLevels.ERROR, fmt.Sprintf("was scraping seed (%s) -> %v", seed, err))
				if err != nil {
					log.Errorf("failed to report error: %v", err)
				}
			}
			helpers.RecordBusinessEvent("scrape_successful")

		}(seed)
	}

	wg.Wait()

	failedProgress := failed.Load()

	log.Infof("<REPORT> All seeds have been scraped. Successfully scraped %d seeds", uint32(len(seeds))-failedProgress)
	if failedProgress > 0 {
		log.Errorf("<REPORT> Failed to scrape %d seeds (%.2f%%)", failedProgress, (float64(failedProgress)/float64(len(seeds)))*100)
	}

	return nil
}
