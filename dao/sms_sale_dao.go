package dao

import (
	"database/sql"
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
