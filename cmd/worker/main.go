// See LICENSE.txt for licensing information.

package main

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	//_ "github.com/go-sql-driver/mysql"
	//_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"omapp/vendor/env"

	"omapp/pkg/model"
	"omapp/pkg/overmapper"
	"omapp/pkg/queue"
)

var (
	wg   sync.WaitGroup
	root string
)

func main() {
	log.Println("Starting worker...")

	root = env.String("OMA_DATA_ROOT")
	limit := env.IntDefault("OMA_W_LIMIT", 4)

	log.Println("Connecting to database...")
	if err := model.Init(); err != nil {
		log.Fatalln(err)
		os.Exit(3)
	}

	log.Println("Connecting to queue...")
	if err := queue.Init(); err != nil {
		log.Fatalln(err)
		os.Exit(3)
	}

	log.Println("Entering main loop.")
	var running int
	for {
		msg, jid, err := queue.Recv()
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}

		if running >= limit {
			wg.Wait()
			running = 0
		}

		running += 1
		wg.Add(1)
		go process(msg, jid)
	}
}

func process(msg *queue.Message, jid uint64) {
	var mr model.Map
	start := time.Now()
	say(jid, "Message:", msg)
	model.Db.First(&mr, msg.Target)

	switch msg.Type {
	case queue.MAP:
		data := path.Join(root, "uploads", strconv.Itoa(msg.Target))
		m, err := overmapper.NewMap(data)
		if err != nil {
			say(jid, "ERROR:", err)
			mr.State = model.BROKEN
			model.Db.Save(&mr)
			goto finish
		}
		say(jid, mr)
		mr.Width = m.Width
		mr.Height = m.Height
		mr.Area = m.Width * m.Height
		mr.Visited = len(m.Maps)
		say(jid, "Drawing...")
		i, err := m.Draw()
		if err != nil {
			say(jid, "ERROR:", err)
			mr.State = model.BROKEN
			model.Db.Save(&mr)
			goto finish
		}
		name := fmt.Sprintf("%d.png", msg.Target)
		target := path.Join(root, "maps", name)
		o, err := os.Create(target)
		if err != nil {
			say(jid, "ERROR:", err)
			mr.State = model.BROKEN
			model.Db.Save(&mr)
			goto finish
		}
		if err := png.Encode(o, i); err != nil {
			say(jid, "ERROR:", err)
			mr.State = model.BROKEN
			model.Db.Save(&mr)
			goto finish
		}
		mr.ImageName = name
		mr.State = model.READY
		model.Db.Save(&mr)
	case queue.DEL:
	}

finish:
	status := "Done"
	if mr.State == model.BROKEN {
		status = "Broken"
		say(jid, "Not ready, adding delete job.")
		djid, err := queue.Send(queue.DEL, msg.Target, 10, 1*time.Hour)
		if err != nil {
			status = "Very broken"
			say(jid, "FATAL: Couldn't add DEL job:", err)
		} else {
			say(jid, "Added delete job:", djid)
		}
	}
	end := time.Now()
	say(jid, "Status:", status, "Took:", end.Sub(start))
	queue.Conn.Delete(jid)
	wg.Done()
}

func say(jid uint64, args ...interface{}) {
	log.Println("Job:", jid, args)
}
