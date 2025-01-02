package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Francesco99975/finexo/cmd/boot"
	"github.com/Francesco99975/finexo/internal/database"
)

func main() {
	err := boot.LoadEnvVariables()
	if err != nil {
		panic(err)
	}

	// Create a root ctx and a CancelFunc which can be used to cancel retentionMap goroutine
	rootCtx := context.Background()
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	port := boot.Environment.Port

	database.Setup(boot.Environment.DSN)

	e := createRouter(ctx)

	go func() {
		fmt.Printf("Running Server on port %s", port)
		e.Logger.Fatal(e.Start(":" + port))
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
