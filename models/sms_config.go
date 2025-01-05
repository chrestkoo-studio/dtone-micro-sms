package models

import "time"

type SmsConfig struct {
	ID           uint32 `json:"id" db:"id"`
	CountryId    uint32 `json:"country_id" db:"country_id"`
	MobilePrefix int    `json:"mobile_prefix" db:"mobile_prefix"`
	Cost         uint64 `json:"cost" db:"cost"`
	Status       uint32 `json:"status" db:"status"`
	CreatedAt    int64  `json:"created_at" db:"created_at"`
	CreatedBy    string `json:"created_by" db:"created_by"`
	UpdatedAt    int64  `json:"updated_at" db:"updated_at"`
	UpdatedBy    string `json:"updated_by" db:"updated_by"`
}

const (
	SmsConfigStatusInactive = iota
	SmsConfigStatusActive
)

const (
	SmsConfigRedisCacheTTL  = 30 * time.Hour
	TryReloadSmsConfigCache = 2
)
