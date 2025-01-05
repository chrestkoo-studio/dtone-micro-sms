package dao

import (
	"database/sql"
	"dtone-micro-sms/models"
	"errors"
)

type SmsLogDAO struct {
	DB *sql.DB
}

func NewSmsLogDAO(db *sql.DB) *SmsLogDAO {
	return &SmsLogDAO{
		DB: db,
	}
}

// Create inserts a new sms log record
func (dao *SmsLogDAO) Create(smsLog *models.SmsLog) (uint64, error) {
	query := `
		INSERT INTO sms_log (partner_id, mobile_no, url, message, input, retry_count, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := dao.DB.QueryRow(query, smsLog.PartnerID, smsLog.MobileNo, smsLog.Url, smsLog.Message, smsLog.Input, 0, smsLog.Status, smsLog.CreatedAt).
		Scan(&smsLog.ID)
	return smsLog.ID, err
}

// GetByID retrieves a SmsLog by ID
func (dao *SmsLogDAO) GetByID(id uint64) (*models.SmsLog, error) {
	var m models.SmsLog
	query := "SELECT id, partner_id, mobile_no, url, message, input, retry_count, status, updated_at FROM sms_log WHERE id = $1"
	err := dao.DB.QueryRow(query, id).Scan(
		&m.ID, &m.PartnerID, &m.MobileNo, &m.Url,
		&m.Input, &m.RetryCount,
		&m.Status, &m.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.New("sms log not found")
	}
	return &m, err
}

// GetPendingSms retrieves a pending SmsLog by ID
func (dao *SmsLogDAO) GetPendingSms(id uint64) (*models.SmsLog, error) {
	var m models.SmsLog
	query := "SELECT id, partner_id, mobile_no, url, message, input, retry_count, status, updated_at FROM sms_log WHERE id = $1 and status = $2"
	err := dao.DB.QueryRow(query, id, models.PartnerStatusPending).Scan(
		&m.ID, &m.PartnerID, &m.MobileNo, &m.Url,
		&m.Input, &m.RetryCount,
		&m.Status, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("m not found")
		}
		return nil, err
	}

	return &m, err
}
