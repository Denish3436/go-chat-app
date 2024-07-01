package main

import (
    "fmt"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

type Message struct {
    Message   string `json:"message"`
    Sender    string `json:"sender"`
    Timestamp string `json:"timestamp"`
    Type      string `json:"type"`
}

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var clients = make(map[*websocket.Conn]string)
var broadcast = make(chan Message)
var history []Message
var mutex = &sync.Mutex{}
const historySize = 20

func main() {
    fs := http.FileServer(http.Dir("public"))
    http.Handle("/", fs)
    http.HandleFunc("/ws", handleConnections)
    go handleMessages()

    fmt.Println("Server started on :8082")
    err := http.ListenAndServe(":8082", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    defer ws.Close()

    mutex.Lock()
    clients[ws] = ""
    mutex.Unlock()

    sendHistory(ws)

    for {
        var msg Message
        err := ws.ReadJSON(&msg)
        if err != nil {
            log.Println(err)
            mutex.Lock()
            username := clients[ws]
            delete(clients, ws)
            mutex.Unlock()
            broadcast <- Message{Sender: "System", Message: username + " has left the chat", Timestamp: time.Now().Format("15:04:05"), Type: "system"}
            break
        }

        if msg.Type == "username" {
            mutex.Lock()
            clients[ws] = msg.Sender
            mutex.Unlock()
            broadcast <- Message{Sender: "System", Message: msg.Sender + " has joined the chat", Timestamp: time.Now().Format("15:04:05"), Type: "system"}
        } else {
            msg.Timestamp = time.Now().Format("15:04:05")
            broadcast <- msg
        }
    }
}

func handleMessages() {
    for {
        msg := <-broadcast

        mutex.Lock()
        if len(history) >= historySize {
            history = history[1:]
        }
        history = append(history, msg)

        for client := range clients {
            err := client.WriteJSON(msg)
            if err != nil {
                log.Println(err)
                client.Close()
                delete(clients, client)
            }
        }
        mutex.Unlock()
    }
}

func sendHistory(ws *websocket.Conn) {
    mutex.Lock()
    defer mutex.Unlock()
    for _, msg := range history {
        err := ws.WriteJSON(msg)
        if err != nil {
            log.Println(err)
            return
        }
    }
}

