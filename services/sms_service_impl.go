package services

import (
	"context"
	"dtone-micro-sms/dto"
)

type SmsServiceImpl interface {
	GetSmsCost(context.Context, *dto.GetSmsCostReqDTO) (*dto.GetSmsCostRespDTO, error)
	ProcessSendSmsBackground(context.Context, *dto.ProcessSendSmsReqDTO) (*dto.ProcessSendSmsRespDTO, error)
	ProcessSendSms(context.Context, uint64) error
	ProcessConfirmSendSmsBatch(context.Context, *dto.ProcessConfirmSendSmsBatchReqDTO) error
}
