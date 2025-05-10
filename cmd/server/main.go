package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Francesco99975/finexo/cmd/boot"
	"github.com/Francesco99975/finexo/internal/database"
	"github.com/Francesco99975/finexo/internal/models"
)

func main() {
	err := boot.LoadEnvVariables()
	if err != nil {
		panic(err)
	}

	if boot.Environment == nil {
		panic("environment is nil")
	}

	// Create a root ctx and a CancelFunc which can be used to cancel retentionMap goroutine
	rootCtx := context.Background()
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	port := boot.Environment.Port

	database.Setup(boot.Environment.DSN)

	_, err = models.InitExchanges(database.DB)
	if err != nil {
		panic(err)
	}

	timeframes, err := models.GroupExachengesByTimeframe(database.DB)
	if err != nil {
		panic(err)
	}

	boot.SetupCronJobs(timeframes)

	boot.SetupDiscoveryCronJob()

	boot.SetupReportCleanupJob()

	e := createRouter(ctx)

	go func() {
		fmt.Printf("Running Server on port %s\n", port)
		fmt.Printf("Accessible locally at: http://localhost:%s\n", port)
		fmt.Printf("Accessible on the network at: http://%s:%s\n", boot.Environment.Host, port)
		fmt.Println("Press Ctrl+C to stop the server and exit.")
		e.Logger.Fatal(e.Start(":" + port))
	}()

	isDBEmpty, err := models.IsSecuritiesTableEmpty(database.DB)
	if err != nil {
		panic(err)
	}

	if isDBEmpty {
		go func() {
			err = boot.SeedDatabase(boot.Environment.DefaultLoad)
			if err != nil {
				e.Logger.Fatal(err)
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
