// See LICENSE.txt for licensing information.

package queue

import (
	"encoding/json"
	"time"

	"github.com/kr/beanstalk"
	"omapp/vendor/env"
)

type JobType int

const (
	MAP JobType = iota
	DEL
)

type Message struct {
	Type   JobType
	Target int
}

var (
	Conn *beanstalk.Conn
	Tube *beanstalk.Tube
)

func Send(t JobType, id int, pri uint32, delay time.Duration) (uint64, error) {
	raw, err := json.Marshal(Message{t, id})
	if err != nil {
		return 0, err
	}
	jid, err := Tube.Put(raw, pri, delay, 24*time.Hour)
	if err != nil {
		return 0, err
	}
	return jid, nil
}

func Recv() (*Message, uint64, error) {
	var jid uint64
	var body []byte
	var err error

	for {
		jid, body, err = Conn.Reserve(1 * time.Hour)
		if cerr, ok := err.(beanstalk.ConnError); ok && cerr.Err == beanstalk.ErrTimeout {
			continue
		} else if err != nil {
			return nil, 0, err
		}
		break
	}

	var msg Message
	if err = json.Unmarshal(body, &msg); err != nil {
		return nil, 0, err
	}
	return &msg, jid, nil
}

func Init() error {
	var err error
	Conn, err = beanstalk.Dial("tcp", env.StringDefault("OMA_Q_MOUNT", "127.0.0.1:11300"))
	if err != nil {
		return err
	}
	name := env.StringDefault("OMA_Q_TUBE", "oma")
	Tube = &beanstalk.Tube{Conn, name}
	Conn.TubeSet.Name = map[string]bool{name: true}
	return nil
}
