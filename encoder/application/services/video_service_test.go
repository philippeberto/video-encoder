package services_test

import (
	"encoder/application/repositories"
	"encoder/application/services"
	"encoder/domain"
	"encoder/framework/database"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func init() {
	err := godotenv.Load("../../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func prepare() (*domain.Video, repositories.VideoRepositoryDb) {
	db := database.NewDbTest()
	defer db.Close()
	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.ResourceID = "abc"
	video.FilePath = os.Getenv("FILE_NAME")
	video.CreatedAt = time.Now()
	repo := repositories.VideoRepositoryDb{Db: db}
	return video, repo
}

func TestServiceDownoad(t *testing.T) {
	video, repo := prepare()
	videoRepo := repositories.VideoRepositoryDb{Db: repo.Db}
	videoService := services.NewVideoService()
	videoService.Video = video
	videoService.VideoRepository = videoRepo
	err := videoService.Download(os.Getenv("INPUT_BUCKET_NAME"))
	require.Nil(t, err)
	err = videoService.Fragment()
	require.Nil(t, err)
	err = videoService.Encode()
	require.Nil(t, err)
	err = videoService.Finish()
	require.Nil(t, err)
}
