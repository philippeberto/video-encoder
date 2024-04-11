package services

import (
	"encoder/application/repositories"
	"encoder/domain"
	"encoder/framework/queue"
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

type JobManager struct {
	Db               *gorm.DB
	Domain           domain.Job
	MessageChannel   chan amqp.Delivery
	JobReturnChannel chan JobWorkerResult
	RabbitMQ         *queue.RabbitMQ
}

type JobNotificationError struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func NewJobManager(
	db *gorm.DB,
	msgChannel chan amqp.Delivery,
	jobReturnChannel chan JobWorkerResult,
	rabbitMQ *queue.RabbitMQ,
) *JobManager {
	return &JobManager{
		Db:               db,
		Domain:           domain.Job{},
		MessageChannel:   msgChannel,
		JobReturnChannel: jobReturnChannel,
		RabbitMQ:         rabbitMQ,
	}
}

func (job *JobManager) Start(channel *amqp.Channel) {
	videoService := NewVideoService()
	videoService.VideoRepository = repositories.VideoRepositoryDb{Db: job.Db}
	jobService := JobService{
		JobRepository: repositories.JobRepositoryDb{Db: job.Db},
		VideoService:  videoService,
	}
	concurrency, err := strconv.Atoi(os.Getenv("CONCURRENCY_WORKERS"))
	if err != nil {
		log.Fatalf("Error parsing CONCURRENCY_WORKERS")
	}
	for qtdProcesses := 0; qtdProcesses < concurrency; qtdProcesses++ {
		go JobWorker(job.Domain, job.MessageChannel, jobService, qtdProcesses, job.JobReturnChannel)
	}
	for jobResult := range job.JobReturnChannel {
		if jobResult.Error != nil {
			err = job.notifyFailure(jobResult)
		} else {
			err = job.notifySuccess(jobResult, channel)
		}
		if err != nil {
			jobResult.Message.Reject(false)
		}
	}
}

func (job *JobManager) notify(jobJson []byte) error {
	err := job.RabbitMQ.Notify(
		string(jobJson),
		"application/json",
		os.Getenv("RABBITMQ_NOTIFICATION_EX"),
		os.Getenv("RABBITMQ_NOTIFICATION_ROUTING_KEY"),
	)
	if err != nil {
		return err
	}
	return nil
}

func (job *JobManager) notifySuccess(jobResult JobWorkerResult, channel *amqp.Channel) error {
	Mutex.Lock()
	jobJson, err := json.Marshal(jobResult.Job)
	Mutex.Unlock()
	if err != nil {
		return err
	}
	err = job.notify(jobJson)
	if err != nil {
		return err
	}
	err = jobResult.Message.Ack(false)
	if err != nil {
		return err
	}
	return nil
}

func (job *JobManager) notifyFailure(jobResult JobWorkerResult) error {
	if jobResult.Job.ID != "" {
		log.Printf("MessageID: %v. Error with job: %v, with video %v. Error: %v",
			jobResult.Message.DeliveryTag,
			jobResult.Job.ID,
			jobResult.Job.Video.ID,
			jobResult.Error.Error())
	} else {
		log.Printf("MessageID: %v. Error persing message: %v", jobResult.Message.DeliveryTag, jobResult.Error.Error())
	}
	errorMesg := JobNotificationError{
		Message: string(jobResult.Message.Body),
		Error:   jobResult.Error.Error(),
	}
	jobJson, err := json.Marshal(errorMesg)
	if err != nil {
		return err
	}
	err = job.notify(jobJson)
	if err != nil {
		return err
	}
	err = jobResult.Message.Reject(false)
	if err != nil {
		return err
	}
	return nil
}
