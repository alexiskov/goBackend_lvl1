package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

type IM struct {
	Name, Message string
}

func auth() IM {
	thisUser := IM{}
	os.Stdout.Write([]byte("Введите Ваш ник: "))
	bufer := make([]byte, 256)
	n, err := os.Stdin.Read(bufer)
	if err != nil {
		panic(err)
	}
	bufer = bufer[0:n]
	thisUser.Name = strings.Trim(string(bufer), " \n\r")
	thisUser.Message = "handshake"
	return thisUser
}

func main() {
	thisUser := auth()
	call, err := net.Dial("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer call.Close()
	handshake, err := json.Marshal(thisUser)
	if err != nil {
		panic(err)
	}
	call.Write(handshake)

	go clientRead(call)
	go clientWrite(call, thisUser)
	for {
	}
}

func clientWrite(conn net.Conn, tu IM) {
	for {
		bufer := make([]byte, 256)
		n, err := os.Stdin.Read(bufer)
		if err != nil {
			log.Println(err)
			continue
		}
		bufer = bufer[0:n]
		tu.Message = strings.Trim(string(bufer), "\n\r")
		msg, err := json.Marshal(tu)
		if err != nil {
			log.Println(err)
			continue
		}
		_, err = conn.Write(msg)
	}
}

func clientRead(conn net.Conn) {

	for {
		bufer := make([]byte, 256*2)
		n, err := conn.Read(bufer)
		if err != nil {
			if err == io.EOF {
				log.Println("Соединение разорвано...")
				os.Exit(0)
			}
			log.Println(err)
			continue
		}
		bufer = bufer[0:n]
		fmt.Println(string(bufer))
	}
}
