package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

const (
	IP   = "127.0.0.1:"
	PORT = "1000"
)

type client chan<- string // an outgoing message channel

var (
	onlineConns = make(map[string]net.Conn)
	entering    = make(chan client)
	leaving     = make(chan client)
	messages    = make(chan string, 1024) // all incoming client messages
)

func broadcaster() {
	clients := make(map[client]bool) // all connected clients
	for {
		select {
		case msg := <-messages:
			// Broadcast incoming message to all
			// clients' outgoing message channels.
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}
func main() {
	l, err := net.Listen("tcp", IP+PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	fmt.Printf("tcp服务端开始监听 %s 端口...", PORT)

	go handleMessage()
	for {
		fmt.Println("loop test")
		//Listener.Accept() 接受连接
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		//处理tcp请求

		addr := fmt.Sprintf("%s", conn.RemoteAddr())
		onlineConns[addr] = conn
		fmt.Println(conn.RemoteAddr(), "is online")

		go handleConnection(conn)

	}
}

func handleMessage() {
	for {
		select {
		case message := <-messages:
			doProcessMessage(message)
		case <-leaving:
			break
		}
	}
}

func doProcessMessage(message string) {
	//# means communication，* means normal order
	contents := strings.Split(message, "#")
	if len(contents) > 1 {
		addr := contents[0]
		sendMessage := strings.Join(contents[1:], "#")
		addr = strings.Trim(addr, "")
		if conn, ok := onlineConns[addr]; ok {
			input, err := Encode(sendMessage)
			errorCheck(err)
			if _, err := conn.Write(input); err != nil {
				fmt.Println(err)
			}
		}
	} else {
		contents := strings.Split(message, "*")

	}

}
func handleConnection(conn net.Conn) {
	//一些代码逻辑...
	fmt.Println("tcp服务端开始处理请求...")

	//delete connection info and print online client
	defer func(conn net.Conn) {
		addr := fmt.Sprintf("%s", conn.RemoteAddr())
		delete(onlineConns, addr)
		for i := range onlineConns {
			fmt.Println("now online client:" + i)
		}
	}(conn)

	reader := bufio.NewReader(conn)
	message, err := Decode(reader)
	errorCheck(err)

	//write to channel
	messages <- message

	if _, err := conn.Write([]byte(message)); err != nil {
		fmt.Println(err)
	}
	fmt.Println("tcp服务端开始处理请求完毕...")

}
func errorCheck(err error) {
	if err != nil {
		fmt.Println("Decode failure", err.Error())
	}
}
func Encode(message string) ([]byte, error) {
	// 读取消息的长度，转换成int32类型（占4个字节）
	var length = int32(len(message))
	var pkg = new(bytes.Buffer)
	// 写入消息头
	err := binary.Write(pkg, binary.LittleEndian, length)
	if err != nil {
		return nil, err
	}
	// 写入消息实体
	err = binary.Write(pkg, binary.LittleEndian, []byte(message))
	if err != nil {
		return nil, err
	}
	return pkg.Bytes(), nil
}
func Decode(reader *bufio.Reader) (string, error) {
	// 读取消息的长度
	lengthByte, _ := reader.Peek(4) // 读取前4个字节的数据
	lengthBuff := bytes.NewBuffer(lengthByte)
	var length int32
	err := binary.Read(lengthBuff, binary.LittleEndian, &length)
	if err != nil {
		return "", err
	}
	// Buffered返回缓冲中现有的可读取的字节数。
	if int32(reader.Buffered()) < length+4 {
		return "", err
	}

	// 读取真正的消息数据
	pack := make([]byte, int(4+length))
	_, err = reader.Read(pack)
	if err != nil {
		return "", err
	}
	return string(pack[4:]), nil
}
