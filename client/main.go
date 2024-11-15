package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
)

func setUpFlags() (string, bool) {
	message := flag.String("send", "", "Message to send to server")
	receive := flag.Bool("receive", false, "Start in receive mode")
	flag.Parse()

	if *message == "" && !*receive {
		log.Fatal("Please provide a message using --send flag or start in receive mode using --receive flag")
	}
	return *message, *receive
}

func main() {
	message, receive := setUpFlags()

	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	log.Println("Connected to server!")
	defer conn.Close()

	if receive {
		log.Println("Listening for messages")
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			message := scanner.Text()
			log.Printf("Received message: %s\n", message)
		}
	} else if message != "" {
		_, err = fmt.Fprintf(conn, "%s\n", message)
		if err != nil {
			log.Fatalf("Failed to send message: %v", err)
		}
		log.Printf("Sent message: %s", message)
	}
}
