package cache

import (
	"context"
	"database/sql"
	"dtone-micro-sms/dao"
	"dtone-micro-sms/repository"
)

func InitSmsConfig(ctx context.Context, db *sql.DB) {
	smsDAO := dao.NewSmsConfigDAO(db)
	smsRepo := repository.NewSmsConfigRepository(smsDAO, nil)
	err := smsRepo.Init(ctx)
	if err != nil {
		panic(err)
	}
}
