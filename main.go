/*
* @Description:
* @Author: Sauron
* @Date: 2022-06-25 17:05:20
 * @LastEditTime: 2022-06-26 19:08:05
 * @LastEditors: Sauron
*/
package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"encoding/binary"
	"flag"
	"io"
	"net"

	"github.com/wutianze/nats.go"
)

var host_port = flag.String("socket listen address", ":8000", "listen on which port")
var nats_address = flag.String("nats", "nats://39.101.140.145:4222", "address of nats server")
var link_num = flag.Int("num", 3, "number of clients(for server) or index of the client(for client)")
var debug = flag.Bool("debug", false, "run as a socket client")

// name: server, cli0, cli1, cli2
var name = flag.String("name", "", "who am I")

func sendMsg(n *net.Conn, b *[]byte) {
	var dataLen []byte = make([]byte, 4)
	binary.BigEndian.PutUint32(dataLen, uint32(len(*b)))
	(*n).Write(dataLen)
	(*n).Write(*b)
}

func main() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		os.Exit(0)
	}()

	flag.Parse()
	if *debug {
		test_conn, err_test := net.Dial("tcp", *host_port)
		if err_test != nil {
			fmt.Println(err_test)
			return
		}
		go func() {
			for {
				bufSize := make([]byte, 4)
				_, err0 := io.ReadFull(test_conn, bufSize)
				if err0 != nil {
					fmt.Println(err0)
					return
				}
				bufSizeUint := binary.BigEndian.Uint32(bufSize)
				data := make([]byte, bufSizeUint)
				_, err1 := io.ReadFull(test_conn, data)
				if err1 != nil {
					fmt.Println(err1)
					return
				}
				if string(data) == "exit" {
					return
				}
				fmt.Println("socket client recv:" + string(data))
			}
		}()
		var to_send []byte = []byte("ff")
		sendMsg(&test_conn, &to_send)
		time.Sleep(time.Duration(1) * time.Second)
		to_send = []byte("gg")
		sendMsg(&test_conn, &to_send)
		time.Sleep(time.Duration(1) * time.Second)
		to_send = []byte("exit")
		sendMsg(&test_conn, &to_send)
		time.Sleep(time.Duration(3) * time.Second)
		return
	}

	tcpServer, err0 := net.ResolveTCPAddr("tcp4", *host_port)
	if err0 != nil {
		fmt.Println(err0)
		return
	}
	listener, err1 := net.ListenTCP("tcp", tcpServer)
	if err1 != nil {
		fmt.Println(err1)
		return
	}

	nc, err2 := nats.IConnect(*nats_address)
	defer nc.IClose()
	if err2 != nil {
		fmt.Println(err2)
		return
	}

	switch {
	case *name == "server":
		for i := 0; i < *link_num; i++ {

			socket_conn, err2 := listener.Accept()
			defer socket_conn.Close()
			if err2 != nil {
				fmt.Println(err2)
				continue
			}
			go func() {
				for {
					bufSize := make([]byte, 4)
					_, err0 := io.ReadFull(socket_conn, bufSize)
					if err0 != nil {
						fmt.Println(err0)
						return
					}
					bufSizeUint := binary.BigEndian.Uint32(bufSize)
					data := make([]byte, bufSizeUint)
					_, err1 := io.ReadFull(socket_conn, data)
					if err1 != nil {
						fmt.Println(err1)
						return
					}
					if string(data) == "exit" {
						return
					}
					nc.IPublish("gtcontrol_"+strconv.Itoa(i), data)
				}
			}()
			_, err3 := nc.ISubscribe("gtlog_"+strconv.Itoa(i), func(m *nats.Msg) {
				fmt.Printf("Received a message: %s\n", string(m.Data))
				sendMsg(&socket_conn, &m.Data)
			})
			if err3 != nil {
				fmt.Println(err3)
			}
		}

	case *name == "client":
		socket_conn, err2 := listener.Accept()
		defer socket_conn.Close()
		if err2 != nil {
			fmt.Println(err2)
			return
		}
		go func() {
			for {
				bufSize := make([]byte, 4)
				_, err0 := io.ReadFull(socket_conn, bufSize)
				if err0 != nil {
					fmt.Println(err0)
					return
				}
				bufSizeUint := binary.BigEndian.Uint32(bufSize)
				data := make([]byte, bufSizeUint)
				_, err1 := io.ReadFull(socket_conn, data)
				if err1 != nil {
					fmt.Println(err1)
					return
				}
				if string(data) == "exit" {
					return
				}
				nc.IPublish("gtlog_"+strconv.Itoa(*link_num), data)
			}
		}()
		_, err3 := nc.ISubscribe("gtcontrol_"+strconv.Itoa(*link_num), func(m *nats.Msg) {
			fmt.Printf("Received a message: %s\n", string(m.Data))
			sendMsg(&socket_conn, &m.Data)
		})
		if err3 != nil {
			fmt.Println(err3)
		}
		fmt.Println("cli0")
	}
	for {
	}

}

//nc.IPublish("gt_log",[]byte("what"))
//time.Sleep(time.Duration(5)*time.Second)
