package main

import (
	"time"
)

type message struct {
	UserID    string
	FullName  string
	Message   string
	When      time.Time
	AvatarURL string
}
