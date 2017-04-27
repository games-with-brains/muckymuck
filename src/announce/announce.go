package main

import {
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//	announce - sits listening on a port, and whenever anyone connects announces a message and disconnects them
//
//	Usage:  announce [port] < message_file
//
//	Author: Eleanor McHugh <eleanor@games-with-brains.com>

const (
	PORT = ":4201"
	PROTOCOL = "tcp"
)

var l log.Logger

func main() {
	var m string
	for s := bufio.NewScanner(os.Stdin); s.Scan(); {
		m += s.Text() + "\r\n"
	}

	l = log.New(os.Stderr, "", 1)
	signal.Ignore(syscall.SIGHUP)
	Listen(PROTOCOL, PORT, func(c net.Conn) {
		defer c.Close()
		l.Println("CONNECTION made from ", host)
		fmt.Fprintln(c, m)
	})
}

func Listen(p, a string, f func(net.Conn)) (e error) {
	var listener net.Listener
	if listener, e = net.Listen(p, a); e == nil {
		for {
			if connection, e := listener.Accept(); e == nil {
				go f(connection)
			}
		}
	} else {
		l.Fatal(e)
	}
	return
}