// See LICENSE.txt for licensing information.

package queue

import (
	"encoding/json"
	"time"

	"github.com/iwanbk/gobeanstalk"
	"omapp/vendor/env"
)

type JobType int

const (
	MAP JobType = iota
	CLEAN
	DELETE
)

type Message struct {
	Type   JobType
	Target int
}

var Conn *gobeanstalk.Conn

func SendMsg(m Message, pri uint32, delay time.Duration) (uint64, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return 0, err
	}
	jid, err := Conn.Put(data, pri, delay, 24*time.Hour)
	if err != nil {
		return 0, err
	}
	return jid, nil
}

func Send(t JobType, target int, pri uint32, delay time.Duration) (uint64, error) {
	return SendMsg(Message{t, target}, pri, delay)
}

func Recv() (*Message, uint64, error) {
	job, err := Conn.Reserve()
	if err != nil {
		return nil, 0, err
	}
	var msg Message
	if err = json.Unmarshal(job.Body, &msg); err != nil {
		return nil, job.ID, err
	}
	return &msg, job.ID, nil
}

func Init() error {
	var err error
	mount := env.StringDefault("OMA_Q_MOUNT", "127.0.0.1:11300")
	Conn, err = gobeanstalk.Dial(mount)
	if err != nil {
		return err
	}
	name := env.StringDefault("OMA_Q_TUBE", "oma")
	if err := Conn.Use(name); err != nil {
		return err
	}
	if _, err := Conn.Watch(name); err != nil {
		return err
	}
	return nil
}
