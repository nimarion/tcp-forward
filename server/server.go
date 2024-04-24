package server

import (
	"fmt"
	"net"
	"sync"
)

type Client struct {
	conn net.Conn
}

type Server struct {
	Port         int
	clients      []*Client
	mu           sync.Mutex
	MessageChan  chan string
	UdpBroadcast string
}

func (s *Server) handleConnection(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	client := &Client{conn: conn}
	s.clients = append(s.clients, client)
	fmt.Printf("Client  %s connected to port %d\n", conn.RemoteAddr().String(), s.Port)
}

func (s *Server) Broadcast(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.clients {
		_, err := client.conn.Write(data)
		// print client port
		if err != nil {
			fmt.Printf("Error writing to client: %v\n", err)

			// Optionally handle error, e.g., remove client from list

			index := -1
			for i, c := range s.clients {
				if c == client {
					index = i
					break
				}
			}

			if index != -1 {
				fmt.Printf("Receiving TCP Client %s disconnected\n", client.conn.RemoteAddr().String())
				s.clients = append(s.clients[:index], s.clients[index+1:]...)
			}
		}
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf(": %d", s.Port))
	if err != nil {
		fmt.Printf("Error starting server on port %d: %s\n", s.Port, err)
		return
	}
	defer listener.Close()
	fmt.Printf("TCP server started and listening on port %d\n", s.Port)

	go s.handleMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection on port %d: %s\n", s.Port, err)
			continue
		}
		// Handle connection asynchronously
		go s.handleConnection(conn)
	}
}

func (s *Server) handleMessage() {
	for msg := range s.MessageChan {
		//fmt.Printf("Received message on port %d: %s\n", s.Port, msg)
		s.Broadcast([]byte(msg))
	}
}
