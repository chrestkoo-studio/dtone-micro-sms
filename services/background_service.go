package services

import (
	"context"
	"dtone-base-sms/custom"
	"dtone-micro-sms/repository"
	"dtone-std-library/logs"
	"time"
)

type SMSBackgroundService struct {
	SmsConfigRepo           *repository.SmsConfigRepository
	SmsSaleRepo             *repository.SmsSaleRepository
	SmsSaleMobileNumberRepo *repository.SmsSaleMobileNumberRepository
	SmsLogRepository        *repository.SmsLogRepository
}

func NewSMSBackgroundService(
	smsConfigRepo *repository.SmsConfigRepository,
	smsSaleRepo *repository.SmsSaleRepository,
	smsSaleMobileNumberRepo *repository.SmsSaleMobileNumberRepository,
	smsLogRepository *repository.SmsLogRepository) *SMSBackgroundService {
	return &SMSBackgroundService{
		SmsConfigRepo:           smsConfigRepo,
		SmsSaleRepo:             smsSaleRepo,
		SmsSaleMobileNumberRepo: smsSaleMobileNumberRepo,
		SmsLogRepository:        smsLogRepository,
	}
}

func (service *SMSBackgroundService) InitSMSBackgroundService() {
	go service.retrySendSms()
}

func (service *SMSBackgroundService) retrySendSms() {
	logMsgTemplate := "[SmsService][retrySendSms]"
	// Start a periodic reset
	ticker := time.NewTicker(1 * time.Minute) // Reset every 3 minutes
	defer ticker.Stop()

	smsService := NewSmsService(service.SmsConfigRepo, service.SmsSaleRepo, service.SmsSaleMobileNumberRepo, service.SmsLogRepository)

	for range ticker.C {
		ctx := custom.SetLogIdIfNotExist(context.Background(), custom.GenerateTraceId())
		nextRetryAt := time.Now()
		logs.WithCtx(ctx).Info("[startRetry]")
		list, err := service.SmsSaleMobileNumberRepo.GetUniqueRetryConfirmedRecords(nextRetryAt.Unix())
		if err != nil {
			logs.WithCtx(ctx).Error("%s[service.SmsSaleMobileNumberRepo.GetUniqueRetryConfirmedRecords][err:%v][nextRetryAt:%v]", logMsgTemplate, err, nextRetryAt)
			continue
		}
		if len(list) < 1 {
			logs.WithCtx(ctx).Error("%s[NoUniqueRetryConfirmedRecords][nextRetryAt:%v]", logMsgTemplate, nextRetryAt.Unix())
			continue
		}

		for _, smsSaleId := range list {
			err = smsService.processSmsToKafka(ctx, smsSaleId)
			if err != nil {
				logs.WithCtx(ctx).Error("%s[smsService.processSmsToKafka][err:%v][smsSaleId:%v]", logMsgTemplate, err, smsSaleId)
				n := nextRetryAt.Add(5 * time.Minute).Unix()
				err = service.SmsSaleMobileNumberRepo.UpdateNextRetryAt(smsSaleId, n)
				if err != nil {
					logs.WithCtx(ctx).Error("%s[service.SmsSaleMobileNumberRepo.UpdateNextRetryAt][err:%v][smsSaleId:%v][n:%v]", logMsgTemplate, err, smsSaleId, n)
				}
			}
		}
		logs.WithCtx(ctx).Info("[endRetry][timeTaken:%s]", time.Since(nextRetryAt).String())
	}
}
