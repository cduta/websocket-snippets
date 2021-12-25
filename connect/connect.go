package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/net/websocket"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "ryoga"
	password = ""
	dbname   = "requests"
)

func sendJSONString(json string, socket *websocket.Conn) {
	websocket.Message.Send(socket, string(json))
}

func dbNotifyListener(conn *pgxpool.Conn, socket *websocket.Conn, cancelContext context.Context) {
	var err error
	const channel = "callback"

	for cancelContext.Err() == nil && err == nil {
		_, err = conn.Exec(cancelContext, "LISTEN "+channel)
		if err != nil {
			if cancelContext.Err() == nil {
				log.Printf("Could not listen to the database notifications of %s: %v\n", channel, err)
			}
			break
		}
		notification, err := conn.Conn().WaitForNotification(cancelContext)
		if err != nil {
			if cancelContext.Err() == nil {
				log.Printf("Could not wait for notification of %s: %v\n", channel, err)
			}
			break
		}

		sendJSONString(notification.Payload, socket)
	}
}

func makeWebSocketConnect(dbpool *pgxpool.Pool) func(*websocket.Conn) {
	return func(socket *websocket.Conn) {
		var err error

		conn, err := dbpool.Acquire(context.Background())
		if err != nil {
			log.Printf("Could not aquire database base connection: %v\n", err)
		}
		defer conn.Release()

		cancelContext, cancel := context.WithCancel(context.Background())
		defer cancel()

		callbackConn, err := dbpool.Acquire(cancelContext)
		if err != nil {
			log.Printf("Could not aquire database callback connection: %v\n", err)
		}
		defer callbackConn.Release()

		log.Printf("Client connected...\n")

		go dbNotifyListener(callbackConn, socket, cancelContext)

		var jsonMessage string

		for {
			err = websocket.Message.Receive(socket, &jsonMessage)
			if err != nil {
				if err.Error() == "EOF" {
					log.Printf("Client disconnected...")
				} else {
					log.Printf("Could not receive message via WebSocket: %v\n", err)
				}
				return
			}
		}
	}
}

func loadSQLFile(path string) (string, error) {
	initSQLBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(initSQLBytes), err
}

func initDB() (*pgxpool.Pool, error) {
	psqlInfo := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=4&pool_max_conn_lifetime=5s&pool_max_conn_idle_time=3s", user, password, host, port, dbname)

	config, err := pgxpool.ParseConfig(psqlInfo)
	if err != nil {
		return nil, err
	}

	dbpool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	initSQL, err := loadSQLFile(filepath.Join("sql", "initialize.sql"))
	if err != nil {
		return nil, err
	}

	_, err = dbpool.Exec(context.Background(), initSQL)
	if err != nil {
		return nil, err
	}

	return dbpool, nil
}

func initHTTPServer(dbpool *pgxpool.Pool) {
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./html/"))))
	http.Handle("/connect", websocket.Handler(makeWebSocketConnect(dbpool)))

	log.Println("Listening on :8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
	return
}

func main() {
	dbpool, err := initDB()
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
		return
	}
	defer dbpool.Close()
	initHTTPServer(dbpool)
}
