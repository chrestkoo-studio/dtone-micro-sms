package dto

import (
	"dtone-base-sms/validation"
	"errors"
	"strings"
)

// ProcessSendSmsReqDTO for creating a send sms request
type ProcessSendSmsReqDTO struct {
	PartnerId            uint64                         `json:"PartnerId" binding:"required"`
	MobileNumberInfoList []*validation.MobileNumberInfo `json:"MobileNumberInfoList" binding:"required"`
	Message              string                         `json:"Message" binding:"required"`
	AutoConfirm          bool                           `json:"AutoConfirm"`
}

// ProcessSendSmsRespDTO for returning result to send a sms request
type ProcessSendSmsRespDTO struct {
	SmsSaleId uint64 `json:"SmsSaleId"`
	TotalCost uint64 `json:"TotalCost"`
}

// Validate Process Send Sms Request
func (req *ProcessSendSmsReqDTO) Validate() error {
	if len(req.MobileNumberInfoList) < 1 {
		return errors.New("mobile no is required")
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		return errors.New("message is required")
	}
	if len(req.Message) > 160 {
		return errors.New("the message should not exceed 160 characters")
	}

	return nil
}
