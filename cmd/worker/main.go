// See LICENSE.txt for licensing information.

package main

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/darkhelmet/env"

	"omapp/pkg/model"
	"omapp/pkg/overmapper"
	"omapp/pkg/queue"
)

var root string

func main() {
	log.Println("Starting worker...")
	root = env.String("OMA_DATA_ROOT")
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
	for {
		msg, jid, err := queue.Recv()
		if err != nil {
			log.Println("ERROR:", err)
			continue
		}
		process(jid, msg)
	}
}

func process(jid uint64, msg *queue.Message) {
	var rec model.Map
	start := time.Now()
	log.Println("Job:", jid, "Message:", msg)
	model.Db.First(&rec, msg.Target)
	switch msg.Type {
	case queue.MAP:
		err := doMap(jid, msg, &rec)
		delay := 10 * time.Second
		if rec.State == model.BROKEN {
			log.Println("Job:", jid, "ERROR:", err)
			delay = 4 * time.Hour
		}
		log.Println("Job:", jid, "Saving record...")
		if err := model.Db.Save(&rec).Error; err != nil {
			log.Println("Job:", jid, "ERROR:", err, "on save for:", msg.Target)
		}
		log.Println("Job:", jid, "Enqueuing clean job...")
		if _, err := queue.Send(queue.CLEAN, msg.Target, 10, delay); err != nil {
			log.Println("Job:", jid, "FATAL:", err, "on enqueue clean for:", msg.Target)
		}
	case queue.CLEAN:
		data := path.Join(root, "uploads", strconv.Itoa(msg.Target))
		log.Println("Job:", jid, "Removing all uploaded files...")
		if err := os.RemoveAll(data); err != nil {
			log.Println("Job:", jid, "ERROR:", err)
		}
		if rec.State == model.BROKEN {
			name := fmt.Sprintf("%d.png", msg.Target)
			target := path.Join(root, "maps", name)
			if _, err := os.Stat(target); err == nil {
				log.Println("Job:", jid, "Removing map image...")
				if err := os.Remove(target); err != nil {
					log.Println("Job:", jid, "ERROR:", err)
				}
			}
		}
	case queue.DELETE:
		target := path.Join(root, "maps", rec.ImageName)
		log.Println("Job:", jid, "Removing map image...")
		if err := os.Remove(target); err != nil {
			log.Println("Job:", jid, "ERROR:", err)
		}
		log.Println("Job:", jid, "Removing database record...")
		if err := model.Db.Delete(&rec).Error; err != nil {
			log.Println("Job:", jid, "FATAL:", err)
		}
	}
	end := time.Now()
	log.Println("Job:", jid, "Took:", end.Sub(start))
	if err := queue.Conn.Delete(jid); err != nil {
		log.Println("Job:", jid, "FATAL:", err, "on done")
	}
}

func doMap(jid uint64, msg *queue.Message, rec *model.Map) error {
	data := path.Join(root, "uploads", strconv.Itoa(msg.Target))
	m, err := overmapper.NewMap(data)
	if err != nil {
		rec.State = model.BROKEN
		return err
	}
	log.Println("Job:", jid, m)
	rec.Width = m.Width
	rec.Height = m.Height
	rec.Area = m.Width * m.Height
	rec.Visited = len(m.Maps)
	log.Println("Job:", jid, "Drawing...")
	i, err := m.Draw()
	if err != nil {
		rec.State = model.BROKEN
		return err
	}
	name := fmt.Sprintf("%d.png", msg.Target)
	target := path.Join(root, "maps", name)
	log.Println("Job:", jid, "Saving...")
	o, err := os.Create(target)
	if err != nil {
		rec.State = model.BROKEN
		return err
	}
	if err := png.Encode(o, i); err != nil {
		rec.State = model.BROKEN
		return err
	}
	log.Println("Job:", jid, "Mapping done.")
	rec.ImageName = name
	rec.State = model.READY
	return nil
}
