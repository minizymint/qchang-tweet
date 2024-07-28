package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type NotificationService struct {
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	amqpConn  *amqp.Connection
	amqpChan  *amqp.Channel
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
	}
}

func (ns *NotificationService) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("Failed to upgrade to websocket: %v", err)
		return
	}
	defer ws.Close()

	ns.clients[ws] = true

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			delete(ns.clients, ws)
			break
		}
		ns.broadcast <- message
	}
}

func (ns *NotificationService) handleMessages() {
	for {
		message := <-ns.broadcast
		for client := range ns.clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Error writing message: %v", err)
				client.Close()
				delete(ns.clients, client)
			}
		}
	}
}

func (ns *NotificationService) connectToRabbitMQ() {
	var err error
	ns.amqpConn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ns.amqpChan, err = ns.amqpConn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	q, err := ns.amqpChan.QueueDeclare(
		"notifications",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	msgs, err := ns.amqpChan.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	go func() {
		for msg := range msgs {
			ns.broadcast <- msg.Body
		}
	}()
}

func main() {
	ns := NewNotificationService()

	go ns.handleMessages()
	ns.connectToRabbitMQ()

	http.HandleFunc("/ws", ns.handleConnections)

	log.Println("Notification service started on :8081")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
