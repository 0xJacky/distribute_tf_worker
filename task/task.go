package task

import (
	"distribute_tf_worker/testflight"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Task struct {
	ID        int    `json:"id"`
	Url       string `json:"url"`
	RequestID string `json:"request_id"`
}

var taskChan = make(chan Task, 20)

var trans map[int]string

func init() {
	trans = map[int]string{
		1: "open",
		2: "close",
		3: "full",
	}
}

func Receive(c *websocket.Conn, done chan struct{}) {
	var task Task

	defer close(done)
	for {
		err := c.ReadJSON(&task)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				log.Println("[Error] ReadJSON", err)
			}
			return
		}

		log.Printf("[Receive] id:%d, url:%s", task.ID, task.Url)
		taskChan <- task
	}
}

func Handle(c *websocket.Conn, done chan struct{}, mutex *sync.Mutex) {
	defer close(done)
	workerChan := make(chan struct{}, 10)
	for task := range taskChan {
		workerChan <- struct{}{}
		go func(task Task) {
			defer func() {
				if err := recover(); err != nil {
					log.Println("[Recover] Handle", err)
				}
			}()

			name, status, err := testflight.ParseTestflightApp(task.Url)

			if err != nil {
				log.Println("[Error] Handle parse app", err)
				return
			}

			log.Printf("[Result] id:%d, name: %s, status: %s", task.ID, name, trans[status])

			type res map[string]interface{}
			mutex.Lock()
			err = c.WriteJSON(&res{
				"id":         task.ID,
				"name":       name,
				"request_id": task.RequestID,
				"status":     status,
			})
			mutex.Unlock()
			if err != nil {
				log.Println("[Error] Handle write json", err)
				return
			}

			<-workerChan
		}(task)
	}
}
