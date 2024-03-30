package repositories_test

import (
	"encoder/application/repositories"
	"encoder/domain"
	"encoder/framework/database"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestNewVideoRepository(test *testing.T) {
	test.Run("Creating new test database", func(t *testing.T) {
		db := database.NewDbTest()
		defer db.Close()
		video := domain.NewVideo()
		video.ID = uuid.NewV4().String()
		video.ResourceID = "abc"
		video.FilePath = "video.mp4"
		video.CreatedAt = time.Now()
		repo := repositories.VideoRepositoryDb{Db: db}
		repo.Insert(video)
		v, err := repo.Find(video.ID)
		if err != nil {
			test.Errorf("Error: %v", err)
		}
		require.NotEmpty(test, video.ID)
		require.Equal(test, v.ID, video.ID)
		require.Nil(test, err)
	})
}
