package main

import (
	"encoding/csv"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/nimarion/tcp-forward/server"
)

func connectToServer(addr string, messageChan chan string) {
	for {
		// Connect to the TCP server
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			fmt.Println("Failed to connect to TCP Server", addr, ":", err)
			// Retry after some time
			time.Sleep(5 * time.Second)
			continue
		}

		defer conn.Close()

		fmt.Printf("Connected to TCP Server %s\n", addr)

		// Create a buffer to read bytes from the connection
		buffer := make([]byte, 1024)

		// Continuously read bytes from the connection until encountering a null byte
		for {
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("Error reading from", addr, ":", err)
				break
			}

			message := ""
			for i := 0; i < n; i++ {
				message += string(buffer[i])
				if buffer[i] == 0 {
					messageChan <- message
					message = ""
					break
				}
			}
		}

		// Auto reconnect after some time
		time.Sleep(5 * time.Second)
	}
}

func removeDuplicates(intList []int) []int {
	encountered := map[int]bool{}
	result := []int{}

	for _, num := range intList {
		if !encountered[num] {
			encountered[num] = true
			result = append(result, num)
		}
	}

	return result
}

func main() {
	file, err := os.Open("connections.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'

	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Map holding client_host-port:server_port mapping
	keyValueMap := make(map[string]int)
	for i, record := range records {
		if i == 0 {
			continue
		}
		// Convert the string value to an integer before assigning it to the map
		port, err := strconv.Atoi(record[1])
		if err != nil {
			// Handle the error if the conversion fails
			// For example, log the error or take appropriate action
			fmt.Println("Error converting string to int:", err)
			continue // Skip this record and move to the next one
		}
		var client = record[0]
		keyValueMap[client] = port
	}

	// All server ports which should be started
	var serverPorts []int
	for _, value := range keyValueMap {
		serverPorts = append(serverPorts, value)
	}
	serverPorts = removeDuplicates(serverPorts)

	// Start TCP Servers
	servers := make([]*server.Server, len(serverPorts))
	for i, port := range serverPorts {
		servers[i] = &server.Server{Port: port, MessageChan: make(chan string)}
		go servers[i].Start()
	}

	// Connect to clients and connect to corresponding message channel of server
	for key, value := range keyValueMap {
		for _, server := range servers {
			if server.Port == value {
				go connectToServer(key, server.MessageChan)
			}
		}
	}

	select {}
}
