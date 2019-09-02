package models

import (
	"time"
)

type Message struct {
	Pk        int       `storm:"id,index,increment"`
	Timestamp time.Time `storm:"index" json:"ts"`
	Content   string    `json:msg`
}
