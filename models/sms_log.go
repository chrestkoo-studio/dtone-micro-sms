package models

type SmsLog struct {
	ID         uint64 `json:"id" db:"id"`
	PartnerID  uint64 `json:"partner_id" db:"uint64"`
	MobileNo   string `json:"mobile_no" db:"mobile_no"`
	Message    string `json:"message" db:"message"`
	Url        string `json:"url" db:"url"`
	Input      string `json:"input" db:"input"`
	Output     string `json:"output" db:"output"`
	RetryCount uint32 `json:"retry_count" db:"retry_count"`
	Status     uint32 `json:"status" db:"status"`
	CreatedAt  int64  `json:"created_at" db:"created_at"`
	UpdatedAt  int64  `json:"updated_at" db:"updated_at"`
}

const (
	PartnerStatusPending = iota
	PartnerStatusDone
)
