package service

import (
	"github.com/4040www/NativeCloud_HR/internal/model"
	"github.com/4040www/NativeCloud_HR/internal/repository"
)

func Clock(req *model.CheckInRequest) error {

	return repository.CreateCheckinRecord(req)

}
