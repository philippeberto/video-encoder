package domain

import (
	"time"

	"github.com/asaskevich/govalidator"
)

type Video struct {
	ID         string    `json:"encoded_video_folder" valid:"uuidv4" gorm:"type:uuid;primary_key;"`
	ResourceID string    `json:"resource_id" valid:"notnull" gorm:"type:varchar(255);"`
	FilePath   string    `json:"file_path" valid:"notnull" gorm:"type:varchar(255);"`
	CreatedAt  time.Time `json:"-" valid:"-"`
	Jobs       []*Job    `json:"-" valid:"-" gorm:"ForeignKey:VideoID;"`
}

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}

func NewVideo() *Video {
	return &Video{}
}

func (video *Video) Validate() error {
	valid, err := govalidator.ValidateStruct(video)
	if !valid {
		return err
	}
	return nil
}
