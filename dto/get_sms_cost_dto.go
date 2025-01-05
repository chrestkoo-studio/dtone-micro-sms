package dto

import (
	"dtone-base-sms/validation"
	"errors"
)

// GetSmsCostReqDTO for creating a send sms request
type GetSmsCostReqDTO struct {
	MobileNumberInfoList []*validation.MobileNumberInfo `json:"MobileNumberInfoList" binding:"required"`
}

// GetSmsCostRespDTO for returning result to send a sms request
type GetSmsCostRespDTO struct {
	TotalCost            uint64 `json:"TotalCost"`
	MobileNumberInfoCost []*MobileNumberInfoCost
}

type MobileNumberInfoCost struct {
	MobileNumberInfo validation.MobileNumberInfo
	TotalCost        uint64 `json:"TotalCost"`
}

// Validate Get sms cost request
func (req *GetSmsCostReqDTO) Validate() error {
	if len(req.MobileNumberInfoList) < 1 {
		return errors.New("mobile number is required")
	}

	return nil
}
