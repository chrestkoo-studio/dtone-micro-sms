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

func (repo *SmsSaleMobileNumberRepository) GetConfirmedRecordBySmsSaleID(smsSaleId uint64) ([]*models.SmsSaleMobileNumber, error) {
	return repo.SmsSaleMobileNumberDAO.GetConfirmedRecordBySmsSaleID(smsSaleId)
}

func (repo *SmsSaleMobileNumberRepository) GetUniqueRetryConfirmedRecords(nextRetryAt int64) ([]uint64, error) {
	return repo.SmsSaleMobileNumberDAO.GetUniqueRetryConfirmedRecords(nextRetryAt)
}

func (repo *SmsSaleMobileNumberRepository) GetPendingRecordByIdSmsSaleID(smsSaleId uint64, ids []uint64) ([]*models.SmsSaleMobileNumber, error) {
	return repo.SmsSaleMobileNumberDAO.GetPendingRecordByIdSmsSaleID(smsSaleId, ids)
}

func (repo *SmsSaleMobileNumberRepository) ConfirmPendingRecord(smsSaleId uint64, smsSaleMobileNumberIds []uint64) error {
	return repo.SmsSaleMobileNumberDAO.ConfirmPendingRecord(smsSaleId, smsSaleMobileNumberIds)
}

func (repo *SmsSaleMobileNumberRepository) CompleteRecordByIds(ids []uint64) error {
	return repo.SmsSaleMobileNumberDAO.CompleteRecordByIds(ids)
}

func (repo *SmsSaleMobileNumberRepository) UpdateNextRetryAt(smsSaleId uint64, nextRetryAt int64) error {
	return repo.SmsSaleMobileNumberDAO.UpdateNextRetryAt(smsSaleId, nextRetryAt)
}
