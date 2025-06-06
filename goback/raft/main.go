package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
	envFile, _ := godotenv.Read("/home/.env")
	db_name := envFile["POSTGRES_DB"]
	user_name := envFile["POSTGRES_USER"]
	pass := envFile["POSTGRES_PASSWORD"]

	// root_pass := os.Getenv("POSTGRES_ROOT_PASSWORD")
	// port := os.Getenv("POSTGRES_PORT")
	type User struct {
		Id   int    `db:"id"`
		Name string `db:"name"`
	}
	r.HandleFunc("/ws", wsEndpoint)
	r.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		info := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=%s", user_name, pass, "postgresql", db_name, "disable")
		fmt.Println(info)
		db, err := sql.Open("postgres", info)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			fmt.Println(err)
			fmt.Println("cant ping")
			db.Close()
		}

		user := User{}
		err = db.QueryRow("Select id, name from role where id=$1", 0).Scan(&user.Id, &user.Name)
		// err = rows.Scan(&user)
		if err != nil {
			panic(err)
		}
		fmt.Println(user.Name)
		fmt.Fprint(w, user.Name)
	})
	log.Fatal(http.ListenAndServe(":8030", r))
}
