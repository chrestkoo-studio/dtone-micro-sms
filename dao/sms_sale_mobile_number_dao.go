package dao

import (
	"database/sql"
	"dtone-base-sms/defined"
	"dtone-micro-sms/models"
	"fmt"
	"strings"
)

type SmsSaleMobileNumberDAO struct {
	DB *sql.DB
}

func NewSmsSaleMobileNumberDAO(db *sql.DB) *SmsSaleMobileNumberDAO {
	return &SmsSaleMobileNumberDAO{
		DB: db,
	}
}

// Create inserts a new sms sale mobile number record
func (dao *SmsSaleMobileNumberDAO) Create(tx *sql.Tx, data *models.SmsSaleMobileNumber) (uint64, error) {
	query := `
		INSERT INTO sms_sale_mobile_number (sms_sale_id, mobile_number, message, total_cost, retry_count, status, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := tx.QueryRow(query, data.SmsSaleId, data.MobileNumber, data.Message, data.TotalCost, data.RetryCount, data.Status, data.CreatedAt, data.CreatedBy).
		Scan(&data.ID)
	return data.ID, err
}

// GetPendingRecordBySmsSaleID retrieve pending sms sale mobile number
func (dao *SmsSaleMobileNumberDAO) GetPendingRecordBySmsSaleID(smsSaleId uint64) ([]*models.SmsSaleMobileNumber, error) {
	query := `
		SELECT id, sms_sale_id, mobile_number, message, total_cost, retry_count, status, created_at, created_by, updated_at, updated_by 
		FROM sms_sale_mobile_number 
		WHERE sms_sale_id = $1 AND status = $2
	`
	rows, err := dao.DB.Query(query, smsSaleId, models.SmsSaleMobileNumberStatusPending)
	if err != nil {
		return nil, fmt.Errorf("error querying pending record: %w", err)
	}
	defer rows.Close()

	var list []*models.SmsSaleMobileNumber
	for rows.Next() {
		var row models.SmsSaleMobileNumber
		err = rows.Scan(
			&row.ID,
			&row.SmsSaleId,
			&row.MobileNumber,
			&row.Message,
			&row.TotalCost,
			&row.RetryCount,
			&row.Status,
			&row.CreatedAt,
			&row.CreatedBy,
			&row.UpdatedAt,
			&row.UpdatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		list = append(list, &row)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return list, nil
}

// CompletePendingRecordByIds complete the pending record by ids
func (dao *SmsSaleMobileNumberDAO) CompletePendingRecordByIds(ids []uint64) error {

	// Construct placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+3)

	// Fill placeholders and args
	for i, v := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+4) // Start from $4
		args[i+3] = v                             // Add id to the args slice
	}
	// Add fixed query arguments for status and updated_by
	args[0] = models.SmsSaleMobileNumberStatusCompleted
	args[1] = defined.DefaultUpdatedBy
	args[2] = defined.DefaultUpdatedBy

	query := fmt.Sprintf(`
		UPDATE sms_sale_mobile_number
		SET status = $1, updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT, updated_by = $2, completed_at = EXTRACT(EPOCH FROM NOW())::BIGINT, completed_by = $3
		WHERE id IN (%s)
	`, strings.Join(placeholders, ", "))

	_, err := dao.DB.Exec(query, args...)
	return err
}
