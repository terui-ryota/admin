package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/terui-ryota/admin/internal/app/admin-web/app"
	"github.com/terui-ryota/admin/internal/app/admin-web/config"
	"github.com/terui-ryota/admin/internal/app/admin-web/handler"
	adapterimpl "github.com/terui-ryota/admin/internal/infrastructure/adapter"
	"github.com/terui-ryota/admin/internal/usecase"
	"github.com/terui-ryota/admin/pkg/logger"
	"github.com/volatiletech/sqlboiler/boil"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	//"github.com/ca-media-nantes/pick/backend/admin/internal/app/admin-web/app"
	//"github.com/ca-media-nantes/pick/backend/admin/internal/app/admin-web/config"
	//"github.com/ca-media-nantes/pick/backend/admin/internal/app/admin-web/handler"
	//adapterimpl "github.com/ca-media-nantes/pick/backend/admin/internal/infrastructure/adapter"
	//"github.com/ca-media-nantes/pick/backend/admin/internal/infrastructure/cache"
	//repositoryimpl "github.com/ca-media-nantes/pick/backend/admin/internal/infrastructure/repository"
	//"github.com/ca-media-nantes/pick/backend/admin/internal/usecase"
)

func configure() config.AdminWebConfig {
	// デフォルトで local 用 configuration を設定します。
	c := config.AdminWebConfig{
		Application: config.Application{
			Port:         8080,
			MonitorPort:  8081,
			GinDebugMode: true,
			ContextPath:  "",
			Domain:       "localhost",
		},
		Account: config.Account{
			ClientID: "xxxxxxxxxxxxxxx",
		},
		GRPCServers: config.GRPCServers{
			OfferItemGRPCServer: config.OfferItemGRPCServer{
				Host: "localhost",
				Port: 19001, // offer_itemのconfig.yamlに設定されているgrpc_portに合わせる
			},
		},
		Databases: config.Databases{
			Primary: config.Database{
				Name:     "admin",
				Host:     "localhost",
				Port:     13306,
				User:     "root",
				Password: "odessa",
				Debug:    true,
			},
			Replica: config.Database{
				Name:     "admin",
				Host:     "localhost",
				Port:     13306,
				User:     "root",
				Password: "odessa",
				Debug:    true,
			},
		},
		ZapConfig: zap.Config{
			Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
			Development: true,
			Encoding:    "console",
		},
		//ZipPasswordConfig: &config.ZipPasswordConfig{
		//	ShipmentZipPassoword: func() string {
		//		password := os.Getenv("OFFER_ITEM_V2_SHIPMENT_PASSWORD")
		//		if password == "" && os.Getenv("ENV") != "dev" { // NOTE: devなら使われていないので許容する
		//			panic("OFFER_ITEM_V2_SHIPMENT_PASSWORD is empty")
		//		}
		//		logger.Default().Infof("OFFER_ITEM_V2_SHIPMENT_PASSWORD: %s", password)
		//		return password
		//	}(),
		//},
	}
	//c.PermanOpenIDConnect.RedirectURL = fmt.Sprintf("https://%s:%d/auth/callback", c.Application.Domain, c.Application.Port)

	env := os.Getenv("ENV")
	// 以下、環境ごとに override
	switch env {
	case "dev":
		c.Application.Domain = "dev-odessa-admin.tools.ndca.jp"
		c.Account.ClientID = "xxxxxxxxxxxxxxxxxxxxxxxx"
		c.Databases.Primary.Host = "dev-pick-blue-aurora-cluster.cluster-c5uofclpyn6k.ap-northeast-1.rds.amazonaws.com"
		c.Databases.Primary.User = "dev_common_user"
		c.Databases.Replica.Host = "dev-pick-blue-aurora-cluster.cluster-ro-c5uofclpyn6k.ap-northeast-1.rds.amazonaws.com"
		c.Databases.Replica.User = "dev_common_user"
	case "stg":
		c.Application.Domain = "stg-odessa-admin.tools.ndca.jp"
		c.Account.ClientID = "xxxxxxxxxxxxxxxxxxxxxxxx"
		c.Databases.Primary.Host = "pick-common-rds-write.nantes.local"
		c.Databases.Primary.User = "stg_common_user"
		c.Databases.Replica.Host = "pick-common-rds-read.nantes.local"
		c.Databases.Replica.User = "stg_common_user"
	case "prd":
		c.Application.Domain = "odessa-admin.tools.ndca.jp"
		c.Account.ClientID = "xxxxxxxxxxxxxxxxxxxxxxxx"
		c.Databases.Primary.Host = "pick-common-rds-write.nantes.local"
		c.Databases.Primary.User = "prd_common_user"
		c.Databases.Replica.Host = "pick-common-rds-read.nantes.local"
		c.Databases.Replica.User = "prd_common_user"
	}
	// リモート環境共通項目
	if slices.Contains([]string{"dev", "stg", "prd"}, env) {
		c.Application.ContextPath = "/x"
		c.GRPCServers.OfferItemGRPCServer = config.OfferItemGRPCServer{
			Host: "offer-item-v2-api.pick.svc.cluster.local",
			Port: 5000,
		}
		c.ZapConfig.Encoding = "json"
		c.Databases.Primary.Port = 3306
		c.Databases.Replica.Port = 3306
		c.Databases.Primary.Password = os.Getenv("DATABASE_PRIMARY_PASSWORD")
		c.Databases.Replica.Password = os.Getenv("DATABASE_REPLICA_PASSWORD")
	}
	return c
}

func loadDB(c config.Database) *sql.DB {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		logger.Default().Errorf("time.LoadLocation: %w", err)
		os.Exit(1)
	}
	mysqlConfig := mysql.Config{
		DBName:               c.Name,
		User:                 c.User,
		Passwd:               c.Password,
		Addr:                 fmt.Sprintf("%s:%d", c.Host, c.Port),
		Net:                  "tcp",
		ParseTime:            true,
		Loc:                  jst,
		AllowNativePasswords: true,
	}
	// db, err := otelsql.Open("mysql", mysqlConfig.FormatDSN())
	db, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		logger.Default().Errorf("sql.Open: %w", err)
		os.Exit(1)
	}
	if err = db.Ping(); err != nil {
		logger.Default().Errorf("db.Ping: %w", err)
		os.Exit(1)
	}
	db.SetMaxIdleConns(c.MaxIdleConn)
	db.SetMaxOpenConns(c.MaxOpenConn)

	boil.SetDB(db)
	boil.DebugMode = c.Debug
	return db
}

func main() {
	c := configure()
	if err := logger.Configure(&c.ZapConfig); err != nil {
		logger.Default().Errorf("logger.Configure: %w", err)
		os.Exit(1)
	}
	//primaryDB := loadDB(c.Databases.Primary)
	//replicaDB := loadDB(c.Databases.Replica)
	//userRepository := repositoryimpl.NewUserRepositoryImpl(primaryDB, replicaDB)
	offerItemAdapter, err := adapterimpl.NewOfferItemAdapterImpl(c.GRPCServers.OfferItemGRPCServer)
	if err != nil {
		logger.Default().Errorf("adapterimpl.NewOfferItemAdapterImpl: %w", err)
		os.Exit(1)
	}
	offerItemUsecase := usecase.NewOfferItemUsecaseImpl(
		offerItemAdapter,
	)
	if err := app.NewApp(
		c.Application,
		//*handler.NewAdminHandler(
		//	usecase.NewAdminUsecaseImpl(
		//		userRepository,
		//	),
		//),
		*handler.NewAuthHandler(
			c.Application,
		),
		*handler.NewOfferItemHandler(
			offerItemUsecase,
		),
		*handler.NewSystemHandler(),
	).Start(); err != nil {
		logger.Default().Errorf("Start: %w", err)
		os.Exit(1)
	}
}
