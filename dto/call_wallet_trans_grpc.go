package dto

import (
	"errors"
)

type CallWalletTransGrpcInfoDTO struct {
	PartnerId uint64
	SmsSaleId uint64
	TotalCost int64
}

// Validate Process Confirm Send Sms Request Batch
func (req *CallWalletTransGrpcInfoDTO) Validate() error {
	if req.PartnerId < 1 {
		return errors.New("partner is required")
	}
	if req.SmsSaleId < 1 {
		return errors.New("sms sale id is required")
	}
	if req.TotalCost < 1 {
		return errors.New("total cost is required")
	}

	return nil
}
