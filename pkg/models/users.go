package models

import (
	"time"
)

type User struct {
	Id        string `gorm:"primary_key;default:gen_random_uuid()"`
	FirstName string
	LastName  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
