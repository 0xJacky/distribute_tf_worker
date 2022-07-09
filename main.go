package main

import (
	"distribute_tf_worker/task"
	"github.com/caarlos0/env/v6"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"
)

type config struct {
	Env   string `env:"Env"`
	Token string `env:"Token"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("%+v\n", err)
	}
	var addr string
	switch cfg.Env {
	case "production":
		addr = "betahub.cn"
	case "development":
		addr = "ibeta.jackyu.cn"
	}

	log.SetFlags(3)

	log.Println("[Info] Distribute Testflight Worker Designed by 0xJacky")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: addr, Path: "/api/admin/testflight/tasks"}
	log.Printf("[Info] connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("[Error] dial:", err)
	}
	defer c.Close()
	log.Println("[Info] connected to the server successfully")

	// 认证
	var j struct {
		Token string
	}
	j.Token = cfg.Token

	err = c.WriteJSON(&j)
	if err != nil {
		log.Println("[Error] Authorization WriteJSON", err)
		return
	}

	done := make(chan struct{})
	var mutex sync.Mutex

	go task.Receive(c, done)
	go task.Handle(c, done, &mutex)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			err = c.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second))
			if err != nil {
				log.Println("[Error] ping server", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("[Error] write close:", err)
				return
			}
			select {
			case <-done:
				return
			case <-time.After(time.Second):
			}
			return
		}
	}
}
