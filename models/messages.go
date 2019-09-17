package models

type Message struct {
	Pk        int    `storm:"id,index,increment"`
	Timestamp int    `json:"ts" storm:"index"`
	Content   string `json:"msg"`
}
