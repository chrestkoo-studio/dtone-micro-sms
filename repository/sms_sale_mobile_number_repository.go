package repository

import (
	"database/sql"
	"dtone-micro-sms/dao"
	"dtone-micro-sms/models"
)

type SmsSaleMobileNumberRepository struct {
	SmsSaleMobileNumberDAO *dao.SmsSaleMobileNumberDAO
}

func NewSmsSaleMobileNumberRepository(smsLogDao *dao.SmsSaleMobileNumberDAO) *SmsSaleMobileNumberRepository {
	return &SmsSaleMobileNumberRepository{SmsSaleMobileNumberDAO: smsLogDao}
}

func (repo *SmsSaleMobileNumberRepository) Create(tx *sql.Tx, data *models.SmsSaleMobileNumber) (uint64, error) {
	return repo.SmsSaleMobileNumberDAO.Create(tx, data)
}

func (repo *SmsSaleMobileNumberRepository) GetPendingRecordBySmsSaleID(smsSaleId uint64) ([]*models.SmsSaleMobileNumber, error) {
	return repo.SmsSaleMobileNumberDAO.GetPendingRecordBySmsSaleID(smsSaleId)
}

func (repo *SmsSaleMobileNumberRepository) CompletePendingRecordByIds(ids []uint64) error {
	return repo.SmsSaleMobileNumberDAO.CompletePendingRecordByIds(ids)
}
