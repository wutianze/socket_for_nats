/*
* @Description:
* @Author: Sauron
* @Date: 2022-06-25 17:05:20
 * @LastEditTime: 2022-06-26 21:30:31
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

var host_port = flag.String("address", ":8000", "listen address, connect address(debug)")
var nats_address = flag.String("nats", "nats://39.101.140.145:4223", "address of nats server")
var link_num = flag.Int("num", 3, "number of clients(for server) or index of the client(for client)")
var debug = flag.Bool("debug", false, "run as a socket client")
var name = flag.String("name", "", "who am I")

func sendMsg(n *net.Conn, b *[]byte) {
	var dataLen []byte = make([]byte, 4)
	binary.LittleEndian.PutUint32(dataLen, uint32(len(*b)))
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
				bufSizeUint := binary.LittleEndian.Uint32(bufSize)
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

	

	nc, err2 := nats.IConnect(*nats_address)
	defer nc.IClose()
	if err2 != nil {
		fmt.Println(err2)
		return
	}

	switch {
	case *name == "server":
		fmt.Println("server")
		for i := 0; i < *link_num; i++ {
			
			tcpServer, err0 := net.ResolveTCPAddr("tcp4", ":"+strconv.Itoa(8000+i))
	if err0 != nil {
		fmt.Println(err0)
		return
	}
	listener, err1 := net.ListenTCP("tcp", tcpServer)
	if err1 != nil {
		fmt.Println(err1)
		return
	}
			socket_conn, err2 := listener.Accept()
			defer socket_conn.Close()
			if err2 != nil {
				fmt.Println(err2)
				i--
				continue
			}
			fmt.Println("socket connected")
			go func() {
				client_index := i
				for {
					fmt.Println("start receive")
					bufSize := make([]byte, 4)
					_, err0 := io.ReadFull(socket_conn, bufSize)
					if err0 != nil {
						fmt.Println(err0)
						return
					}
					bufSizeUint := binary.LittleEndian.Uint32(bufSize)
					fmt.Println("msg size is:",bufSizeUint)
					data := make([]byte, bufSizeUint)
					_, err1 := io.ReadFull(socket_conn, data)
					if err1 != nil {
						fmt.Println(err1)
						return
					}
					if string(data) == "exit" {
						return
					}
					fmt.Printf("Socket Received a message: %s\n", string(data))
					nc.IPublish("gtcontrol_"+strconv.Itoa(client_index), data)
				}
			}()
			_, err3 := nc.ISubscribe("gtlog_"+strconv.Itoa(i), func(m *nats.Msg) {
				fmt.Printf("Nats Received a message: %s\n", string(m.Data))
				sendMsg(&socket_conn, &m.Data)
			})
			if err3 != nil {
				fmt.Println(err3)
			}
		}
		for {
	}

	case *name == "client":
		fmt.Println("client")
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
		socket_conn, err2 := listener.Accept()
		defer socket_conn.Close()
		if err2 != nil {
			fmt.Println(err2)
			return
		}
		go func() {
			for {
				fmt.Println("start receive")
				bufSize := make([]byte, 4)
				_, err0 := io.ReadFull(socket_conn, bufSize)
				if err0 != nil {
					fmt.Println(err0)
					return
				}
				bufSizeUint := binary.LittleEndian.Uint32(bufSize)
				fmt.Println("msg size is:",bufSizeUint)
				data := make([]byte, bufSizeUint)
				_, err1 := io.ReadFull(socket_conn, data)
				if err1 != nil {
					fmt.Println(err1)
					return
				}
				if string(data) == "exit" {
					return
				}
				fmt.Printf("Socket Received a message: %s\n", string(data))
				nc.IPublish("gtlog_"+strconv.Itoa(*link_num), data)
			}
		}()
		_, err3 := nc.ISubscribe("gtcontrol_"+strconv.Itoa(*link_num), func(m *nats.Msg) {
			fmt.Printf("NATS Received a message: %s\n", string(m.Data))
			sendMsg(&socket_conn, &m.Data)
		})
		if err3 != nil {
			fmt.Println(err3)
		}
		for {
	}
	}
}

//nc.IPublish("gt_log",[]byte("what"))
//time.Sleep(time.Duration(5)*time.Second)
