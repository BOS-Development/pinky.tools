package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/annymsMthd/industry-tool/internal/client"
	"github.com/annymsMthd/industry-tool/internal/controllers"
	log "github.com/annymsMthd/industry-tool/internal/logging"
	"github.com/annymsMthd/industry-tool/internal/repositories"
	"github.com/annymsMthd/industry-tool/internal/runners"
	"github.com/annymsMthd/industry-tool/internal/updaters"
	"github.com/annymsMthd/industry-tool/internal/web"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var rootCmd = &cobra.Command{
	Use:   "industry-tool",
	Short: "eve group industry tool",
	Long:  `eve group industry tool`,
	Run: func(cmd *cobra.Command, args []string) {
		settings, err := GetSettings()
		if err != nil {
			log.Fatal("failed getting settings", "error", err)
		}

		err = settings.DatabaseSettings.WaitForDatabaseToBeOnline(30)
		if err != nil {
			log.Fatal("failed waiting for database", "error", err)
		}

		err = settings.DatabaseSettings.MigrateUp()
		if err != nil {
			log.Fatal("failed to migrate database", "error", err)
		}

		db, err := settings.DatabaseSettings.EnsureDatabaseExistsAndGetConnection()
		if err != nil {
			log.Fatal("failed to get database", "error", err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		group, ctx := errgroup.WithContext(ctx)

		log.Info("starting services")

		router := web.NewRouter(settings.Port, settings.BackendKey)

		fuzzWorks := client.NewFuzzWorks(&http.Client{})

		charactersRepository := repositories.NewCharacterRepository(db)
		itemTypesRepository := repositories.NewItemTypeRepository(db)
		charactersAssetRepository := repositories.NewCharacterAssets(db)
		regionsRepository := repositories.NewRegions(db)
		constellationsRepository := repositories.NewConstellations(db)
		systemRepository := repositories.NewSolarSystems(db)
		stationsRepository := repositories.NewStations(db)
		usersRepository := repositories.NewUserRepository(db)
		assetsRepository := repositories.NewAssets(db)
		playerCorporationRepostiory := repositories.NewPlayerCorporations(db)
		playerCorporationAssetsRepository := repositories.NewCorporationAssets(db)
		stockpileMarkersRepository := repositories.NewStockpileMarkers(db)
		marketPricesRepository := repositories.NewMarketPrices(db)

		esiClient := client.NewEsiClient(settings.OAuthClientID, settings.OAuthClientSecret)

		assetUpdater := updaters.NewAssets(charactersAssetRepository, charactersRepository, stationsRepository, playerCorporationRepostiory, playerCorporationAssetsRepository, esiClient)
		staticUpdater := updaters.NewStatic(fuzzWorks, itemTypesRepository, regionsRepository, constellationsRepository, systemRepository, stationsRepository)
		marketPricesUpdater := updaters.NewMarketPrices(marketPricesRepository, esiClient)

		controllers.NewStatic(router, staticUpdater)
		controllers.NewCharacters(router, charactersRepository)
		controllers.NewUsers(router, usersRepository, assetUpdater)
		controllers.NewAssets(router, assetsRepository)
		controllers.NewCorporations(router, esiClient, playerCorporationRepostiory)
		controllers.NewStockpileMarkers(router, stockpileMarkersRepository)
		controllers.NewStockpiles(router, assetsRepository)
		controllers.NewMarketPrices(router, marketPricesUpdater)
		controllers.NewJanice(router)

		group.Go(router.Run(ctx))

		// Start market price update scheduler
		marketPricesRunner := runners.NewMarketPricesRunner(marketPricesUpdater, 6*time.Hour)
		group.Go(func() error {
			return marketPricesRunner.Run(ctx)
		})

		log.Info("services started")

		eventChan := make(chan os.Signal, 1)
		signal.Notify(eventChan, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-eventChan:
		case <-ctx.Done():
		}

		log.Info("services stopping")

		cancel()

		if err := group.Wait(); err != nil {
			log.Fatal("errgroup failed", "error", err)
		}
	},
}

// Execute the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
