package repository

import (
	"github.com/4040www/NativeCloud_HR/internal/db"
	"github.com/4040www/NativeCloud_HR/internal/model"
)

func FindUserByEmail(email string) (*model.Employee, error) {
	var emp model.Employee
	if err := db.GetDB().Where("email = ?", email).First(&emp).Error; err != nil {
		return nil, err
	}
	return &emp, nil
}
