package repository

import (
	"errors"

	"github.com/4040www/NativeCloud_HR/internal/db"
	"github.com/4040www/NativeCloud_HR/internal/model"
)

func CreateCheckinRecord(record *model.CheckInRequest) error {
	if record == nil {
		return errors.New("record is nil")
	}
	return db.GetDB().Create(record).Error
}
