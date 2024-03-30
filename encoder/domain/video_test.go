package domain_test

import (
	"encoder/domain"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestValidateIfVideoIsEmpty(t *testing.T) {
	video := domain.NewVideo()
	err := video.Validate()
	require.Error(t, err)
}

func TestVideoIDNotUUID(t *testing.T) {
	video := domain.NewVideo()
	video.ID = "abc"
	video.ResourceID = "abc"
	video.CreatedAt = time.Now()
	video.FilePath = "./video.mp4"
	err := video.Validate()
	require.Error(t, err)
}

func TestValidNewVideo(t *testing.T) {
	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.ResourceID = "abc"
	video.CreatedAt = time.Now()
	video.FilePath = "./video.mp4"
	err := video.Validate()
	require.Nil(t, err)
}
