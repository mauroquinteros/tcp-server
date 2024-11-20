package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type Server struct {
	clients  map[net.Conn]bool
	channels map[string]map[net.Conn]bool
	mutex    sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		clients:  make(map[net.Conn]bool),
		channels: make(map[string]map[net.Conn]bool),
	}
}

func (s *Server) subscribeToChannel(channel string, conn net.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.channels[channel]; !exists {
		s.channels[channel] = make(map[net.Conn]bool)
	}
	s.channels[channel][conn] = true
	log.Printf("Client %s subscribed to channel %s", conn.RemoteAddr(), channel)
}

func (s *Server) broadcast(message string, channel string, sender net.Conn) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if subscribers, exists := s.channels[channel]; exists {
		for client := range subscribers {
			if client != sender {
				_, err := fmt.Fprintf(client, "%s\n", message)
				if err != nil {
					log.Printf("Error broadcasting to client: %v", err)
				}
			}
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("New client connected: %s", conn.RemoteAddr())

	s.mutex.Lock()
	s.clients[conn] = true
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		delete(s.clients, conn)
		for _, subscribers := range s.channels {
			delete(subscribers, conn)
		}

		log.Printf("Client disconnected: %s", conn.RemoteAddr())
		s.mutex.Unlock()
	}()

	scanner := bufio.NewScanner(conn)
	const maxCapacity = 10 * 1024 * 1024
	scanner.Buffer(make([]byte, 0, maxCapacity), maxCapacity)

	for scanner.Scan() {
		rawMessage := scanner.Text()
		parts := strings.Split(rawMessage, "|")

		if len(parts) < 2 {
			log.Printf("Invalid message format from %s: %s", conn.RemoteAddr(), rawMessage)
			continue
		}
		channel := strings.TrimPrefix(parts[0], "CHANNEL:")

		if len(parts) == 2 {
			content := strings.TrimPrefix(parts[1], "MESSAGE:")
			messageType := "BROADCAST"
			if content == "SUBSCRIBE" {
				messageType = "SUBSCRIBE"
			}

			switch messageType {
			case "SUBSCRIBE":
				s.subscribeToChannel(channel, conn)
			case "BROADCAST":
				log.Printf("Received message from %s on channel %s: %s", conn.RemoteAddr(), channel, content)
				s.subscribeToChannel(channel, conn)
				s.broadcast(content, channel, conn)
			}
		} else if len(parts) == 3 {
			fileName := strings.TrimPrefix(parts[1], "FILE:")
			content := strings.TrimPrefix(parts[2], "CONTENT:")
			log.Printf("Received file from %s on channel %s: %s", conn.RemoteAddr(), channel, fileName)
			s.subscribeToChannel(channel, conn)
			fileContent := fmt.Sprintf("FILE:%s|CONTENT:%s", fileName, content)
			s.broadcast(fileContent, channel, conn)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from connection: %v", err)
	}
}

func (s *Server) start(port string) {
	listener, err := net.Listen("tcp", ":"+port)

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	log.Println("Server started on port", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func main() {
	server := NewServer()
	server.start("3000")
}
