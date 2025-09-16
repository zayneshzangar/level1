package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

// TelnetClient represents a telnet client.
type TelnetClient struct {
	host    string
	port    string
	timeout time.Duration
	conn    net.Conn
}

// NewTelnetClient creates a new TelnetClient instance.
func NewTelnetClient(host, port string, timeout time.Duration) *TelnetClient {
	return &TelnetClient{
		host:    host,
		port:    port,
		timeout: timeout,
	}
}

// Connect establishes a TCP connection to the server.
func (tc *TelnetClient) Connect() error {
	dialer := &net.Dialer{Timeout: tc.timeout}
	conn, err := dialer.Dial("tcp", tc.host+":"+tc.port)
	if err != nil {
		return fmt.Errorf("failed to connect to %s:%s: %v", tc.host, tc.port, err)
	}
	tc.conn = conn
	return nil
}

// Run starts the telnet client, handling input and output concurrently.
func (tc *TelnetClient) Run() error {
	if tc.conn == nil {
		return fmt.Errorf("no connection established")
	}
	defer tc.conn.Close()

	// Channel to signal when to stop
	done := make(chan struct{})

	// Read from socket, write to STDOUT
	go func() {
		scanner := bufio.NewScanner(tc.conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "error reading from socket: %v\n", err)
		}
		close(done) // Signal when server closes connection
	}()

	// Read from STDIN, write to socket
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			_, err := fmt.Fprintln(tc.conn, scanner.Text())
			if err != nil {
				fmt.Fprintf(os.Stderr, "error writing to socket: %v\n", err)
				close(done)
				return
			}
		}
		if err := scanner.Err(); err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "error reading from stdin: %v\n", err)
		}
		// Ctrl+D (EOF) received
		close(done)
	}()

	// Wait for either goroutine to finish
	<-done
	return nil
}

func main() {
	// Parse command-line flags
	timeout := flag.Duration("timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--timeout=<duration>] <host> <port>\n", os.Args[0])
		os.Exit(1)
	}

	host := flag.Arg(0)
	port := flag.Arg(1)

	// Create and run telnet client
	client := NewTelnetClient(host, port, *timeout)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err := client.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
