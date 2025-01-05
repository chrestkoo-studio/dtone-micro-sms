package models

type SmsSale struct {
	ID          uint64 `json:"id" db:"id"`
	PartnerID   uint64 `json:"partner_id" db:"uint64"`
	Message     string `json:"message" db:"message"`
	TotalCost   uint64 `json:"total_cost" db:"total_cost"`
	Status      uint32 `json:"status" db:"status"`
	CreatedAt   int64  `json:"created_at" db:"created_at"`
	CreatedBy   string `json:"created_by" db:"created_by"`
	UpdatedAt   int64  `json:"updated_at" db:"updated_at"`
	UpdatedBy   string `json:"updated_by" db:"updated_by"`
	ApprovedAt  int64  `json:"approved_at" db:"approved_at"`
	ApprovedBy  string `json:"approved_by" db:"approved_by"`
	RejectedAt  int64  `json:"rejected_at" db:"rejected_at"`
	RejectedBy  string `json:"rejected_by" db:"rejected_by"`
	CancelledAt int64  `json:"cancelled_at" db:"cancelled_at"`
	CancelledBy string `json:"cancelled_by" db:"cancelled_by"`
}

const (
	SmsSaleStatusPending = iota
	SmsSaleStatusApproved
	SmsSaleStatusRejected
	SmsSaleStatusCancelled
)
