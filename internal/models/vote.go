package models

import "time"

type Vote struct {
	Nickname  string
	Voice     int32
	Thread    int32
	UpdatedAt time.Time
}
