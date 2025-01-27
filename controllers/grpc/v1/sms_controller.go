package v1

import (
	"context"
	"dtone-base-sms/controller"
	"dtone-base-sms/custom"
	"dtone-base-sms/validation"
	"dtone-micro-sms/dto"
	"dtone-micro-sms/services"
	"dtone-proto/sms"
	"dtone-std-library/json"
	"dtone-std-library/logs"
	"fmt"
	"google.golang.org/grpc/codes"
)

type SmsController struct {
	controller.BaseController
	SmsService services.SmsServiceImpl
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
		Message:              in.GetMessage(),
		AutoConfirm:          in.GetAutoConfirm(),
	}
	if err := req.Validate(); err != nil {
		logs.WithCtx(ctx).Error("%s[req.Validate][err:%v][req:%v]", logMsgTemplate, err, json.String(req))
		return &sms.ProcessSendSmsResp{
			Code:    uint32(codes.InvalidArgument),
			Message: err.Error(),
		}, nil
	}

	_, err := ctrl.SmsService.ProcessSendSmsBackground(ctx, &req)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[ctrl.SmsService.ProcessSendSmsBackground][err:%v][req:%v]", logMsgTemplate, err, json.String(req))
		return &sms.ProcessSendSmsResp{
			Code:    uint32(codes.Internal),
			Message: "process send sms failed",
		}, nil
	}

	return &sms.ProcessSendSmsResp{
		Code:    uint32(codes.OK),
		Message: "success",
	}, nil
}

func (ctrl *SmsController) ProcessConfirmSendSmsBatchV1(ctx context.Context, in *sms.ProcessConfirmSendSmsBatchReq) (*sms.ProcessConfirmSendSmsBatchResp, error) {
	logMsgTemplate := fmt.Sprintf("[SmsController][ProcessConfirmSendSmsBatchV1][in:%s]", json.String(in))
	ctx = custom.SetLogIdIfNotExist(ctx, "")
	logs.WithCtx(ctx).Info("%s[received]", logMsgTemplate)

	var list []*dto.ConfirmSendSmsSaleInfo
	err := json.ParseE(json.Stringify(in.GetConfirmSendSmsSaleInfoList()), &list)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[json.ParseE][err:%v]", logMsgTemplate, err)
		return &sms.ProcessConfirmSendSmsBatchResp{
			Code:    uint32(codes.InvalidArgument),
			Message: err.Error(),
		}, nil
	}

	req := dto.ProcessConfirmSendSmsBatchReqDTO{
		PartnerId:                  in.GetPartnerId(),
		ConfirmSendSmsSaleInfoList: list,
	}
	if err = req.Validate(); err != nil {
		logs.WithCtx(ctx).Error("%s[req.Validate][err:%v][req:%v]", logMsgTemplate, err, json.String(req))
		return &sms.ProcessConfirmSendSmsBatchResp{
			Code:    uint32(codes.InvalidArgument),
			Message: err.Error(),
		}, nil
	}

	err = ctrl.SmsService.ProcessConfirmSendSmsBatch(ctx, &req)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[ctrl.SmsService.ProcessConfirmSendSmsBatch][err:%v][req:%v]", logMsgTemplate, err, json.String(req))
		return &sms.ProcessConfirmSendSmsBatchResp{
			Code:    uint32(codes.Internal),
			Message: "process send sms failed",
		}, nil
	}

	return &sms.ProcessConfirmSendSmsBatchResp{
		Code:    uint32(codes.OK),
		Message: "success",
	}, nil
}
