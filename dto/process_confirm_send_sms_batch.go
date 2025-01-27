package dto

import (
	"errors"
)

// ProcessConfirmSendSmsBatchReqDTO for processing confirm send sms batch request
type ProcessConfirmSendSmsBatchReqDTO struct {
	PartnerId                  uint64
	ConfirmSendSmsSaleInfoList []*ConfirmSendSmsSaleInfo
}

// ConfirmSendSmsSaleInfo for confirm send sms sale info
type ConfirmSendSmsSaleInfo struct {
	SmsSaleId             uint64
	SmsSaleMobileNumberId uint64
}

// Validate Process Confirm Send Sms Request Batch
func (req *ProcessConfirmSendSmsBatchReqDTO) Validate() error {
	if req.PartnerId < 1 {
		return errors.New("partner is required")
	}
	if len(req.ConfirmSendSmsSaleInfoList) < 1 {
		return errors.New("confirm send sms info is required")
	}

	return nil
}
