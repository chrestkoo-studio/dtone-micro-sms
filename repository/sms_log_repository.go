package repository

import (
	"dtone-micro-sms/dao"
	"dtone-micro-sms/models"
)

type SmsLogRepository struct {
	SmsLogDAO    *dao.SmsLogDAO
}

func NewSmsLogRepository(smsLogDao *dao.SmsLogDAO) *SmsLogRepository {
	return &SmsLogRepository{SmsLogDAO: smsLogDao}
}

func (repo *SmsLogRepository) Create(partner *models.SmsLog) (uint64, error) {
	return repo.SmsLogDAO.Create(partner)
}

func (repo *SmsLogRepository) GetPartnerByUsername(id uint64) (*models.SmsLog, error) {
	return repo.SmsLogDAO.GetByID(id)
}

func (repo *SmsLogRepository) GetActivePartnerByUsername(id uint64) (*models.SmsLog, error) {
	return repo.SmsLogDAO.GetPendingSms(id)
}
