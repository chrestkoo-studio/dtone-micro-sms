package services

import (
	"context"
	"dtone-base-sms/config"
	"dtone-base-sms/custom"
	"dtone-base-sms/defined"
	baseModel "dtone-base-sms/model"
	"dtone-micro-sms/dto"
	"dtone-micro-sms/models"
	"dtone-micro-sms/repository"
	"dtone-proto/wallet"
	"dtone-std-library/array"
	"dtone-std-library/grpc"
	"dtone-std-library/grpc/metadata"
	"dtone-std-library/json"
	"dtone-std-library/kafka"
	"dtone-std-library/logs"
	"dtone-std-library/redis"
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
	logMsgTemplate := fmt.Sprintf("[SmsService][GetSmsCost][data:%s]", json.String(data))

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
	logMsgTemplate := fmt.Sprintf("[SmsService][ProcessSendSmsBackground][data:%s]", json.String(data))
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

	smsSaleMobileNumberIds := make([]uint64, len(smsCost.MobileNumberInfoCost))
	for i, info := range smsCost.MobileNumberInfoCost {
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
		smsSaleMobileNumberId, err := service.SmsSaleMobileNumberRepo.Create(tx, smsSaleMobileNumberData)
		if err != nil {
			logs.WithCtx(ctx).Error("%s[service.SmsSaleMobileNumberRepo.Create][err:%v][smsSaleMobileNumberData:%v]", logMsgTemplate, err, json.String(smsSaleMobileNumberData))
			return nil, err
		}
		smsSaleMobileNumberIds[i] = smsSaleMobileNumberId
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		err = fmt.Errorf("failed to commit transaction: %w", err)
		logs.WithCtx(ctx).Error("%s[tx.Commit][err:%v]", logMsgTemplate, err)
		return nil, err
	}

	if data.AutoConfirm {
		err = service.ConfirmSendSms(ctx, data.PartnerId, smsSaleId, smsSaleMobileNumberIds, int64(smsCost.TotalCost))
		if err != nil {
			logs.WithCtx(ctx).Error("%s[service.ConfirmSendSms][err:%v]", logMsgTemplate, err)
			return nil, err
		}
	}

	return &dto.ProcessSendSmsRespDTO{SmsSaleId: smsSaleId, TotalCost: smsCost.TotalCost}, nil
}

func (service *SmsService) ProcessConfirmSendSmsBatch(ctx context.Context, data *dto.ProcessConfirmSendSmsBatchReqDTO) error {
	logMsgTemplate := fmt.Sprintf("[SmsService][ProcessConfirmSendSmsBatch][data:%s]", json.String(data))
	// process by sms sale id
	uniqMapList := make(map[uint64][]uint64) // map[SmsSaleId][]uint64{SmsSaleMobileNumberId}
	for _, v := range data.ConfirmSendSmsSaleInfoList {
		smsSaleMobileNumberIds, ok := uniqMapList[v.SmsSaleId]
		if ok {
			if !array.In(smsSaleMobileNumberIds, v.SmsSaleMobileNumberId) {
				uniqMapList[v.SmsSaleId] = append(uniqMapList[v.SmsSaleId], v.SmsSaleMobileNumberId)
			}
		} else {
			uniqMapList[v.SmsSaleId] = append(uniqMapList[v.SmsSaleId], v.SmsSaleMobileNumberId)
		}
	}
	for smsSaleId, smsSaleMobileNumberIds := range uniqMapList {
		validList, err := service.SmsSaleMobileNumberRepo.GetPendingRecordByIdSmsSaleID(smsSaleId, smsSaleMobileNumberIds)
		if err != nil {
			logs.WithCtx(ctx).Error("%s[service.SmsSaleMobileNumberRepo.GetPendingRecordByIdSmsSaleID][err:%v][smsSaleId:%d][smsSaleMobileNumberIds:%s]", logMsgTemplate, err, smsSaleId, json.String(smsSaleMobileNumberIds))
			return err
		}
		if len(validList) < 1 {
			logs.WithCtx(ctx).Info("%s[ValidListNoRecord][smsSaleId:%d][smsSaleMobileNumberIds:%s]", logMsgTemplate, err, smsSaleId, json.String(smsSaleMobileNumberIds))
			continue
		}
		validSmsSaleMobileNumberIds := make([]uint64, len(validList))
		var totalCost uint64
		for i, v := range validList {
			validSmsSaleMobileNumberIds[i] = v.ID
			totalCost += v.TotalCost
		}
		err = service.ConfirmSendSms(ctx, data.PartnerId, smsSaleId, validSmsSaleMobileNumberIds, int64(totalCost))
		if err != nil {
			logs.WithCtx(ctx).Error("%s[service.ConfirmSendSms][err:%v][PartnerId:%d][smsSaleId:%d][validSmsSaleMobileNumberIds:%s][totalCost:%d]", logMsgTemplate, err, data.PartnerId, smsSaleId, json.String(validSmsSaleMobileNumberIds), totalCost)
			return err
		}
	}

	return nil
}

func (service *SmsService) ConfirmSendSms(ctx context.Context, partnerId, smsSaleId uint64, smsSaleMobileNumberIds []uint64, totalCost int64) error {
	logMsgTemplate := fmt.Sprintf("[SmsService][ConfirmSendSms][partnerId:%d][smsSaleId:%d][smsSaleMobileNumberIds:%v][totalCost:%d]", partnerId, smsSaleId, json.String(smsSaleMobileNumberIds), totalCost)

	k := defined.Gen(defined.CacheConfirmSendSmsBatch, partnerId, smsSaleId)
	if ok, _ := redis.RDB().SetNx(k, true, time.Duration(config.GetConfig().MicroSmsProcConfig.RedisCachePeriodSecond)*time.Second); !ok {
		logs.WithCtx(ctx).Info("%s[duplicatedProcess][k:%d]", logMsgTemplate, k)
		return nil
	}
	defer redis.RDB().Del(k)

	callWalletTransGrpcInfo := &dto.CallWalletTransGrpcInfoDTO{
		PartnerId: partnerId,
		SmsSaleId: smsSaleId,
		TotalCost: totalCost,
	}
	err := service.CallWalletTransGrpc(ctx, callWalletTransGrpcInfo)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[service.callWalletTransGrpc][err:%v][callWalletTransGrpcInfo:%v]", logMsgTemplate, err, json.String(callWalletTransGrpcInfo))
		return err
	}

	service.SmsSaleMobileNumberRepo.ConfirmPendingRecord(smsSaleId, smsSaleMobileNumberIds)
	service.SmsSaleRepo.ConfirmPendingRecordById(smsSaleId)

	err = service.processSmsToKafka(ctx, smsSaleId)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[service.processSmsToKafka][err:%v][smsSaleId:%s]", logMsgTemplate, err, smsSaleId)
		return err
	}

	return nil
}

func (service *SmsService) processSmsToKafka(ctx context.Context, smsSaleId uint64) error {
	logMsgTemplate := fmt.Sprintf("[SmsService][ConfirmSendSms][smsSaleId:%d]", smsSaleId)

	topic := config.GetConfig().KafkaConsumer.Topics[defined.GetKafkaMicroSmsTopic(defined.KafkaMicroSms)]
	dataByte := json.Stringify(SmsKafkaData{SmsSaleId: smsSaleId})
	err := kafka.Cli().Send(context.Background(), kafka.NewMessage(topic, dataByte))
	if err != nil {
		logs.WithCtx(ctx).Error("%s[kafka.Cli().Send][err:%v][topic:%v][dataByte:%s]", logMsgTemplate, err, topic, dataByte)
		return err
	}

	return nil
}

func (service *SmsService) ProcessSendSms(ctx context.Context, smsSaleId uint64) error {
	logMsgTemplate := fmt.Sprintf("[SmsService][ProcessSendSms][smsSaleId:%d]", smsSaleId)
	list, err := service.SmsSaleMobileNumberRepo.GetConfirmedRecordBySmsSaleID(smsSaleId)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[service.SmsSaleMobileNumberRepo.GetConfirmedRecordBySmsSaleID][err:%v]", logMsgTemplate, err)
		return err
	}

	var result sync.Map
	wg := sync.WaitGroup{}
	for _, info := range list {
		wg.Add(1)
		go func(smsSaleId uint64, info *models.SmsSaleMobileNumber) {
			defer wg.Done()
			k := defined.Gen(defined.CacheSendSms, info.ID)
			if ok, _ := redis.RDB().SetNx(k, json.String(info), time.Duration(config.GetConfig().MicroSmsProcConfig.RedisCachePeriodSecond)*time.Second); !ok {
				logs.WithCtx(ctx).Info("%s[duplicatedProcess][k:%d]", logMsgTemplate, k)
				return
			}
			defer redis.RDB().Del(k)

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
		service.SmsSaleMobileNumberRepo.CompleteRecordByIds(smsSaleMobileNumberIds)
	}

	var completeSmsSaleStatus bool
	if len(smsSaleMobileNumberIds) > 0 && len(list) == len(smsSaleMobileNumberIds) { // logically no more confirm records, need to update complete status
		completeSmsSaleStatus = true
	} else { // need to check is there any confirm records. If no more, need to update complete status
		list, _ = service.SmsSaleMobileNumberRepo.GetConfirmedRecordBySmsSaleID(smsSaleId)
		if len(list) < 1 {
			completeSmsSaleStatus = true
		}
	}
	if completeSmsSaleStatus {
		service.SmsSaleRepo.CompleteRecordById(smsSaleId)
	}

	logs.WithCtx(ctx).Info("%s[done][totalRecord:%d]", logMsgTemplate, len(list))
	return nil
}

func (service *SmsService) CallWalletTransGrpc(ctx context.Context, data *dto.CallWalletTransGrpcInfoDTO) error {
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
