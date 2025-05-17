package service

import (
	messagequeue "github.com/4040www/NativeCloud_HR/internal/messageQueue"
	"github.com/4040www/NativeCloud_HR/internal/model"
)

func Clock(req *model.CheckInRequest) error {
	return messagequeue.SendCheckIn(req)
}
