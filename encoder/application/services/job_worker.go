package services

import (
	"encoder/domain"
	"encoder/framework/utils"
	"encoding/json"
	"os"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

type JobWorkerResult struct {
	Job     domain.Job
	Message *amqp.Delivery
	Error   error
}

var Mutex = &sync.Mutex{}

func JobWorker(
	job domain.Job,
	messageChannel chan amqp.Delivery,
	jobService JobService,
	workerID int,
	returnChannel chan JobWorkerResult,
) {
	for message := range messageChannel {
		err := utils.IsJson(string(message.Body))
		if err != nil {
			returnChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		Mutex.Lock()
		err = json.Unmarshal(message.Body, &jobService.VideoService.Video)
		if err != nil {
			returnChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		jobService.VideoService.Video.ID = uuid.NewV4().String()
		Mutex.Unlock()
		err = jobService.VideoService.Video.Validate()
		if err != nil {
			returnChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		Mutex.Lock()
		err = jobService.VideoService.InsertVideo()
		Mutex.Unlock()
		if err != nil {
			returnChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		job.Video = jobService.VideoService.Video
		job.OutputBucketPath = os.Getenv("OUTPUT_BUCKET_NAME")
		job.ID = uuid.NewV4().String()
		job.Status = "STARTING"
		job.CreatedAt = time.Now()
		Mutex.Lock()
		_, err = jobService.JobRepository.Insert(&job)
		Mutex.Unlock()
		if err != nil {
			returnChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		jobService.Job = &job
		err = jobService.Start()
		if err != nil {
			returnChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		returnChannel <- returnJobResult(job, message, nil)
	}

}

func returnJobResult(job domain.Job, message amqp.Delivery, err error) JobWorkerResult {
	return JobWorkerResult{
		Job:     job,
		Message: &message,
		Error:   err,
	}
}
