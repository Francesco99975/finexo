package boot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/helpers"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/internal/tools"
	"github.com/labstack/gommon/log"
)

func SetupCronJobs(timeframes map[string][]string) {
	for timeframe, exchanges := range timeframes {

		closeTime := strings.Split(timeframe, "-")[1]
		minute, err := strconv.Atoi(strings.Split(closeTime, ":")[1])
		if err != nil {
			log.Errorf("<CRON> Error while parsing minute: %v", err)
			continue
		}
		hour, err := strconv.Atoi(strings.Split(closeTime, ":")[0])
		if err != nil {
			log.Errorf("<CRON> Error while parsing hour: %v", err)
			continue
		}

		err = tools.AddJob(timeframe, fmt.Sprintf("%d %d * * *", minute, hour), func() {
			seeds, err := models.GetAllTickersFromExchanges(database.DB, exchanges)
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

		err := SeedDatabase(Environment.DefaultLoad)
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
