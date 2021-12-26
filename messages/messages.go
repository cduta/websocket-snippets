package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/net/websocket"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "ryoga"
	password = ""
	dbname   = "messages"
)

func dbNotifyListener(conn *pgxpool.Conn, socket *websocket.Conn, notifyContext context.Context, userName string, listenerReady chan bool) {
	defer conn.Release()

	var err error
	const CHANNEL = "callback"

	_, err = conn.Exec(notifyContext, "LISTEN "+CHANNEL)
	if err != nil {
		log.Printf("%s: Could not listen to the database notifications of %s: %v\n", userName, CHANNEL, err)
		if notifyContext.Err() != nil {
			log.Printf("NotifyContext says: %v\n", notifyContext.Err())
		}
		listenerReady <- false
		return
	}

	listenerReady <- true

	for notifyContext.Err() == nil && err == nil {
		notification, err := conn.Conn().WaitForNotification(notifyContext)
		if err == nil {
			websocket.Message.Send(socket, notification.Payload)
		}
	}
}

func spawnNotifyListener(dbpool *pgxpool.Pool, userName string, socket *websocket.Conn, notifyContext context.Context) bool {
	var err error
	var listenerReady chan bool = make(chan bool)

	callbackConn, err := dbpool.Acquire(notifyContext)
	if err != nil {
		log.Printf("%s: Could not aquire database callback connection: %v\n", userName, err)
		return false
	}

	go dbNotifyListener(callbackConn, socket, notifyContext, userName, listenerReady)

	return <-listenerReady
}

func makeWebSocketConnect(dbpool *pgxpool.Pool) func(*websocket.Conn) {
	return func(socket *websocket.Conn) {
		var err error
		var userName string = uuid.NewString()
		var jsonData string

		conn, err := dbpool.Acquire(context.Background())
		if err != nil {
			log.Printf("%s: Could not aquire database base connection: %v\n", userName, err)
			return
		}
		defer conn.Release()

		notifyContext, cancel := context.WithCancel(context.Background())
		defer cancel()

		if !spawnNotifyListener(dbpool, userName, socket, notifyContext) {
			log.Printf("%s: Could not spawn notify listener.\n", userName)
			return
		}

		err = conn.QueryRow(context.Background(), "SELECT format_json('message_change', array_to_json(recent_messages()) :: text)").Scan(&jsonData)
		if err != nil {
			log.Printf("%s: Could not aquire initial messages: %v\n", userName, err)
			return
		}
		websocket.Message.Send(socket, jsonData)

		// Only when reaching here, is the connection truly established
		log.Printf("%s connected...\n", userName)

		_, err = conn.Exec(context.Background(), "SELECT add_user($1)", userName)
		if err != nil {
			log.Printf("Could not add user %s: %v\n", userName, err)
			if context.Background().Err() != nil {
				log.Printf("BackgroundContext says: %v\n", context.Background().Err())
			}
			return
		}

		for {
			err = websocket.Message.Receive(socket, &jsonData)
			if err != nil {
				_, conErr := conn.Exec(context.Background(), "SELECT remove_user($1)", userName)
				if conErr != nil {
					log.Printf("Could not remove user %s: %v\n", userName, conErr)
					if context.Background().Err() != nil {
						log.Printf("BackgroundContext says: %v\n", context.Background().Err())
					}
				}
				if err.Error() == "EOF" {
					log.Printf("%s disconnected...", userName)
				} else {
					log.Printf("Could not receive message via WebSocket: %v\n", err)
				}
				return
			}
			_, err = conn.Exec(context.Background(), "SELECT receive_data($1,$2)", userName, jsonData)
			if err != nil {
				log.Printf("Could not receive data from user %s: %v\n", userName, err)
				if context.Background().Err() != nil {
					log.Printf("BackgroundContext says: %v\n", context.Background().Err())
				}
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
	psqlInfo := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=32&pool_max_conn_lifetime=5s&pool_max_conn_idle_time=3s", user, password, host, port, dbname)

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
	http.Handle("/messages", websocket.Handler(makeWebSocketConnect(dbpool)))

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
