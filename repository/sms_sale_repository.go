package repository

import (
	"database/sql"
	"dtone-micro-sms/dao"
	"dtone-micro-sms/models"
)

type SmsSaleRepository struct {
	SmsSaleDAO *dao.SmsSaleDAO
}

func NewSmsSaleRepository(smsLogDao *dao.SmsSaleDAO) *SmsSaleRepository {
	return &SmsSaleRepository{SmsSaleDAO: smsLogDao}
}

func (repo *SmsSaleRepository) Create(tx *sql.Tx, data *models.SmsSale) (uint64, error) {
	return repo.SmsSaleDAO.Create(tx, data)
}

func (repo *SmsSaleRepository) ConfirmPendingRecordById(id uint64) error {
	return repo.SmsSaleDAO.ConfirmPendingRecordById(id)
}

func (repo *SmsSaleRepository) CompleteRecordById(id uint64) error {
	return repo.SmsSaleDAO.CompleteRecordById(id)
}

/*
func (repo *SmsSaleRepository) GetActivePartnerByUsername(id uint64) (*models.SmsSale, error) {
	return repo.SmsSaleDAO.GetPendingSms(id)
}
*/
