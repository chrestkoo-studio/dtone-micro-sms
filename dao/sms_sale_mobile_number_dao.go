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

// GetUniqueRetryConfirmedRecords retrieve retry confirm sms sale mobile number
func (dao *SmsSaleMobileNumberDAO) GetUniqueRetryConfirmedRecords(nextRetryAt int64) ([]uint64, error) {
	query := `
		SELECT sms_sale_id
		FROM sms_sale_mobile_number 
		WHERE status = $1 AND next_retry_at <= $2 
		GROUP BY sms_sale_id
		ORDER BY sms_sale_id ASC
		limit 100
	`
	rows, err := dao.DB.Query(query, models.SmsSaleMobileNumberStatusConfirmed, nextRetryAt)
	if err != nil {
		return nil, fmt.Errorf("error querying unique confirmed record: %w", err)
	}
	defer rows.Close()

	var list []uint64
	for rows.Next() {
		var row models.SmsSaleMobileNumber
		err = rows.Scan(&row.SmsSaleId)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		list = append(list, row.SmsSaleId)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return list, nil
}

// GetConfirmedRecordBySmsSaleID retrieve confirmed sms sale mobile number
func (dao *SmsSaleMobileNumberDAO) GetConfirmedRecordBySmsSaleID(smsSaleId uint64) ([]*models.SmsSaleMobileNumber, error) {
	query := `
		SELECT id, sms_sale_id, mobile_number, message, total_cost, retry_count, status, created_at, created_by, updated_at, updated_by 
		FROM sms_sale_mobile_number 
		WHERE sms_sale_id = $1 AND status = $2
	`
	rows, err := dao.DB.Query(query, smsSaleId, models.SmsSaleMobileNumberStatusConfirmed)
	if err != nil {
		return nil, fmt.Errorf("error querying confirmed record: %w", err)
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

// GetPendingRecordByIdSmsSaleID retrieve sms sale mobile number pending records by ids and sms sale id
func (dao *SmsSaleMobileNumberDAO) GetPendingRecordByIdSmsSaleID(smsSaleId uint64, ids []uint64) ([]*models.SmsSaleMobileNumber, error) {

	// Construct placeholders for the IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids)+2)

	// Fill placeholders and args
	for i, v := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+3) // Start from $3
		args[i+2] = v                             // Add id to the args slice
	}
	// Add fixed query arguments for status and updated_by
	args[0] = smsSaleId
	args[1] = models.SmsSaleMobileNumberStatusPending

	query := fmt.Sprintf(`
		SELECT id, sms_sale_id, mobile_number, message, total_cost, retry_count, status, created_at, created_by, updated_at, updated_by 
		FROM sms_sale_mobile_number 
		WHERE sms_sale_id = $1 AND status = $2 AND id IN (%s)`, strings.Join(placeholders, ", "))

	rows, err := dao.DB.Query(query, args...)
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

// ConfirmPendingRecord complete the pending record
func (dao *SmsSaleMobileNumberDAO) ConfirmPendingRecord(smsSaleId uint64, smsSaleMobileNumberIds []uint64) error {

	// Construct placeholders for the IN clause
	placeholders := make([]string, len(smsSaleMobileNumberIds))
	args := make([]interface{}, len(smsSaleMobileNumberIds)+3)

	// Fill placeholders and args
	for i, v := range smsSaleMobileNumberIds {
		placeholders[i] = fmt.Sprintf("$%d", i+4) // Start from $4
		args[i+3] = v                             // Add id to the args slice
	}
	// Add fixed query arguments for status and updated_by
	args[0] = models.SmsSaleMobileNumberStatusConfirmed
	args[1] = defined.DefaultUpdatedBy
	args[2] = smsSaleId

	query := fmt.Sprintf(`
		UPDATE sms_sale_mobile_number
		SET status = $1, updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT, updated_by = $2
		WHERE sms_sale_id = $3 AND id IN (%s)`, strings.Join(placeholders, ", "))

	_, err := dao.DB.Exec(query, args...)
	return err
}

// CompleteRecordByIds complete the record by ids
func (dao *SmsSaleMobileNumberDAO) CompleteRecordByIds(ids []uint64) error {

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

// UpdateNextRetryAt update completed
func (dao *SmsSaleMobileNumberDAO) UpdateNextRetryAt(smsSaleId uint64, nextRetryAt int64) error {

	query := fmt.Sprintf(`
		UPDATE sms_sale_mobile_number
		SET updated_at = EXTRACT(EPOCH FROM NOW())::BIGINT, updated_by = $1, retry_count = retry_count+1, next_retry_at = $2
		WHERE sms_sale_id = $3 AND status = $4`)

	_, err := dao.DB.Exec(query, defined.DefaultUpdatedBy, nextRetryAt, smsSaleId, models.SmsSaleMobileNumberStatusCompleted)
	return err
}
