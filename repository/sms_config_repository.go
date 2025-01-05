package repository

import (
	"context"
	"dtone-base-sms/custom"
	"dtone-micro-sms/dao"
	"dtone-micro-sms/models"
	"dtone-std-library/containers/maps"
	"dtone-std-library/json"
	"dtone-std-library/logs"
	"fmt"
)

type SmsConfigRepository struct {
	DAO *dao.SmsConfigDAO
}

var SmsConfigRepoCache maps.Inf[int, *models.SmsConfig]

func NewSmsConfigRepository(dao *dao.SmsConfigDAO, list []*models.SmsConfig) *SmsConfigRepository {
	repo := &SmsConfigRepository{DAO: dao}
	if len(list) > 0 { // this is for testing purposes.
		SmsConfigRepoCache = maps.New[int, *models.SmsConfig](new(maps.RWMutex[int, *models.SmsConfig]))
		for _, v := range list {
			SmsConfigRepoCache.Set(v.MobilePrefix, v)
		}
	}
	return repo
}

func (repo *SmsConfigRepository) Create(data *models.SmsConfig) (uint32, error) {
	return repo.DAO.Create(data)
}

func (repo *SmsConfigRepository) Init(ctx context.Context) error {
	logMsgTemplate := "[SmsConfigRepository][Init]"
	err := repo.ReloadAllActive(ctx)
	if err != nil {
		logs.WithCtx(ctx).Error(logMsgTemplate, err)
		return err
	}
	repo.Print(ctx)
	return nil
}

func (repo *SmsConfigRepository) ReloadById(ctx context.Context, id uint32) error {
	logMsgTemplate := "[SmsConfigRepository][Reload]"
	result, err := repo.DAO.GetActiveSmsConfigByID(ctx, id)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[repo.DAO.GetSmsConfigByID][err:%v]", logMsgTemplate, err)
	}

	SmsConfigRepoCache.Set(result.MobilePrefix, result)

	return nil
}

func (repo *SmsConfigRepository) ReloadAllActive(ctx context.Context) error {
	logMsgTemplate := "[SmsConfigRepository][ReloadAllActive]"
	ctx = custom.SetLogIdIfNotExist(ctx, "")

	list, err := repo.DAO.GetActiveSmsConfigs()
	if err != nil {
		logs.WithCtx(ctx).Error("%s[repo.DAO.GetActiveSmsConfigs][err:%v]", logMsgTemplate, err)
		return err
	}
	if len(list) < 1 {
		logs.WithCtx(ctx).Error("%s[repo.DAO.GetActiveSmsConfigs][err:NoActiveSmsConfigs]", logMsgTemplate)
		return err
	}
	SmsConfigRepoCache = maps.New[int, *models.SmsConfig](new(maps.RWMutex[int, *models.SmsConfig]))
	for _, v := range list {
		SmsConfigRepoCache.Set(v.MobilePrefix, v)
	}
	return nil
}

func (repo *SmsConfigRepository) Print(ctx context.Context) {
	logMsgTemplate := "[SmsConfigRepository][Print]"
	ctx = custom.SetLogIdIfNotExist(ctx, "")
	fmt.Println(SmsConfigRepoCache.Len())
	SmsConfigRepoCache.Range(func(k int, v *models.SmsConfig) error {
		logs.WithCtx(ctx).Info("%s[k:%v][v:%v]", logMsgTemplate, k, json.String(v))
		return nil
	})
	logs.WithCtx(ctx).Info("%s[done][total:%d]", logMsgTemplate, SmsConfigRepoCache.Len())
}

func (repo *SmsConfigRepository) UpdateById(ctx context.Context, id uint32, data *models.SmsConfig) error {
	logMsgTemplate := fmt.Sprintf("[SmsConfigRepository][UpdateById][id:%d][data:%v]", id, json.String(data))

	// Step 1: Update the database
	err := repo.DAO.UpdateById(ctx, id, data)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[repo.DAO.UpdateById][err:%v]", logMsgTemplate, err)
		return err
	}

	// Step 1: Reload the cache by id
	err = repo.ReloadById(ctx, id)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[repo.ReloadById][err:%v]", logMsgTemplate, err)
		return err
	}

	return nil
}
