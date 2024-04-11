package services

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"cloud.google.com/go/storage"
)

type VideoUpload struct {
	Paths        []string
	VideoPath    string
	OutputBucket string
	Errors       []string
}

func NewVideoUpload() *VideoUpload {
	return &VideoUpload{}
}

func (videoUpload *VideoUpload) UploadObject(objectPath string, client *storage.Client, ctx context.Context) error {
	path := strings.Split(objectPath, os.Getenv("LOCAL_STORAGE_PATH")+"/")
	f, err := os.Open(objectPath)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := client.Bucket(videoUpload.OutputBucket).Object(path[1]).NewWriter(ctx)
	// writer.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	if _, err := io.Copy(writer, f); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}

func (videoUpoad *VideoUpload) ProcessUpload(concurrency int, doneUpload chan string) error {
	input := make(chan int, runtime.NumCPU())
	returnChannel := make(chan string)
	err := videoUpoad.loadPaths()
	if err != nil {
		return err
	}
	uploadClient, ctx, err := getClientUpload()
	if err != nil {
		return err
	}
	for process := 0; process < concurrency; process++ {
		go videoUpoad.uploadWorker(input, returnChannel, uploadClient, ctx)
	}
	go func() {
		for x := 0; x < len(videoUpoad.Paths); x++ {
			input <- x
		}
		close(input)
	}()
	for r := range returnChannel {
		if r != "" {
			doneUpload <- r
			break
		}
	}

	return nil
}

func (videoUpload *VideoUpload) loadPaths() error {
	err := filepath.Walk(videoUpload.VideoPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			videoUpload.Paths = append(videoUpload.Paths, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func getClientUpload() (*storage.Client, context.Context, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	return client, ctx, nil
}

func (videoUpload *VideoUpload) uploadWorker(input chan int, returnChannel chan string, uploadClient *storage.Client, ctx context.Context) {
	for i := range input {
		err := videoUpload.UploadObject(videoUpload.Paths[i], uploadClient, ctx)
		if err != nil {
			videoUpload.Errors = append(videoUpload.Errors, videoUpload.Paths[i])
			log.Printf("Error during the upload: %v. Error: %v", videoUpload.Paths[i], err)
			returnChannel <- err.Error()
		}
		returnChannel <- ""
	}
	returnChannel <- "upload completed"
}
