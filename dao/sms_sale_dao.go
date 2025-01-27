package dao

import (
	"database/sql"
	"dtone-base-sms/defined"
	"dtone-micro-sms/models"
)

type SmsSaleDAO struct {
	DB *sql.DB
}

func NewSmsSaleDAO(db *sql.DB) *SmsSaleDAO {
	return &SmsSaleDAO{
		DB: db,
	}
}

// Create inserts a new sms sale log record
func (dao *SmsSaleDAO) Create(tx *sql.Tx, data *models.SmsSale) (uint64, error) {
	query := `
		INSERT INTO sms_sale (partner_id, message, total_cost, status, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := tx.QueryRow(query, data.PartnerID, data.Message, data.TotalCost, data.Status, data.CreatedAt, data.CreatedBy).
		Scan(&data.ID)
	return data.ID, err
}

func (dao *SmsSaleDAO) ConfirmPendingRecordById(id uint64) error {
	query := `
		UPDATE sms_sale
		SET status = $1,
		    updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT,
		    approved_at = EXTRACT(EPOCH FROM NOW())::BIGINT,
		    updated_by = $2
		    approved_by = $3
		WHERE id = $4 AND status = $5
	`
	_, err := dao.DB.Exec(query,
		models.SmsSaleStatusConfirmed,
		defined.DefaultUpdatedBy,
		defined.DefaultUpdatedBy,
		id,
		models.SmsSaleStatusPending,
	)
	return err
}

func (dao *SmsSaleDAO) CompleteRecordById(id uint64) error {
	query := `
		UPDATE sms_sale
		SET status = $1,
		    updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT,
		    updated_by = $2
		WHERE id = $3
	`
	_, err := dao.DB.Exec(query,
		models.SmsSaleStatusCompleted,
		defined.DefaultUpdatedBy,
		id,
	)
	return err
}
