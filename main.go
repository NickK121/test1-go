package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"test1-go/model"
	"time"
)

var done chan os.Signal

func handleConnection(conn net.Conn) {
	user := model.AppendUser(conn)
	user.Run()
}

func Serve(port string) error {
	fmt.Println("Run as server")
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("cannot create connection: %s", err.Error())
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			return fmt.Errorf("cannot create connection: %s", err.Error())
		}
		go handleConnection(conn)
	}
}

func Client(port string) error {
	fmt.Println("Run as client")
	conn, err := net.Dial("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("cannot create connection: %s", err.Error())
	}
	user := model.NewUser(conn)
	id, err := user.ReadUserId()
	if err != nil {
		return fmt.Errorf("cannot read user id: %s", err.Error())
	}
	fmt.Println("UserId=", id)

	time.Sleep(2 * time.Second)
	if id == 1 {
		// send message to all users
		user.WriteMessage(0, []byte("Message to all"))

		// send message to 2 users
		user.WriteMessage(2, []byte("Message to 2 user"))
	}

	time.Sleep(2 * time.Second)
	done <- syscall.SIGQUIT
	return nil
}

func main() {
	var port string
	var help bool
	var server bool

	flag.StringVar(&port, "port", "8888", "port")
	flag.BoolVar(&server, "server", false, "run as server")
	flag.BoolVar(&help, "help", false, "Show this text")

	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	// check port
	if port == "" {
		fmt.Fprint(os.Stderr, "Port not defined.\n\n")
		flag.Usage()
		return
	}

	done = make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	if server {
		go func() {
			err := Serve(port)
			if err != nil {
				log.Fatal(err.Error())
			}
		}()
	} else {
		go func() {
			err := Client(port)
			if err != nil {
				log.Fatal(err.Error())
			}
		}()
	}

	<-done
	model.CloseAllConnections()
}
