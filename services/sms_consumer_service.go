package services

import (
	"context"
	"dtone-base-sms/config"
	"dtone-base-sms/custom"
	"dtone-base-sms/defined"
	"dtone-std-library/json"
	"dtone-std-library/kafka"
	"dtone-std-library/logs"
	"dtone-std-library/redis"
	"fmt"
	"time"
)

// SameKfkMsgStruct 注单数据
type SameKfkMsgStruct struct {
	TraceId   string `json:"traceId"`
	CreatedAt int64  `json:"createdAt"`
}

type SmsKafkaData struct {
	SmsSaleId uint64
}

// initiate mysql consumer
func (service *KafkaService) startSmsDataConsumer(ctx context.Context, msg kafka.Message, broker []string, manualCommit bool) {
	if msg.Value == nil {
		return
	}

	logMsgTemplate := "[sms_consumer_service][startSmsDataConsumer]"
	kafkaInfo := kafkaLogger(ctx, msg)
	kfkMsg := fmt.Sprintf("[Broker:%v][GroupId:%s][Alias:%v][Partition:%v][Topic:%v][Time:%v][Offset:%v]",
		broker, fmt.Sprintf("%s", ctx.Value("ConsumerGroupId")), fmt.Sprintf("%s", ctx.Value("Alias")), msg.Partition, msg.Topic, msg.Time, msg.Offset)
	traceId := custom.GetLogId(ctx)
	redisKfkKey := fmt.Sprintf("[GroupId=%s][Alias=%v][Partition=%v][Topic=%v][Time=%v][Offset=%v]",
		fmt.Sprintf("%s", ctx.Value("ConsumerGroupId")), fmt.Sprintf("%s", ctx.Value("Alias")), msg.Partition, msg.Topic, msg.Time.UnixNano(), msg.Offset)
	logs.WithCtx(ctx).Info("%s[KafkaOnReceived][kfkMsg:%v][redisKfkKey:%v]", logMsgTemplate, kfkMsg, redisKfkKey)

	flowFrom := "kafka"
	ctx = context.WithValue(ctx, "kfkMsg", kfkMsg)
	ctx = context.WithValue(ctx, "flowFrom", flowFrom)

	// start redis prevent duplicate run same kfk msg
	lockKfkMsgPeriodSecond := 10 * time.Second
	if config.GetConfig().MicroSmsProcConfig.LockKfkMsgPeriodSecond > 0 {
		lockKfkMsgPeriodSecond = time.Duration(config.GetConfig().MicroSmsProcConfig.LockKfkMsgPeriodSecond) * time.Second
	}
	redisSameKfkMsgName := defined.Gen(defined.CacheSmsConsumer, "Sms", "SameKfkMsg", redisKfkKey)
	redisSameKfkMsgRst, err := redis.RDB().SetNx(redisSameKfkMsgName, string(json.Stringify(SameKfkMsgStruct{TraceId: traceId, CreatedAt: time.Now().UnixNano()})), lockKfkMsgPeriodSecond)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[redisSetNxError][err:%s][redisSameKfkMsgName:%s][kfkMsg:%s]", logMsgTemplate, err, redisSameKfkMsgName, kfkMsg)
		return
	}
	if !redisSameKfkMsgRst {
		logs.WithCtx(ctx).Info("%s[duplicatedSameKfkMsg][redisSameKfkMsgName:%s][kfkMsg:%s]", logMsgTemplate, redisSameKfkMsgName, kfkMsg)
		return
	}
	// end redis prevent duplicate run same kfk msg

	canCommit := service.processSmsConsumerData(ctx, msg.Value, kafkaInfo)
	if !canCommit {
		logs.WithCtx(ctx).Error("%s[consumedError][msgValue:%s][kafkaInfo:%s]", logMsgTemplate, msg.Value, kafkaInfo)
		return
	}
	if manualCommit {
		err = msg.Commit()
		if err != nil {
			logs.WithCtx(ctx).Error("%s[CommitError][err:%v]", logMsgTemplate, err) // logically commit failed should delete redis key also
		}
	}
	redis.RDB().Del(redisSameKfkMsgName) // logically commit failed should delete redis key also
}

func (service *KafkaService) processSmsConsumerData(ctx context.Context, msg []byte, ext string) bool {
	msgTemplate := "[sms_consumer_service][processSmsConsumerData]"
	consumedAt := time.Now()

	var record SmsKafkaData
	err := json.ParseE(msg, &record)
	if err != nil {
		logs.WithCtx(ctx).Error("%v[json_Unmarshal_error][err:%v][msg:%s]", msgTemplate, err, msg)
		return false
	}
	logs.WithCtx(ctx).Info("%v[start][ext:%v][msg:%s]", msgTemplate, ext, msg)

	processSendSmsTimeAt := time.Now()
	smsService := NewSmsService(service.SmsConfigRepo, service.SmsSaleRepo, service.SmsSaleMobileNumberRepo, service.SmsLogRepository)
	err = smsService.ProcessSendSms(ctx, record.SmsSaleId)
	if err != nil {
		logs.WithCtx(ctx).Error("%s[ProcessSendSms][err:%v][SmsSaleId:%d]", msgTemplate, err, record.SmsSaleId)
		return false
	}
	timeTakenProcessSendSmsTimeAt := time.Since(processSendSmsTimeAt).String()
	timeFinishAll := time.Since(consumedAt).String()

	logs.WithCtx(ctx).Info("%v[FinishAll]%v", msgTemplate, fmt.Sprintf("[timeTakenProcessSendSmsTimeAt:%s][timeFinishAll:%s]", timeTakenProcessSendSmsTimeAt, timeFinishAll))
	return true
}
