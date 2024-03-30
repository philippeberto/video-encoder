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

func TestJobRepositoryDbInsert(t *testing.T) {
	t.Run("Inserts a job on database", func(t *testing.T) {
		db := database.NewDbTest()
		defer db.Close()
		video := domain.NewVideo()
		video.ID = uuid.NewV4().String()
		video.ResourceID = "abc"
		video.FilePath = "video.mp4"
		video.CreatedAt = time.Now()
		repo := repositories.VideoRepositoryDb{Db: db}
		repo.Insert(video)
		job, error := domain.NewJob("output_path", "pending", video)
		require.Nil(t, error)
		repoJob := repositories.JobRepositoryDb{Db: db}
		repoJob.Insert(job)
		j, err := repoJob.Find(job.ID)
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		require.NotEmpty(t, j.ID)
		require.Equal(t, j.ID, job.ID)
		require.Equal(t, j.VideoID, video.ID)
		require.Nil(t, err)
	})
	t.Run("Updates a job on database", func(t *testing.T) {
		db := database.NewDbTest()
		defer db.Close()
		video := domain.NewVideo()
		video.ID = uuid.NewV4().String()
		video.ResourceID = "abc"
		video.FilePath = "video.mp4"
		video.CreatedAt = time.Now()
		repo := repositories.VideoRepositoryDb{Db: db}
		repo.Insert(video)
		job, error := domain.NewJob("output_path", "pending", video)
		require.Nil(t, error)
		repoJob := repositories.JobRepositoryDb{Db: db}
		repoJob.Insert(job)
		job.Status = "Complete"
		repoJob.Update(job)
		j, err := repoJob.Find(job.ID)
		if err != nil {
			t.Errorf("Error: %v", err)
		}
		require.NotEmpty(t, j.ID)
		require.Equal(t, j.ID, job.ID)
		require.Equal(t, j.VideoID, video.ID)
		require.Equal(t, j.Status, "Complete")
		require.Nil(t, err)
	})
}
