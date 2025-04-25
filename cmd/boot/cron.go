package boot

import (
	"fmt"

	"github.com/Francesco99975/finexo/internal/models"
	"github.com/Francesco99975/finexo/internal/tools"
	"github.com/labstack/gommon/log"
)

func SetupCronJobs(exchanges []models.Exchange) {
	for _, exchange := range exchanges {
		suffix := "."
		if exchange.Suffix.Valid {
			suffix = exchange.Suffix.String
		}
		err := tools.AddJob(suffix, fmt.Sprintf("%d %d * * *", exchange.CloseTime.Time.Minute(), exchange.CloseTime.Time.Hour()), func() {

			err := SeedDatabase(500, suffix)
			if err != nil {
				log.Errorf("Error while seeding database: %v", err)
			}
		})
		if err != nil {
			log.Errorf("Error while creating job: %v", err)
		}
	}
}
