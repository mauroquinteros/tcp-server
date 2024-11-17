package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
)

func setUpFlags() (string, string, bool) {
	message := flag.String("send", "", "Message to send to server")
	receive := flag.Bool("receive", false, "Start in receive mode")
	channel := flag.String("channel", "", "Channel to subscribe to")
	flag.Parse()

	if *channel == "" {
		log.Fatal("Please specify a channel using --channel flag")
	}

	if *message == "" && !*receive {
		log.Fatal("Please provide a message using --send flag or start in receive mode using --receive flag")
	}
	return *channel, *message, *receive
}

func main() {
	channel, message, receive := setUpFlags()

	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	log.Println("Connected to server!")
	defer conn.Close()

	if receive {
		subscribeMsg := fmt.Sprintf("CHANNEL:%s|MESSAGE:SUBSCRIBE", channel)
		_, err = fmt.Fprintf(conn, "%s\n", subscribeMsg)
		if err != nil {
			log.Fatalf("Failed to subscribe: %v", err)
		}
		log.Printf("Subscribed to channel %s", channel)

		scanner := bufio.NewScanner(conn)
		log.Printf("Listening for messages on channel %s", channel)
		for scanner.Scan() {
			message := scanner.Text()
			log.Printf("Received message: %s\n", message)
		}
	} else if message != "" {
		broadcastMsg := fmt.Sprintf("CHANNEL:%s|MESSAGE:%s", channel, message)
		_, err = fmt.Fprintf(conn, "%s\n", broadcastMsg)
		if err != nil {
			log.Fatalf("Failed to send message %v", err)
		}
		log.Printf("Sent message %s on channel %s", message, channel)
	}
}
