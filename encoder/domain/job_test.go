package domain_test

import (
	"encoder/domain"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestNewJob(t *testing.T) {
	t.Run("Should return an error when job is empty", func(t *testing.T) {
		job, err := domain.NewJob("", "", nil)
		require.Error(t, err)
		require.Nil(t, job)
	})

	t.Run("Should generate a job ID when job is created", func(t *testing.T) {
		video := domain.NewVideo()
		video.ID = uuid.NewV4().String()
		video.ResourceID = "abc"
		video.CreatedAt = time.Now()
		video.FilePath = "video.mp4"

		job, err := domain.NewJob("output_path", "WAITING", video)
		require.Nil(t, err)
		require.NotNil(t, job)
		require.NotEmpty(t, job.ID)
	})

	t.Run("Should validate when job is created", func(t *testing.T) {
		video := domain.NewVideo()
		video.ID = uuid.NewV4().String()
		video.ResourceID = "abc"
		video.CreatedAt = time.Now()
		video.FilePath = "video.mp4"

		job, err := domain.NewJob("output_path", "WAITING", video)
		require.Nil(t, err)
		require.Nil(t, job.Validate())
	})
}
