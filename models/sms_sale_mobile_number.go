package models

type SmsSaleMobileNumber struct {
	ID           uint64 `json:"id" db:"id"`
	SmsSaleId    uint64 `json:"sms_sale_id" db:"sms_sale_id"`
	MobileNumber string `json:"mobile_number" db:"mobile_number"`
	Message      string `json:"message" db:"message"`
	TotalCost    uint64 `json:"total_cost" db:"total_cost"`
	RetryCount   uint32 `json:"retry_count" db:"retry_count"`
	Status       uint32 `json:"status" db:"status"`
	CreatedAt    int64  `json:"created_at" db:"created_at"`
	CreatedBy    string `json:"created_by" db:"created_by"`
	UpdatedAt    int64  `json:"updated_at" db:"updated_at"`
	UpdatedBy    string `json:"updated_by" db:"updated_by"`
	CompletedAt  int64  `json:"completed_at" db:"completed_at"`
	CompletedBy  string `json:"completed_by" db:"completed_by"`
	RejectedAt   int64  `json:"rejected_at" db:"rejected_at"`
	RejectedBy   string `json:"rejected_by" db:"rejected_by"`
	CancelledAt  int64  `json:"cancelled_at" db:"cancelled_at"`
	CancelledBy  string `json:"cancelled_by" db:"cancelled_by"`
}

const (
	SmsSaleMobileNumberStatusPending = iota
	SmsSaleMobileNumberStatusConfirmed
	SmsSaleMobileNumberStatusCompleted
	SmsSaleMobileNumberStatusRejected
	SmsSaleMobileNumberStatusCancelled
)
