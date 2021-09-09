package model

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

type User struct {
	id   int
	conn net.Conn
	in   chan []byte
}

func NewUser(conn net.Conn) *User {
	user := User{
		id:   NextId(),
		conn: conn,
		in:   make(chan []byte),
	}
	return &user
}

func AppendUser(conn net.Conn) *User {
	user := NewUser(conn)
	users = append(users, *user)
	return user
}

var users []User

func NextId() int {
	return len(users) + 1
}

func (u *User) ReciveMsg() {
	f := func() {
		buff := <-u.in
		fmt.Printf("User %d recive: %s\n", u.id, string(buff))
	}
	go f()
}

func (u *User) Send(buff []byte) error {
	size, err := u.conn.Write(buff)
	if err != nil || size != len(buff) {
		return fmt.Errorf("ERROR: written less than expected - size:%d, length of buff:%d", size, len(buff))
	}
	return nil
}

func (u *User) ReadUserId() (int, error) {
	buff := make([]byte, 2)

	// get user ID (recepient, if ID == 0 then broadcast message)
	size, err := u.conn.Read(buff)
	if err != nil {
		return -1, err
	}
	if size == 0 {
		return -1, errors.New("read less then 1 byte")
	}
	userId, err := strconv.Atoi(string(buff[:size]))
	if err != nil {
		return -1, err
	}
	return userId, nil
}

func (u *User) ReadMessage() ([]byte, error) {
	buff := make([]byte, 2)

	// get user ID (recepient, if ID == 0 then broadcast message)
	size, err := u.conn.Read(buff)
	if err != nil {
		return nil, err
	}
	if size != 2 {
		return nil, errors.New("read less then 2 bytes")
	}

	lengthMsg, err := strconv.Atoi(string(buff))
	if err != nil {
		return nil, err
	}
	buff = make([]byte, lengthMsg)
	size, err = u.conn.Read(buff)
	if err != nil {
		return nil, err
	}
	if size != lengthMsg {
		return nil, fmt.Errorf("read less then %d bytes", lengthMsg)
	}
	return buff, nil
}

func (u *User) WriteMessage(userId int, buff []byte) error {
	uid := fmt.Sprintf("%02d", userId)
	size := []byte(uid + strconv.Itoa(len(buff)))
	msg := append(size, buff...)
	return u.Send(msg)
}

func (u *User) Run() {
	u.ReciveMsg()
	fmt.Printf("User %d started\n", u.id)

	u.Send([]byte(strconv.Itoa(u.id)))
	for {
		userId, err := u.ReadUserId()
		if err == io.EOF {
			u.conn.Close()
			break
		}
		if err != nil {
			log.Println("ERROR: ", err.Error())
			continue
		}

		buff, err := u.ReadMessage()
		if err == io.EOF {
			u.conn.Close()
			break
		}
		if err != nil {
			log.Println("ERROR: ", err.Error())
			continue
		}

		if userId == 0 {
			MessengerInstance().SendAll(buff)
		} else {
			MessengerInstance().Send(userId, buff)
		}
	}
}

func CloseAllConnections() {
	for _, u := range users {
		u.conn.Close()
	}
}
