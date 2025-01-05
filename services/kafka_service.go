package services

import (
	"context"
	"dtone-base-sms/config"
	"dtone-base-sms/custom"
	"dtone-base-sms/defined"
	"dtone-micro-sms/repository"
	"dtone-std-library/kafka"
	"dtone-std-library/logs"
	"fmt"
	kafka2 "github.com/segmentio/kafka-go"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	ActiveSmsConsumerSession sync.Map
	channelLimit             = 1 // default available channel at least must be one
	endRunning               = make(chan bool, 1)
)

type KafkaService struct {
	SmsConfigRepo           *repository.SmsConfigRepository
	SmsSaleRepo             *repository.SmsSaleRepository
	SmsSaleMobileNumberRepo *repository.SmsSaleMobileNumberRepository
	SmsLogRepository        *repository.SmsLogRepository
}

func NewKafkaService(
	smsConfigRepo *repository.SmsConfigRepository,
	smsSaleRepo *repository.SmsSaleRepository,
	smsSaleMobileNumberRepo *repository.SmsSaleMobileNumberRepository,
	smsLogRepository *repository.SmsLogRepository) *KafkaService {
	return &KafkaService{
		SmsConfigRepo:           smsConfigRepo,
		SmsSaleRepo:             smsSaleRepo,
		SmsSaleMobileNumberRepo: smsSaleMobileNumberRepo,
		SmsLogRepository:        smsLogRepository,
	}
}

func (service *KafkaService) InitConsumer(ctx context.Context) {
	service.initSmsKafkaConsumer(ctx)
}

// todo: initiate sql consumer
func InitiateSmsKafkaObserver(ctx context.Context) {
	ctx = custom.SetLogIdIfNotExist(ctx, "")
	msgTemplate := "[kafka_service][InitiateSmsKafkaObserver][%v]%v"
	logs.WithCtx(ctx).Info(msgTemplate, "start", "")
	select {
	case <-ctx.Done():
		err := ctx.Value("kfk_err")
		if err != nil {
			if val, ok := err.(kafka2.Error); ok {
				logs.WithCtx(ctx).Error(msgTemplate, "InitiateMysqlKafkaObserverKafka2Error", fmt.Sprintf("[valError:%v]", val.Error()))
			} else {
				logs.WithCtx(ctx).Error(msgTemplate, "InitiateMysqlKafkaObserverCtxError", fmt.Sprintf("[UnknownError%v]", err))
			}
		}
	}

}

func kafkaLogger(ctx context.Context, msg kafka.Message) string {
	ctx = custom.SetLogIdIfNotExist(ctx, "")
	msgTemplate := "[kafka_service][kafkaLogger][%v]%v"
	logs.WithCtx(ctx).Info(msgTemplate, "start", "")
	alias := ctx.Value("Alias")
	kafkaName := ctx.Value("KafkaName")
	kafkaIdentifier := ctx.Value("KafkaIdentifier")
	kafkaGroupId := ctx.Value("ConsumerGroupId")

	receiverHostName, err := os.Hostname()

	if err != nil {
		logs.WithCtx(ctx).Error(msgTemplate, "GetHostNameError", "[err:%v]", err)
		os.Exit(1)
	}

	if kafkaGroupId == nil {
		kafkaGroupId = "unknownGroupId"
	}

	kafkaInfo := fmt.Sprintf("%s %s %d %d %s", alias, msg.Topic, msg.Partition, msg.Offset, receiverHostName)
	if kafkaName != nil && kafkaIdentifier != nil {
		kafkaInfo = fmt.Sprintf("%s %s %s %s", kafkaName, kafkaIdentifier, kafkaInfo, kafkaGroupId)
	}

	return kafkaInfo
}

func (service *KafkaService) initSmsKafkaConsumer(ctx context.Context) {
	kafkaSms := defined.KafkaMicroSms
	logs.Info("initializing initSmsKafkaConsumer")

	mysqlGameKafkaCtxInfo := getKafkaContextInfo(kafkaSms, defined.GetKafkaMicroSmsTopic(kafkaSms))
	cloudCtx := initKafkaContext(ctx, mysqlGameKafkaCtxInfo)
	cloudCtx = context.WithValue(cloudCtx, logs.LogIdCode, custom.GenerateTraceId())

	//开始 mysql consumer 手续
	go service.startSmsConsumer(cloudCtx, kafkaSms, config.GetConfig().KafkaConsumer.BrokersAddrs, config.GetConfig().KafkaConsumer.Count, config.GetConfig().KafkaConsumer.ManualCommit)
	ActiveSmsConsumerSession.Store(mysqlGameKafkaCtxInfo.Alias, mysqlGameKafkaCtxInfo.Alias)
	go InitiateSmsKafkaObserver(cloudCtx)

	logs.WithCtx(cloudCtx).Info("[kafka_service][consumer][initMysqlGameKafkaUgs][InitiateSmsKafkaObserver][Broker:%v][GroupId:%s][Alias:%v][Topic:%v]",
		config.GetConfig().KafkaConsumer.BrokersAddrs, mysqlGameKafkaCtxInfo.ConsumerGroupId, mysqlGameKafkaCtxInfo.Alias, mysqlGameKafkaCtxInfo.Topic)

}

type SmsKafkaContext struct {
	ConsumerGroupId string
	KafkaName       string
	Alias           string
	Topic           string
}

func getKafkaContextInfo(kafkaSms string, topicId int) *SmsKafkaContext {
	return &SmsKafkaContext{
		ConsumerGroupId: fmt.Sprintf("consumer-group-%v", kafkaSms),
		KafkaName:       fmt.Sprintf("MicroSmsKafka%v", kafkaSms),
		Alias:           "micro-sms-alias-" + kafkaSms,
		Topic:           config.GetConfig().KafkaConsumer.Topics[topicId],
	}
}

func initKafkaContext(ctx context.Context, data *SmsKafkaContext) context.Context {
	ctx = custom.SetLogIdIfNotExist(ctx, "")
	ctx = context.WithValue(ctx, "KafkaIdentifier", strconv.Itoa(int(time.Now().UnixMilli())))
	ctx = context.WithValue(ctx, "KafkaName", data.KafkaName)
	ctx = context.WithValue(ctx, "ConsumerGroupId", data.ConsumerGroupId)
	ctx = context.WithValue(ctx, "Alias", data.Alias)
	ctx = context.WithValue(ctx, "Topic", data.Topic)
	ctx = context.WithValue(ctx, "mode", "sms")

	return ctx
}

func (service *KafkaService) startSmsConsumer(subCtx context.Context, kafkaSms string, broker []string, spawnCount int, manualCommit bool) {
	msgTemplate := "[kafka_service][startSmsConsumer]"
	for {
		ctx, cancel := context.WithCancel(subCtx)

		if config.GetConfig().MicroSmsProcConfig.ChannelLimit > 0 {
			channelLimit = config.GetConfig().MicroSmsProcConfig.ChannelLimit
		}
		c := make(chan func(ctx context.Context, msg kafka.Message, broker []string, manualCommit bool), channelLimit)
		for i := 0; i < channelLimit; i++ {
			switch kafkaSms {
			case defined.KafkaMicroSms:
				c <- service.startSmsDataConsumer
			default:
				logs.WithCtx(ctx).Error("%v[InvalidKafkaGameConsumer][kafkaSms:%v]", msgTemplate, kafkaSms)
			}
		}

		kafka.NewConsumer(ctx, &kafka.ConsumerOption{
			AliasName:    fmt.Sprintf("%s", ctx.Value("Alias")),
			Count:        spawnCount,
			BrokersAddrs: broker,
			GroupID:      fmt.Sprintf("%s", ctx.Value("ConsumerGroupId")),
			Topics:       []string{fmt.Sprintf("%s", ctx.Value("Topic"))},
			ManualCommit: manualCommit,
			MaxBytes:     10e6, // 10MB
			OnReceive: func(ctx context.Context, msg kafka.Message) {
				do := <-c
				go func(ctx context.Context, msg kafka.Message, broker []string, manualCommit bool) {
					do(ctx, msg, broker, manualCommit)
					c <- do
				}(ctx, msg, broker, manualCommit)
			},
			OnError: func(err error) {
				logs.WithCtx(ctx).Error("%v[KafkaError][err:%v]", msgTemplate, err)
				endRunning <- true
			},
		})

		if len(endRunning) == 0 {
			logs.WithCtx(ctx).Info("%v[Connection]%v", msgTemplate, fmt.Sprintf("Kafka 连接成功: Topic:<%v>, 启动Count:<%d>, Alias:<%v>, GroupID:<%v>, BrokersAddrs:<%#v>",
				[]string{fmt.Sprintf("%s", ctx.Value("Topic"))},
				spawnCount,
				ctx.Value("Alias"),
				fmt.Sprintf("%s", ctx.Value("ConsumerGroupId")),
				broker))
		}

		//deadlock if success
		<-endRunning

		cancel()

		kafka.Close(fmt.Sprintf("%s", ctx.Value("Alias")))

		// 如果出现了异常则尝试1分钟后等待
		logs.WithCtx(ctx).Error("%v[KafkaServiceError]%v", msgTemplate, fmt.Sprintf("Kafka服务 topic: %v 出现中断异常，正在重启！！！", fmt.Sprintf("%s-%s", ctx.Value("Alias"), ctx.Value("Topic"))))
		time.Sleep(time.Second * 60)

	}

	//common.KafkaConsumerAlias = append(common.KafkaConsumerAlias, fmt.Sprintf("%s", ctx.Value("Alias")))
}
