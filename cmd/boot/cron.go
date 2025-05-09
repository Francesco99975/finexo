package boot

import (
	"fmt"
	"time"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/internal/tools"
	"github.com/labstack/gommon/log"
)

func SetupCronJobs(exchanges []models.Exchange) {
	for _, exchange := range exchanges {

		err := tools.AddJob(exchange.Title, fmt.Sprintf("%d %d * * *", exchange.CloseTime.Time.Minute(), exchange.CloseTime.Time.Hour()), func() {
			seeds, err := models.GetAllTickerFromExchange(database.DB, exchange.Title)
			if err != nil || len(seeds) == 0 {
				log.Errorf("<CRON> Error while getting seeds: %v", err)
				return
			}
			err = SyncDatabase(seeds)
			if err != nil {
				log.Errorf("<CRON> Error while syncing database: %v", err)
			}
		})
		if err != nil {
			log.Errorf("<CRON> Error while creating sync job: %v", err)
		}
	}
}

func SetupDiscoveryCronJob() {
	err := tools.AddJob("rng", fmt.Sprintf("%d %d * * *", 0, 9), func() {

		load := 2000
		if Environment.GoEnv == "development" {
			load = 50
		}

		err := SeedDatabase(load)
		if err != nil {
			log.Errorf("<CRON> Error while seeding database: %v", err)
		}
	})
	if err != nil {
		log.Errorf("<CRON> Error while creating discovery job: %v", err)
	}
}

func SetupReportCleanupJob() {
	err := tools.AddJob("cleanup", fmt.Sprintf("%d %d * * *", 0, 0), func() {
		freq := 24 * time.Hour * 7
		if Environment.GoEnv == "development" {
			freq = 24 * time.Hour
		}
		helpers.Cleanup(freq)
	})
	if err != nil {
		log.Errorf("<CRON> Error while creating cleanup job: %v", err)
	}
}
