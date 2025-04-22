package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/cpu"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
var clients = make(map[*websocket.Conn]bool)

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		panic("cant establish connection")
	}
	ticker := time.NewTicker(time.Second)

	defer func() {
		c.Close()
		delete(clients, c)
		ticker.Stop()
	}()

	clients[c] = true

	for {
		select {
		case <-ticker.C:
			// out, err := exec.Command("sh", "/var/scripts/t.sh").Output()
			// if err != nil {
			// 	fmt.Println(err)
			// 	panic(err)
			// }
			cpuUsage, err := cpu.Percent(time.Second, false)
			if err != nil {
				log.Printf("Error getting CPU usage: %s", err.Error())
			} else {
				log.Printf("CPU Usage: %.2f%%", cpuUsage[0])
				handleMessage([]byte(fmt.Sprintf("%f", cpuUsage[0])))
			}
		}
	}
}

func handleMessage(msg []byte) {
	for cli := range clients {
		_ = cli.WriteMessage(websocket.TextMessage, msg)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ws", wsEndpoint)
	log.Fatal(http.ListenAndServe(":8030", r))
}
