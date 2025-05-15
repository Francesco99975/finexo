package boot

import (
	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/models"
	"github.com/labstack/gommon/log"
)

func syncDatabaseJob(exchanges []string) {
	seeds, err := models.GetAllTickersFromExchanges(database.DB, exchanges)
	if err != nil || len(seeds) == 0 {
		log.Errorf("<CRON> Error while getting seeds: %v", err)
		return
	}
	err = SyncDatabase(seeds)
	if err != nil {
		log.Errorf("<CRON> Error while syncing database: %v", err)
	}
}
