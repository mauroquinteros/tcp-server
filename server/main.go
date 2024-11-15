package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

type Server struct {
	clients map[net.Conn]bool
	mutex   sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		clients: make(map[net.Conn]bool),
	}
}

func (s *Server) broadcast(message string, sender net.Conn) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for client := range s.clients {
		if client != sender {
			_, err := fmt.Fprintf(client, "%s\n", message)
			if err != nil {
				log.Printf("Error broadcasting to client: %v", err)
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
		log.Printf("Client disconnected: %s", conn.RemoteAddr())
		s.mutex.Unlock()
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		log.Printf("Received message from %s: %s", conn.RemoteAddr(), message)
		s.broadcast(message, conn)
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
