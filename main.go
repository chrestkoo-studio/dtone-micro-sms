package main

import (
	"context"
	"database/sql"
	"dtone-base-sms/defined"
	"dtone-base-sms/starter"
	"dtone-micro-sms/cache"
	"dtone-micro-sms/controllers/grpc/v1"
	"dtone-micro-sms/dao"
	"dtone-micro-sms/repository"
	"dtone-micro-sms/services"
	"dtone-proto/sms"
	"dtone-std-library/dbase"
	"dtone-std-library/grpc"
	"log"
)

func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctx := context.Background()
	/*ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()*/
	cfg, err := starter.InitConfig("env/local-dev.env")
	if err != nil {
		log.Fatalf("config.LoadConfig err: %v", err)
	}

	db := dbase.InitPostgresDB(cfg.PostgresDBConfig)
	defer db.Close()

	starter.InitRedis(cfg.RedisConfig)
	starter.InitKFKProducer(cfg.KafkaProducer)
	starter.InitKafkaTopic(cfg.KafkaConsumer.BrokersAddrs, cfg.KafkaConsumer.Topics)
	starter.InitGRPC(defined.DtOneMicroGrpcPartner)

	// Init Local Cache
	cache.InitLocalCache(ctx, db)
	go initBackgroundProcess(ctx, db)

	// gRPC server setup
	srv := grpc.NewServer()

	sms.RegisterSmsServiceServer(srv.Srv, initSmsControllerDependency(db))
	starter.RunGRPC(srv)
}

func initSmsControllerDependency(db *sql.DB) *v1.SmsController {

	// Initialize Dependency injection SmsConfig DAO & Repo
	smsConfigDAO := dao.NewSmsConfigDAO(db)
	smsConfigRepo := repository.NewSmsConfigRepository(smsConfigDAO, nil)

	// Initialize Dependency injection SmsSale DAO & Repo
	smsSaleDAO := dao.NewSmsSaleDAO(db)
	smsSaleRepo := repository.NewSmsSaleRepository(smsSaleDAO)

	// Initialize Dependency injection SmsSaleMobileNumber DAO & Repo
	smsSaleMobileNumberDAO := dao.NewSmsSaleMobileNumberDAO(db)
	smsSaleMobileNumberRepo := repository.NewSmsSaleMobileNumberRepository(smsSaleMobileNumberDAO)

	// Initialize Dependency injection SmsLog DAO & Repo
	smsLogDAO := dao.NewSmsLogDAO(db)
	smsLogRepo := repository.NewSmsLogRepository(smsLogDAO)

	// Initialize Dependency Sms Service
	service := services.NewSmsService(smsConfigRepo, smsSaleRepo, smsSaleMobileNumberRepo, smsLogRepo)
	return v1.NewSmsController(service)
}

func initBackgroundProcess(ctx context.Context, db *sql.DB) {

	// Initialize Dependency injection SmsConfig DAO & Repo
	smsConfigDAO := dao.NewSmsConfigDAO(db)
	smsConfigRepo := repository.NewSmsConfigRepository(smsConfigDAO, nil)

	// Initialize Dependency injection SmsSale DAO & Repo
	smsSaleDAO := dao.NewSmsSaleDAO(db)
	smsSaleRepo := repository.NewSmsSaleRepository(smsSaleDAO)

	// Initialize Dependency injection SmsSaleMobileNumber DAO & Repo
	smsSaleMobileNumberDAO := dao.NewSmsSaleMobileNumberDAO(db)
	smsSaleMobileNumberRepo := repository.NewSmsSaleMobileNumberRepository(smsSaleMobileNumberDAO)

	// Initialize Dependency injection SmsLog DAO & Repo
	smsLogDAO := dao.NewSmsLogDAO(db)
	smsLogRepo := repository.NewSmsLogRepository(smsLogDAO)

	go services.NewKafkaService(smsConfigRepo, smsSaleRepo, smsSaleMobileNumberRepo, smsLogRepo).InitConsumer(ctx)                  // Initialize Dependency kafka Service
	go services.NewSMSBackgroundService(smsConfigRepo, smsSaleRepo, smsSaleMobileNumberRepo, smsLogRepo).InitSMSBackgroundService() // // Initialize Dependency Sms Background Service
}
