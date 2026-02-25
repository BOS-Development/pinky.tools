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
		contactsRepository := repositories.NewContacts(db)
		contactPermissionsRepository := repositories.NewContactPermissions(db)
		forSaleItemsRepository := repositories.NewForSaleItems(db)
		purchaseTransactionsRepository := repositories.NewPurchaseTransactions(db)
		buyOrdersRepository := repositories.NewBuyOrders(db)
		salesAnalyticsRepository := repositories.NewSalesAnalytics(db)
		sdeDataRepository := repositories.NewSdeDataRepository(db)
		industryCostIndicesRepository := repositories.NewIndustryCostIndices(db)
		piPlanetsRepository := repositories.NewPiPlanets(db)
		piTaxConfigRepository := repositories.NewPiTaxConfig(db)
		piLaunchpadLabelsRepository := repositories.NewPiLaunchpadLabels(db)
		characterSkillsRepository := repositories.NewCharacterSkills(db)
		industryJobsRepository := repositories.NewIndustryJobs(db)
		jobQueueRepository := repositories.NewJobQueue(db)
		productionPlansRepository := repositories.NewProductionPlans(db)
		planRunsRepository := repositories.NewPlanRuns(db)

		var esiClient *client.EsiClient
		if settings.EsiBaseURL != "" {
			esiClient = client.NewEsiClientWithHTTPClient(settings.OAuthClientID, settings.OAuthClientSecret, &http.Client{}, settings.EsiBaseURL)
		} else {
			esiClient = client.NewEsiClient(settings.OAuthClientID, settings.OAuthClientSecret)
		}

		sdeClient := client.NewSdeClient(&http.Client{})

		contactRulesRepository := repositories.NewContactRules(db)
		autoSellContainersRepository := repositories.NewAutoSellContainers(db)
		discordNotificationsRepository := repositories.NewDiscordNotifications(db)

		assetUpdater := updaters.NewAssets(charactersAssetRepository, charactersRepository, stationsRepository, playerCorporationRepostiory, playerCorporationAssetsRepository, esiClient, usersRepository, settings.AssetUpdateConcurrency)
		sdeUpdater := updaters.NewSde(sdeClient, esiClient, sdeDataRepository, itemTypesRepository, regionsRepository, constellationsRepository, systemRepository, stationsRepository)
		marketPricesUpdater := updaters.NewMarketPrices(marketPricesRepository, esiClient)
		ccpPricesUpdater := updaters.NewCcpPrices(esiClient, marketPricesRepository)
		costIndicesUpdater := updaters.NewIndustryCostIndices(esiClient, industryCostIndicesRepository)
		autoSellUpdater := updaters.NewAutoSell(autoSellContainersRepository, forSaleItemsRepository, marketPricesRepository, stockpileMarkersRepository, purchaseTransactionsRepository)
		contactRulesUpdater := updaters.NewContactRules(contactsRepository, contactRulesRepository, contactPermissionsRepository, db)

		// Discord integration (optional — only enabled when DISCORD_BOT_TOKEN is set)
		var discordClient *client.DiscordClient
		var purchaseNotifier controllers.PurchaseNotifierInterface
		var contractCreatedNotifier controllers.ContractCreatedNotifierInterface
		var notificationsUpdater *updaters.NotificationsUpdater
		if settings.DiscordBotToken != "" {
			discordClient = client.NewDiscordClient(settings.DiscordBotToken)
			notificationsUpdater = updaters.NewNotifications(discordNotificationsRepository, discordClient)
			purchaseNotifier = notificationsUpdater
			contractCreatedNotifier = notificationsUpdater
			log.Info("discord notifications enabled")
		} else {
			log.Info("discord notifications disabled (no DISCORD_BOT_TOKEN)")
		}

		autoBuyConfigsRepository := repositories.NewAutoBuyConfigs(db)
		autoBuyUpdater := updaters.NewAutoBuy(autoBuyConfigsRepository, buyOrdersRepository, marketPricesRepository, purchaseTransactionsRepository)
		autoFulfillUpdater := updaters.NewAutoFulfill(db, buyOrdersRepository, forSaleItemsRepository, purchaseTransactionsRepository, contactPermissionsRepository, usersRepository, purchaseNotifier)

		characterSkillsUpdater := updaters.NewCharacterSkillsUpdater(usersRepository, charactersRepository, characterSkillsRepository, esiClient)
		industryJobsUpdater := updaters.NewIndustryJobsUpdater(usersRepository, charactersRepository, playerCorporationRepostiory, industryJobsRepository, jobQueueRepository, esiClient)

		piUpdater := updaters.NewPiUpdater(usersRepository, charactersRepository, piPlanetsRepository, esiClient, systemRepository, sdeDataRepository)
		if notificationsUpdater != nil {
			piUpdater.WithStallNotifier(notificationsUpdater)
		}

		assetUpdater.WithAutoSellUpdater(autoSellUpdater)
		marketPricesUpdater.WithAutoSellUpdater(autoSellUpdater)
		assetUpdater.WithAutoBuyUpdater(autoBuyUpdater)
		marketPricesUpdater.WithAutoBuyUpdater(autoBuyUpdater)
		assetUpdater.WithAutoFulfillUpdater(autoFulfillUpdater)
		marketPricesUpdater.WithAutoFulfillUpdater(autoFulfillUpdater)

		controllers.NewStatic(router, sdeUpdater)
		controllers.NewCharacters(router, charactersRepository, assetUpdater, esiClient, contactRulesUpdater)
		controllers.NewUsers(router, usersRepository, usersRepository)
		controllers.NewAssets(router, assetsRepository)
		controllers.NewCorporations(router, esiClient, playerCorporationRepostiory, assetUpdater, contactRulesUpdater)
		controllers.NewStockpileMarkers(router, stockpileMarkersRepository)
		controllers.NewStockpiles(router, assetsRepository)
		controllers.NewMarketPrices(router, marketPricesUpdater)
		controllers.NewJanice(router)
		controllers.NewContacts(router, contactsRepository, contactPermissionsRepository, db)
		controllers.NewContactPermissions(router, contactPermissionsRepository)
		controllers.NewForSaleItems(router, forSaleItemsRepository, contactPermissionsRepository)
		controllers.NewPurchases(router, db, purchaseTransactionsRepository, forSaleItemsRepository, contactPermissionsRepository, usersRepository, purchaseNotifier, contractCreatedNotifier)
		controllers.NewBuyOrders(router, buyOrdersRepository, contactPermissionsRepository, autoFulfillUpdater)
		controllers.NewItemTypes(router, itemTypesRepository)
		controllers.NewAnalytics(router, salesAnalyticsRepository)
		controllers.NewAutoSellContainers(router, autoSellContainersRepository, autoSellUpdater, forSaleItemsRepository)
		controllers.NewAutoBuyConfigs(router, autoBuyConfigsRepository, autoBuyUpdater, buyOrdersRepository, autoFulfillUpdater)
		controllers.NewReactions(router, sdeDataRepository, marketPricesRepository, industryCostIndicesRepository)
		controllers.NewContactRules(router, contactRulesRepository, contactRulesUpdater)
		if discordClient != nil {
			controllers.NewDiscordNotifications(router, discordNotificationsRepository, discordClient, notificationsUpdater)
		}
		controllers.NewPi(router, piPlanetsRepository, piTaxConfigRepository, sdeDataRepository, charactersRepository, systemRepository, itemTypesRepository, marketPricesRepository, piLaunchpadLabelsRepository, stockpileMarkersRepository)
		controllers.NewIndustry(router, industryJobsRepository, jobQueueRepository, sdeDataRepository, marketPricesRepository, industryCostIndicesRepository)
		userStationsRepository := repositories.NewUserStations(db)
		controllers.NewProductionPlans(router, productionPlansRepository, sdeDataRepository, jobQueueRepository, marketPricesRepository, industryCostIndicesRepository, charactersRepository, playerCorporationRepostiory, userStationsRepository, planRunsRepository)
		controllers.NewUserStations(router, userStationsRepository)

		transportProfilesRepo := repositories.NewTransportProfiles(db)
		jfRoutesRepo := repositories.NewJFRoutes(db)
		transportJobsRepo := repositories.NewTransportJobs(db)
		triggerConfigRepo := repositories.NewTransportTriggerConfig(db)
		controllers.NewTransportation(router, transportProfilesRepo, jfRoutesRepo, transportJobsRepo, triggerConfigRepo, jobQueueRepository, marketPricesRepository, systemRepository, esiClient)

		group.Go(router.Run(ctx))

		// Start SDE update scheduler (24h)
		sdeRunner := runners.NewSdeRunner(sdeUpdater, 24*time.Hour)
		group.Go(func() error {
			return sdeRunner.Run(ctx)
		})

		// Start market price update scheduler (6h)
		marketPricesRunner := runners.NewMarketPricesRunner(marketPricesUpdater, 6*time.Hour)
		group.Go(func() error {
			return marketPricesRunner.Run(ctx)
		})

		// Start CCP adjusted prices update scheduler (1h)
		ccpPricesRunner := runners.NewCcpPricesRunner(ccpPricesUpdater, 1*time.Hour)
		group.Go(func() error {
			return ccpPricesRunner.Run(ctx)
		})

		// Start industry cost indices update scheduler (1h)
		costIndicesRunner := runners.NewIndustryCostIndicesRunner(costIndicesUpdater, 1*time.Hour)
		group.Go(func() error {
			return costIndicesRunner.Run(ctx)
		})

		// Start asset update scheduler (configurable, default 1h)
		assetsRunner := runners.NewAssetsRunner(assetUpdater, usersRepository, time.Duration(settings.AssetUpdateIntervalSec)*time.Second)
		group.Go(func() error {
			return assetsRunner.Run(ctx)
		})

		// Start PI update scheduler (configurable, default 1h)
		piRunner := runners.NewPiRunner(piUpdater, time.Duration(settings.PiUpdateIntervalSec)*time.Second)
		group.Go(func() error {
			return piRunner.Run(ctx)
		})

		// Start contract sync scheduler (15 minutes)
		contractSyncUpdater := updaters.NewContractSync(purchaseTransactionsRepository, charactersRepository, playerCorporationRepostiory, esiClient)
		contractSyncRunner := runners.NewContractSyncRunner(contractSyncUpdater, 15*time.Minute)
		group.Go(func() error {
			return contractSyncRunner.Run(ctx)
		})

		// Start character skills update scheduler (configurable, default 6h)
		skillsRunner := runners.NewCharacterSkillsRunner(characterSkillsUpdater, time.Duration(settings.SkillsUpdateIntervalSec)*time.Second)
		group.Go(func() error {
			return skillsRunner.Run(ctx)
		})

		// Start industry jobs update scheduler (configurable, default 10m)
		industryJobsRunner := runners.NewIndustryJobsRunner(industryJobsUpdater, time.Duration(settings.IndustryJobsUpdateIntervalSec)*time.Second)
		group.Go(func() error {
			return industryJobsRunner.Run(ctx)
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
