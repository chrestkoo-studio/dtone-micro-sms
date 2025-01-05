package v1

import (
	"context"
	"dtone-base-sms/config"
	"dtone-base-sms/controller"
	"dtone-base-sms/custom"
	"dtone-base-sms/defined"
	baseModel "dtone-base-sms/model"
	"dtone-base-sms/validation"
	"dtone-micro-sms/dto"
	"dtone-micro-sms/services"
	"dtone-proto/sms"
	"dtone-proto/wallet"
	"dtone-std-library/json"
	"dtone-std-library/kafka"
	"dtone-std-library/logs"
	"fmt"
	"google.golang.org/grpc/codes"
	"time"
)

type SmsController struct {
	controller.BaseController
	SmsService *services.SmsService
	sms.UnimplementedSmsServiceServer
}

func NewSmsController(service *services.SmsService) *SmsController {
	return &SmsController{SmsService: service}
}

func (ctrl *SmsController) GetSmsCostV1(ctx context.Context, in *sms.GetSmsCostReq) (*sms.GetSmsCostResp, error) {
	logMsgTemplate := fmt.Sprintf("[SmsController][GetSmsCostV1][in:%s]", json.String(in))
	ctx = custom.SetLogIdIfNotExist(ctx, "")
	logs.WithCtx(ctx).Info("%s[received]", logMsgTemplate)

	var mobileNumberInfoList []*validation.MobileNumberInfo
	tmp := make(map[string]bool)
	for _, v := range in.GetMobileNumberInfoList() {
		if _, ok := tmp[fmt.Sprintf("%v-%v", v.GetCountryCode(), v.GetNationalNumberUint64())]; !ok {
			tmp[fmt.Sprintf("%v-%v", v.GetCountryCode(), v.GetNationalNumberUint64())] = true
			mobileNumberInfoList = append(mobileNumberInfoList, &validation.MobileNumberInfo{
				RegionCode:           v.GetRegionCode(),
				CountryCode:          int(v.GetCountryCode()),
				NationalNumberUint64: v.GetNationalNumberUint64(),
			})
		}
	}

	req := dto.GetSmsCostReqDTO{
		MobileNumberInfoList: mobileNumberInfoList,
	}
	if err := req.Validate(); err != nil {
		logs.WithCtx(ctx).Error("%s[req.Validate][err:%v][req:%v]", logMsgTemplate, err, json.String(req))
		return &sms.GetSmsCostResp{
			Code:    uint32(codes.InvalidArgument),
			Message: err.Error(),
		}, nil
	}

	resp, err := ctrl.SmsService.GetSmsCost(ctx, &req)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[ctrl.SmsService.GetSmsCost][err:%v][req:%v]", logMsgTemplate, err, json.String(req))
		return &sms.GetSmsCostResp{
			Code:    uint32(codes.Internal),
			Message: err.Error(),
		}, nil
	}

	return &sms.GetSmsCostResp{
		Code:    uint32(codes.OK),
		Message: "success ",
		Data: &sms.GetSmsCostRespData{
			TotalCost: resp.TotalCost,
		},
	}, nil
}

func (ctrl *SmsController) ProcessSendSmsV1(ctx context.Context, in *sms.ProcessSendSmsReq) (*sms.ProcessSendSmsResp, error) {
	logMsgTemplate := fmt.Sprintf("[SmsController][CreateWalletTransV1][in:%s]", json.String(in))
	ctx = custom.SetLogIdIfNotExist(ctx, "")
	logs.WithCtx(ctx).Info("%s[received]", logMsgTemplate)
	err := kafka.Cli().Send(ctx, kafka.NewMessage(config.GetConfig().KafkaConsumer.Topics[defined.GetKafkaMicroSmsTopic(defined.KafkaMicroSms)], json.Stringify(services.SmsKafkaData{SmsSaleId: 10})))
	if err != nil {
		logs.WithCtx(ctx).Error("%s[kafka.Cli().Send][err:%v]", logMsgTemplate, err)
		return nil, err
	}
	return nil, err

	var mobileNumberInfoList []*validation.MobileNumberInfo
	tmp := make(map[string]bool)
	for _, v := range in.GetMobileNumberInfoList() {
		if _, ok := tmp[fmt.Sprintf("%v-%v", v.GetCountryCode(), v.GetNationalNumberUint64())]; !ok {
			tmp[fmt.Sprintf("%v-%v", v.GetCountryCode(), v.GetNationalNumberUint64())] = true
			mobileNumberInfoList = append(mobileNumberInfoList, &validation.MobileNumberInfo{
				RegionCode:           v.GetRegionCode(),
				CountryCode:          int(v.GetCountryCode()),
				NationalNumberUint64: v.GetNationalNumberUint64(),
			})
		}
	}

	req := dto.ProcessSendSmsReqDTO{
		PartnerId:            in.GetPartnerId(),
		MobileNumberInfoList: mobileNumberInfoList,
		Message:              in.Message,
	}
	if err := req.Validate(); err != nil {
		logs.WithCtx(ctx).Error("%s[req.Validate][err:%v][req:%v]", logMsgTemplate, err, json.String(req))
		return &sms.ProcessSendSmsResp{
			Code:    uint32(codes.InvalidArgument),
			Message: err.Error(),
		}, nil
	}

	// Start a transaction
	tx, err := ctrl.SmsService.SmsConfigRepo.DAO.DB.Begin()
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

	resp, err := ctrl.SmsService.ProcessSendSmsBackground(ctx, tx, &req)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[ctrl.SmsService.ProcessSendSmsBackground][err:%v][req:%v]", logMsgTemplate, err, json.String(req))
		return &sms.ProcessSendSmsResp{
			Code:    uint32(codes.Internal),
			Message: err.Error(),
		}, nil
	}

	ctrl.ConnGRpc(ctx, defined.DtOneMicroGrpcWallet)
	if ctrl.GrpcConn == nil {
		ctrl.ConnGRpc(ctx, defined.DtOneMicroGrpcWallet)
	}
	walletGRPCClient := wallet.NewWalletServiceClient(ctrl.GrpcConn.Conn())
	if walletGRPCClient == nil {
		err := fmt.Errorf("grpc client %s is nil [errCode:%d]", defined.DtOneMicroGrpcWallet, defined.ErrGRPCClientDisconnected)
		logs.WithCtx(ctx).Error("%s[wallet.NewWalletServiceClient][err:%v]", logMsgTemplate, err)
		return &sms.ProcessSendSmsResp{
			Code:    uint32(codes.Internal),
			Message: err.Error(),
		}, nil
	}

	partnerWalletReq := wallet.CreateWalletTransReq{
		PartnerId:     in.GetPartnerId(),
		WalletTypeId:  baseModel.GetDefaultWalletConfig().ID,
		TransType:     defined.WalletTransTypeSms,
		TransId:       resp.SmsSaleId,
		TotalIn:       0,
		TotalOut:      int64(resp.TotalCost),
		AdditionalMsg: "",
		Remark:        "",
		TransAt:       time.Now().Unix(),
		CreatedBy:     defined.DefaultCreatedBy,
	}
	partnerWalletResp, err := walletGRPCClient.CreateWalletTransV1(ctrl.GrpcCtx, &partnerWalletReq)
	defer ctrl.GrpcCtxCancel()
	if err != nil {
		logs.WithCtx(ctx).Error("%s[walletGRPCClient.CreateWalletTransV1][err:%v][req:%s]", logMsgTemplate, err, json.String(partnerWalletReq))
		return &sms.ProcessSendSmsResp{
			Code:    uint32(codes.Internal),
			Message: err.Error(),
		}, nil
	}

	if partnerWalletResp.Code != uint32(codes.OK) {
		logs.WithCtx(ctx).Error("%s[walletGRPCClient.CreateWalletTransV1][resp][err:%v][partnerWalletReq:%v]", logMsgTemplate, err, json.String(partnerWalletReq))
		return &sms.ProcessSendSmsResp{
			Code:    uint32(codes.InvalidArgument),
			Message: partnerWalletResp.Message,
		}, nil
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		err = fmt.Errorf("failed to commit transaction: %w", err)
		logs.WithCtx(ctx).Error("%s[tx.Commit][err:%v]", logMsgTemplate, err)
		return nil, err
	}

	/*err = kafka.Cli().Send(ctx, kafka.NewMessage(topic, json.Stringify(map[string]string{"SmsSaleId": fmt.Sprintf("%d", resp.SmsSaleId)})))
	if err != nil {
		logs.WithCtx(ctx).Error("%s[kafka.Cli().Send][err:%v]", logMsgTemplate, err)
		return nil, err
	}*/

	return &sms.ProcessSendSmsResp{
		Code:    uint32(codes.OK),
		Message: "success",
	}, nil
}
