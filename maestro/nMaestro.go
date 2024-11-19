package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Book struct {
	ID         int
	Title      string
	Genres     []string
	AvgRating  float64
	NumRatings int
}

type Peti struct {
	Send     int
	Opc      int
	MovGenre []string
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var env1, env2 net.Conn
var mu sync.Mutex

func initWorkerConnections() error {
	var err error
	time.Sleep(time.Second * 80)
	env1, err = net.Dial("tcp", "trabajador1:9002")
	if err != nil {
		return fmt.Errorf("error connecting to worker 1: %v", err)
	}

	env2, err = net.Dial("tcp", "trabajador2:9003")
	if err != nil {
		return fmt.Errorf("error connecting to worker 2: %v", err)
	}

	fmt.Println("Connections to workers established")
	return nil
}

func closeWorkerConnections() {
	if env1 != nil {
		env1.Close()
	}
	if env2 != nil {
		env2.Close()
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Client connected via WebSocket")

	for {
		var petic Peti
		err := conn.ReadJSON(&petic)
		if err != nil {
			log.Println("Error reading JSON:", err)
			break
		}

		requestBytes, _ := json.Marshal(petic)

		mu.Lock()
		fmt.Println(string(requestBytes))
		fmt.Fprint(env1, string(requestBytes)+"\n")
		fmt.Fprint(env2, string(requestBytes)+"\n")

		mu.Unlock()

		workerResponse1, err := bufio.NewReader(env1).ReadString('\n')
		if err != nil {
			log.Println("Error reading response from worker 1:", err)
			continue
		}

		workerResponse2, err := bufio.NewReader(env2).ReadString('\n')
		if err != nil {
			log.Println("Error reading response from worker 2:", err)
			continue
		}

		var recommendations1, recommendations2 []Book
		json.Unmarshal([]byte(workerResponse1), &recommendations1)
		json.Unmarshal([]byte(workerResponse2), &recommendations2)
		commonRecommendations := findCommonRecommendations(recommendations1, recommendations2)

		err = conn.WriteJSON(commonRecommendations)
		if err != nil {
			log.Println("Error sending recommendations:", err)
			break
		}
	}
}

func findCommonRecommendations(recs1, recs2 []Book) []Book {
	common := []Book{}
	recsMap := make(map[string]Book)

	for _, rec := range recs1 {
		recsMap[rec.Title] = rec
	}

	for _, rec := range recs2 {
		if _, exists := recsMap[rec.Title]; exists {
			common = append(common, rec)
		}
	}
	return common
}

func main() {
	// Inicializar la conexión con los trabajadores
	err := initWorkerConnections()
	if err != nil {
		log.Fatal("Failed to initialize worker connections:", err)
	}
	defer closeWorkerConnections()

	http.HandleFunc("/ws", wsHandler)
	fmt.Println("WebSocket server running on port 10001")
	log.Fatal(http.ListenAndServe("0.0.0.0:10001", nil))
}
