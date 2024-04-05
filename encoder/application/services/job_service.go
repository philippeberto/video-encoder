package services

import (
	"encoder/application/repositories"
	"encoder/domain"
	"errors"
	"os"
	"strconv"
)

type JobService struct {
	Job           *domain.Job
	JobRepository repositories.JobRepository
	VideoService  VideoService
}

func (job *JobService) Start() error {
	err := job.changeJobStatus("DOWNLOADING")
	if err != nil {
		return job.failJob(err)
	}
	err = job.VideoService.Download(os.Getenv("INPUT_BUCKET_NAME"))
	if err != nil {
		return job.failJob(err)
	}
	err = job.changeJobStatus("FRAGMENTING")
	if err != nil {
		return job.failJob(err)
	}
	err = job.VideoService.Fragment()
	if err != nil {
		return job.failJob(err)
	}
	err = job.changeJobStatus("ENCODING")
	if err != nil {
		return job.failJob(err)
	}
	err = job.VideoService.Encode()
	if err != nil {
		return job.failJob(err)
	}
	err = job.changeJobStatus("UPLOADING")
	if err != nil {
		return job.failJob(err)
	}
	err = job.performUpload()
	if err != nil {
		return job.failJob(err)
	}
	err = job.changeJobStatus("FINISHING")
	if err != nil {
		return job.failJob(err)
	}
	err = job.VideoService.Finish()
	if err != nil {
		return job.failJob(err)
	}
	err = job.changeJobStatus("COMPLETED")
	if err != nil {
		return job.failJob(err)
	}
	return nil
}

func (job *JobService) changeJobStatus(status string) error {
	var err error
	job.Job.Status = status
	job.Job, err = job.JobRepository.Update(job.Job)
	if err != nil {
		return job.failJob(err)
	}
	return nil
}

func (job *JobService) failJob(error error) error {
	job.Job.Status = "FAILED"
	job.Job.Error = error.Error()
	_, err := job.JobRepository.Update(job.Job)
	if err != nil {
		return err
	}
	return error
}

func (job *JobService) performUpload() error {
	err := job.changeJobStatus("UPLOADING")
	if err != nil {
		job.failJob(err)
	}
	videoUpload := NewVideoUpload()
	videoUpload.OutputBucket = os.Getenv("OUTPUT_BUCKET_NAME")
	videoUpload.VideoPath = os.Getenv("LOCAL_STORAGE_PATH") + "/" + job.VideoService.Video.ID
	concurrency, _ := strconv.Atoi(os.Getenv("CONCURRENCY_UPLOAD"))
	doneUpload := make(chan string)
	go videoUpload.ProcessUpload(concurrency, doneUpload)
	uploadResult := <-doneUpload
	if uploadResult != "upload completed" {
		job.failJob(errors.New(uploadResult))
	}
	return nil
}
