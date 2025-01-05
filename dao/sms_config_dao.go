package dao

import (
	"context"
	"database/sql"
	"dtone-micro-sms/models"
	"fmt"
)

type SmsConfigDAO struct {
	DB *sql.DB
}

func NewSmsConfigDAO(db *sql.DB) *SmsConfigDAO {
	return &SmsConfigDAO{
		DB: db,
	}
}

// Create inserts a new sms config record
func (dao *SmsConfigDAO) Create(data *models.SmsConfig) (uint32, error) {
	query := `
		INSERT INTO sms_log (country_id, mobile_prefix, cost, status, created_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := dao.DB.QueryRow(query, data.ID, data.CountryId, data.MobilePrefix, data.Cost, data.Status, data.CreatedAt, data.CreatedBy).
		Scan(&data.ID)
	return data.ID, err
}

// GetActiveSmsConfigs retrieve active sms configs
func (dao *SmsConfigDAO) GetActiveSmsConfigs() ([]*models.SmsConfig, error) {
	query := `
		SELECT id, country_id, mobile_prefix, cost, status, created_at, created_by, updated_at, updated_by 
		FROM sms_config 
		WHERE status = $1
	`
	rows, err := dao.DB.Query(query, models.SmsConfigStatusActive)
	if err != nil {
		return nil, fmt.Errorf("error querying active SMS configs: %w", err)
	}
	defer rows.Close()

	var list []*models.SmsConfig
	for rows.Next() {
		var row models.SmsConfig
		err = rows.Scan(
			&row.ID,
			&row.CountryId,
			&row.MobilePrefix,
			&row.Cost,
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

// GetActiveSmsConfigByID retrieve active sms config by id
func (dao *SmsConfigDAO) GetActiveSmsConfigByID(ctx context.Context, id uint32) (*models.SmsConfig, error) {
	query := `
		SELECT id, country_id, mobile_prefix, cost, status, created_at, created_by, updated_at, updated_by
		FROM sms_config
		WHERE id = $1 AND status = $2
	`
	row := dao.DB.QueryRowContext(ctx, query, id, models.SmsConfigStatusActive)

	var smsConfig models.SmsConfig
	if err := row.Scan(
		&smsConfig.ID,
		&smsConfig.CountryId,
		&smsConfig.MobilePrefix,
		&smsConfig.Cost,
		&smsConfig.Status,
		&smsConfig.CreatedAt,
		&smsConfig.CreatedBy,
		&smsConfig.UpdatedAt,
		&smsConfig.UpdatedBy,
	); err != nil {
		return nil, err
	}
	return &smsConfig, nil
}

func (dao *SmsConfigDAO) UpdateById(ctx context.Context, id uint32, data *models.SmsConfig) error {
	query := `
		UPDATE sms_config
		SET country_id = $1,
		    mobile_prefix = $2,
		    cost = $3,
		    status = $4,
		    updated_at = $5,
		    updated_by = $6
		WHERE id = $7
	`
	_, err := dao.DB.ExecContext(ctx, query,
		data.CountryId,
		data.MobilePrefix,
		data.Cost,
		data.Status,
		data.UpdatedAt,
		data.UpdatedBy,
		id,
	)
	return err
}
