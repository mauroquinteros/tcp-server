package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
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

func readFile(path string) (string, string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file: %v", err)
	}

	fileName := filepath.Base(path)
	encoded := base64.StdEncoding.EncodeToString(content)
	return fileName, encoded, nil
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
		if err := os.MkdirAll("downloads", 0755); err != nil {
			log.Fatalf("Failed to create downloads directory: %v", err)
		}

		subscribeMsg := fmt.Sprintf("CHANNEL:%s|MESSAGE:SUBSCRIBE", channel)
		_, err = fmt.Fprintf(conn, "%s\n", subscribeMsg)
		if err != nil {
			log.Fatalf("Failed to subscribe: %v", err)
		}
		log.Printf("Subscribed to channel %s", channel)

		scanner := bufio.NewScanner(conn)
		const maxCapacity = 10 * 1024 * 1024
		scanner.Buffer(make([]byte, 0, maxCapacity), maxCapacity)
		log.Printf("Listening for messages on channel %s", channel)

		for scanner.Scan() {
			message := scanner.Text()
			parts := strings.Split(message, "|")

			if len(parts) == 2 && strings.HasPrefix(parts[0], "FILE:") {
				fileName := strings.TrimPrefix(parts[0], "FILE:")
				content := strings.TrimPrefix(parts[1], "CONTENT:")

				decoded, err := base64.StdEncoding.DecodeString(content)
				if err != nil {
					log.Fatalf("Failed to decode file: %v", err)
					continue
				}

				filePath := filepath.Join("downloads", fileName)
				if err := os.WriteFile(filePath, decoded, 0644); err != nil {
					log.Printf("Error saving file: %v", err)
					continue
				}

				log.Printf("Received and saved file: %s", fileName)
			} else {
				log.Printf("Received message: %s", message)
			}
		}
	} else if message != "" {
		file, err := os.Stat(message)

		if err == nil && !file.IsDir() {
			fileName, encoded, err := readFile(message)
			if err != nil {
				log.Fatalf("Failed to read file: %v", err)
			}

			broadcastMsg := fmt.Sprintf("CHANNEL:%s|FILE:%s|CONTENT:%s", channel, fileName, encoded)
			_, err = fmt.Fprintf(conn, "%s\n", broadcastMsg)

			if err != nil {
				log.Fatalf("Failed to send file %v", err)
			}
			log.Printf("Sent file %s (size: %d bytes) on channel %s", fileName, file.Size(), channel)
		} else {
			log.Println("The message is", message)
			broadcastMsg := fmt.Sprintf("CHANNEL:%s|MESSAGE:%s", channel, message)
			_, err = fmt.Fprintf(conn, "%s\n", broadcastMsg)

			if err != nil {
				log.Fatalf("Failed to send message %v", err)
			}
			log.Printf("Sent message %s on channel %s", message, channel)
		}
	}
}
