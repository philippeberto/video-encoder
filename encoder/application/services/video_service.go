package services

import (
	"context"
	"encoder/application/repositories"
	"encoder/domain"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"cloud.google.com/go/storage"
)

type VideoService struct {
	Video           *domain.Video
	VideoRepository repositories.VideoRepository
}

func NewVideoService() VideoService {
	return VideoService{}
}

func (v *VideoService) Download(bucketName string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	bucket := client.Bucket(bucketName)
	object := bucket.Object(v.Video.FilePath)
	reader, err := object.NewReader(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	newFile, err := os.Create(os.Getenv("LOCAL_STORAGE_PATH") + "/" + v.Video.ID + ".mp4")
	if err != nil {
		return err
	}
	newFile.Write(body)
	defer newFile.Close()
	log.Printf("Video %v has been downloaded", v.Video.ID)

	return nil
}

func (v *VideoService) Fragment() error {
	err := os.Mkdir(os.Getenv("LOCAL_STORAGE_PATH")+"/"+v.Video.ID, os.ModePerm)
	if err != nil {
		return err
	}
	source := os.Getenv("LOCAL_STORAGE_PATH") + "/" + v.Video.ID + ".mp4"
	target := os.Getenv("LOCAL_STORAGE_PATH") + "/" + v.Video.ID + ".frag"
	cmd := exec.Command("mp4fragment", source, target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	printOutput(output)
	return nil
}

func (v *VideoService) Encode() error {
	cmdArgs := []string{}
	cmdArgs = append(cmdArgs, os.Getenv("LOCAL_STORAGE_PATH")+"/"+v.Video.ID+".frag")
	cmdArgs = append(cmdArgs, "--use-segment-timeline")
	cmdArgs = append(cmdArgs, "-o")
	cmdArgs = append(cmdArgs, os.Getenv("LOCAL_STORAGE_PATH")+"/"+v.Video.ID)
	cmdArgs = append(cmdArgs, "-f")
	cmdArgs = append(cmdArgs, "--exec-dir")
	cmdArgs = append(cmdArgs, os.Getenv("BIN_PATH"))
	cmd := exec.Command("mp4dash", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	printOutput(output)
	return nil
}

// removes files from local storage
func (v *VideoService) Finish() error {
	err := os.Remove(os.Getenv("LOCAL_STORAGE_PATH") + "/" + v.Video.ID + ".mp4")
	if err != nil {
		log.Println("Error removing file: ", v.Video.ID+".mp4")
		return err
	}
	err = os.Remove(os.Getenv("LOCAL_STORAGE_PATH") + "/" + v.Video.ID + ".frag")
	if err != nil {
		log.Println("Error removing file: ", v.Video.ID+".frag")
		return err
	}
	err = os.RemoveAll(os.Getenv("LOCAL_STORAGE_PATH") + "/" + v.Video.ID)
	if err != nil {
		log.Println("Error removing folder: ", v.Video.ID)
		return err
	}
	log.Println("Files have been removed", v.Video.ID)
	return nil
}

func (v *VideoService) InsertVideo() error {
	_, err := v.VideoRepository.Insert(v.Video)
	if err != nil {
		return err
	}
	return nil
}

func printOutput(output []byte) {
	if len(output) > 0 {
		log.Printf("======> Output: %s\n", string(output))
	}
}
