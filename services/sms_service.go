package services

import (
	"context"
	"dtone-base-sms/cache"
	"dtone-base-sms/config"
	"dtone-base-sms/custom"
	"dtone-base-sms/defined"
	baseModel "dtone-base-sms/model"
	"dtone-micro-sms/dto"
	"dtone-micro-sms/models"
	"dtone-micro-sms/repository"
	"dtone-proto/wallet"
	"dtone-std-library/grpc"
	"dtone-std-library/grpc/metadata"
	"dtone-std-library/json"
	"dtone-std-library/kafka"
	"dtone-std-library/logs"
	"fmt"
	"google.golang.org/grpc/codes"
	"sync"
	"time"
)

type SmsService struct {
	SmsConfigRepo           *repository.SmsConfigRepository
	SmsSaleRepo             *repository.SmsSaleRepository
	SmsSaleMobileNumberRepo *repository.SmsSaleMobileNumberRepository
	SmsLogRepository        *repository.SmsLogRepository
}

func NewSmsService(
	smsConfigRepo *repository.SmsConfigRepository,
	smsSaleRepo *repository.SmsSaleRepository,
	smsSaleMobileNumberRepo *repository.SmsSaleMobileNumberRepository,
	smsLogRepository *repository.SmsLogRepository) *SmsService {
	return &SmsService{
		SmsConfigRepo:           smsConfigRepo,
		SmsSaleRepo:             smsSaleRepo,
		SmsSaleMobileNumberRepo: smsSaleMobileNumberRepo,
		SmsLogRepository:        smsLogRepository,
	}
}

func (service *SmsService) GetSmsCost(ctx context.Context, data *dto.GetSmsCostReqDTO) (*dto.GetSmsCostRespDTO, error) {
	logMsgTemplate := fmt.Sprintf("[SmsService][GetSmsCost][req:%s]", json.String(data))

	var totalCost uint64
	mobileNoInfoCost := make([]*dto.MobileNumberInfoCost, 0)
	for _, v := range data.MobileNumberInfoList {
		var smsConfig *models.SmsConfig
		for i := 0; i < models.TryReloadSmsConfigCache; i++ {
			cacheSmsConfig, ok := repository.SmsConfigRepoCache.Get(v.CountryCode)
			if cacheSmsConfig != nil {
				smsConfig = cacheSmsConfig
				break
			}
			if !ok || cacheSmsConfig == nil {
				logs.WithCtx(ctx).Error("%s[service.SmsConfigRepo.Cache.Get][i:%d][err:MissingSmsConfig][v:%v]", logMsgTemplate, i, json.String(v))
				service.SmsConfigRepo.ReloadAllActive(ctx)
				continue
			}
		}
		if smsConfig == nil {
			logs.WithCtx(ctx).Error("%s[smsConfigCache][err:MissingSmsConfig][v:%v]", logMsgTemplate, json.String(v))
			continue
		}
		mobileNoInfoCost = append(mobileNoInfoCost, &dto.MobileNumberInfoCost{
			TotalCost:        smsConfig.Cost,
			MobileNumberInfo: *v,
		})
		totalCost += smsConfig.Cost
	}

	return &dto.GetSmsCostRespDTO{TotalCost: totalCost, MobileNumberInfoCost: mobileNoInfoCost}, nil
}

func (service *SmsService) ProcessSendSmsBackground(ctx context.Context, data *dto.ProcessSendSmsReqDTO) (*dto.ProcessSendSmsRespDTO, error) {
	logMsgTemplate := fmt.Sprintf("[SmsService][ProcessSendSmsBackground][req:%s]", json.String(data))
	smsCostReq := &dto.GetSmsCostReqDTO{
		MobileNumberInfoList: data.MobileNumberInfoList,
	}
	smsCost, err := service.GetSmsCost(ctx, smsCostReq)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[service.GetSmsCost][err:%v][data:%v]", logMsgTemplate, err, json.String(smsCostReq))
		return nil, err
	}
	if smsCost == nil {
		logs.WithCtx(ctx).Error("%s[smsCost][err:smsCostNil][data:%v]", logMsgTemplate, json.String(smsCostReq))
		return nil, err
	}

	// Start a transaction
	tx, err := service.SmsConfigRepo.DAO.DB.Begin()
	if err != nil {
		err = fmt.Errorf("failed to begin transaction: %w", err)
		logs.WithCtx(ctx).Error("%s[service.SmsConfigRepo.DAO.DB.Begin][err:%v]", logMsgTemplate, err)
		return nil, err
	}

	// Ensure transaction is rolled back in case of failure
	defer func() {
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				logs.WithCtx(ctx).Error("%s[tx.Rollback][err:%v]", logMsgTemplate, err)
			}
		}
	}()

	// start save sms sale
	smsSaleData := &models.SmsSale{
		PartnerID: data.PartnerId,
		Message:   data.Message,
		TotalCost: smsCost.TotalCost,
		Status:    models.SmsSaleStatusPending,
		CreatedAt: time.Now().Unix(),
		CreatedBy: defined.DefaultCreatedBy,
	}
	smsSaleId, err := service.SmsSaleRepo.Create(tx, smsSaleData)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[service.SmsSaleRepo.Create][err:%v][smsSaleData:%v]", logMsgTemplate, err, json.String(smsSaleData))
		return nil, err
	}

	for _, info := range smsCost.MobileNumberInfoCost {
		smsSaleMobileNumberData := &models.SmsSaleMobileNumber{
			SmsSaleId:    smsSaleId,
			MobileNumber: fmt.Sprintf("+%d%d", info.MobileNumberInfo.CountryCode, info.MobileNumberInfo.NationalNumberUint64),
			Message:      data.Message,
			TotalCost:    info.TotalCost,
			RetryCount:   0,
			Status:       models.SmsSaleMobileNumberStatusPending,
			CreatedAt:    time.Now().Unix(),
			CreatedBy:    defined.DefaultCreatedBy,
		}
		_, err = service.SmsSaleMobileNumberRepo.Create(tx, smsSaleMobileNumberData)
		if err != nil {
			logs.WithCtx(ctx).Error("%s[service.SmsSaleMobileNumberRepo.Create][err:%v][smsSaleMobileNumberData:%v]", logMsgTemplate, err, json.String(smsSaleMobileNumberData))
			return nil, err
		}
	}

	callWalletTransGrpcInfo := &CallWalletTransGrpcInfo{
		PartnerId: data.PartnerId,
		SmsSaleId: smsSaleId,
		TotalCost: int64(smsCost.TotalCost),
	}
	err = service.CallWalletTransGrpc(ctx, callWalletTransGrpcInfo)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[service.callWalletTransGrpc][err:%v][callWalletTransGrpcInfo:%v]", logMsgTemplate, err, json.String(callWalletTransGrpcInfo))
		return nil, err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		err = fmt.Errorf("failed to commit transaction: %w", err)
		logs.WithCtx(ctx).Error("%s[tx.Commit][err:%v]", logMsgTemplate, err)
		return nil, err
	}

	topic := config.GetConfig().KafkaConsumer.Topics[defined.GetKafkaMicroSmsTopic(defined.KafkaMicroSms)]
	dataByte := json.Stringify(SmsKafkaData{SmsSaleId: smsSaleId})
	err = kafka.Cli().Send(ctx, kafka.NewMessage(topic, dataByte))
	if err != nil {
		logs.WithCtx(ctx).Error("%s[kafka.Cli().Send][err:%v][topic:%v][dataByte:%s]", logMsgTemplate, err, topic, dataByte)
		return nil, err
	}

	return &dto.ProcessSendSmsRespDTO{SmsSaleId: smsSaleId, TotalCost: smsCost.TotalCost}, nil
}

func (service *SmsService) ProcessSendSms(ctx context.Context, smsSaleId uint64) error {
	logMsgTemplate := fmt.Sprintf("[SmsService][ProcessSendSms][smsSaleId:%d]", smsSaleId)
	list, err := service.SmsSaleMobileNumberRepo.GetPendingRecordBySmsSaleID(smsSaleId)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[service.SmsSaleMobileNumberRepo.GetPendingRecordBySmsSaleID][err:%v]", logMsgTemplate, err)
		return err
	}

	var result sync.Map
	wg := sync.WaitGroup{}
	for _, info := range list {
		wg.Add(1)
		go func(smsSaleId uint64, info *models.SmsSaleMobileNumber) {
			defer wg.Done()
			k := defined.Gen(defined.CacheSendSms, info.ID)
			if ok := cache.GetJson(k, &info); ok {
				logs.WithCtx(ctx).Info("%s[duplicatedProcess][smsSaleMobileNumberId:%d]", logMsgTemplate, info.ID)
				return
			}

			cache.SetJson(k, info, time.Duration(config.GetConfig().MicroSmsProcConfig.RedisCachePeriodSecond)*time.Second)

			// start call send sms api.
			logs.WithCtx(ctx).Info("%s[startDoSmsAction][info:%v]", logMsgTemplate, json.String(info))

			result.Store(fmt.Sprintf("%d-%d", smsSaleId, info.ID), info.ID)

		}(smsSaleId, info)
	}
	wg.Wait()

	smsSaleMobileNumberIds := make([]uint64, 0)
	result.Range(func(key, value interface{}) bool {
		smsSaleMobileNumberIds = append(smsSaleMobileNumberIds, value.(uint64))
		return true
	})

	result.Delete(smsSaleId)

	if len(smsSaleMobileNumberIds) > 0 {
		service.SmsSaleMobileNumberRepo.CompletePendingRecordByIds(smsSaleMobileNumberIds)
	}

	logs.WithCtx(ctx).Info("%s[done][totalRecord:%d]", logMsgTemplate, len(list))
	return nil
}

type CallWalletTransGrpcInfo struct {
	PartnerId uint64
	SmsSaleId uint64
	TotalCost int64
}

func (service *SmsService) CallWalletTransGrpc(ctx context.Context, data *CallWalletTransGrpcInfo) error {
	logMsgTemplate := fmt.Sprintf("[SmsService][callWalletTransGrpc][data:%v]", json.String(data))
	grpcConn, err := grpc.Get(defined.DtOneMicroGrpcWallet)
	if err != nil {
		retryConnCount := 3
		for i := 0; i < retryConnCount; i++ {
			grpcConn, err = grpc.Get(defined.DtOneMicroGrpcWallet)
			if err != nil {
				logs.WithCtx(ctx).Error("[grpc.Get][serviceName:%s][errCode:%d][err:%v]", defined.DtOneMicroGrpcWallet, defined.ErrGRPCServerDisconnected, err)
			}
			if grpcConn != nil {
				break
			}
		}
	}
	if grpcConn == nil {
		logs.WithCtx(ctx).Error("%s[grpc.Get][err:%v][service:%s][grpcConn:%v]", logMsgTemplate, err, defined.DtOneMicroGrpcWallet, "grpcConn is nil")
		return err
	}

	walletGRPCClient := wallet.NewWalletServiceClient(grpcConn.Conn())
	if walletGRPCClient == nil {
		err = fmt.Errorf("grpc client %s is nil [errCode:%d]", defined.DtOneMicroGrpcWallet, defined.ErrGRPCClientDisconnected)
		logs.WithCtx(ctx).Error("%s[wallet.NewWalletServiceClient][err:%v]", logMsgTemplate, err)
		return err
	}

	grpcCtx, grpcCtxCancel := metadata.NewOutgoing().WithTimeout(custom.GetGRPCContextTimeout()).SetPairs(logs.LogIdCode, custom.GetLogId(ctx)).Ctx()
	partnerWalletReq := wallet.CreateWalletTransReq{
		PartnerId:     data.PartnerId,
		WalletTypeId:  baseModel.GetDefaultWalletConfig().ID,
		TransType:     defined.WalletTransTypeSms,
		TransId:       data.SmsSaleId,
		TotalIn:       0,
		TotalOut:      data.TotalCost,
		AdditionalMsg: "",
		Remark:        "",
		TransAt:       time.Now().Unix(),
		CreatedBy:     defined.DefaultCreatedBy,
	}
	partnerWalletResp, err := walletGRPCClient.CreateWalletTransV1(grpcCtx, &partnerWalletReq)
	defer grpcCtxCancel()
	if err != nil {
		logs.WithCtx(ctx).Error("%s[walletGRPCClient.CreateWalletTransV1][err:%v][req:%s]", logMsgTemplate, err, json.String(partnerWalletReq))
		return err
	}

	if partnerWalletResp.Code != uint32(codes.OK) {
		logs.WithCtx(ctx).Error("%s[walletGRPCClient.CreateWalletTransV1][resp][err:%v][partnerWalletReq:%v]", logMsgTemplate, err, json.String(partnerWalletReq))
		return err
	}
	return nil
}
