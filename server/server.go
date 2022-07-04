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

//session interface 流程 用例，即实现场景
//发不全的情况，回车换行判断读取结束
//尽量别用，for test
//forword service + saas service 环境
//fake  forward service 用于测速，支持多个

var (
	onlineConns = make(map[string]*net.TCPConn)
	leaving     = make(chan string)
	messages    = make(chan string, 1000) // all incoming client messages
)

func main() {

	addr, err := net.ResolveTCPAddr("tcp", IP+PORT)
	if err != nil {
		panic(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(err)
	}
	defer l.Close()
	fmt.Printf("tcp服务端开始监听 %s 端口...\n", PORT)
	//ch 入参
	go handleMessage()
	for {
		//Listener.Accept() 接受连接
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			return
		}
		//处理tcp请求
		addr := fmt.Sprintf("%s", conn.RemoteAddr())
		onlineConns[addr] = conn
		for i, _ := range onlineConns {
			fmt.Println(i, "is online")
		}
		go handleConnection(conn)

	}
}

//多用入参
func handleConnection(conn *net.TCPConn) {
	//一些代码逻辑...
	fmt.Println("tcp服务端开始处理请求...")

	//delete connection info and print online client
	defer func(conn net.TCPConn) {
		addr := fmt.Sprintf("%s", conn.RemoteAddr())
		fmt.Println(conn.RemoteAddr(), "has left")
		//

		delete(onlineConns, addr)
		//del
		for i := range onlineConns {
			fmt.Println("now online client:" + i)
		}
	}(*conn)
	//logrus
	for {
		//reader handle mistake
		reader := bufio.NewReader(conn)
		message, err := Decode(reader)
		errorCheck(err)
		messages <- message
		fmt.Println(message)
	}

	//write to channel
	//msgCh

	fmt.Println("tcp服务端开始处理请求完毕...")

}
func handleMessage() {
	for {
		select {
		case message := <-messages:
			doProcessMessage(message)
		case <-leaving:
			break
		default:

		}
	}
}

func doProcessMessage(message string) {
	//id 替代ip
	//# means communication，* means action
	contents := strings.Split(message, "#")
	//if 是#号分割的内容
	if len(contents) > 1 {
		addr := contents[0]
		sendMessage := strings.Join(contents[1:], "#")
		if conn, ok := onlineConns[addr]; ok {
			input, err := Encode(sendMessage)
			errorCheck(err)
			if _, err := conn.Write(input); err != nil {
				fmt.Println("online info send failure")
			}
		}
	} else { //action set
		contents := strings.Split(message, "*")
		action := contents[1]
		if strings.ToUpper(action) == "LIST" {
			addr := contents[0]

			var ips string = ""
			for i := range onlineConns {
				ips = ips + "|" + i
			}
			if conn, ok := onlineConns[addr]; ok {
				input, err := Encode(ips)
				errorCheck(err)
				if _, err := conn.Write(input); err != nil {
					fmt.Println("online info send failure")
				}
			}
		}
		//broadcast
		//sourceIp*broadcast*news
		if strings.ToUpper(action) == "BROADCAST" {
			addr := contents[0]
			sendMessage := strings.Join(contents[2:], "")
			for _, conn := range onlineConns {
				input, err := Encode(sendMessage)
				errorCheck(err)
				//源地址不广播
				if onlineConns[addr] != conn {
					if _, err := conn.Write(input); err != nil {
						fmt.Println("online info send failure")
					}
				}
			}
		}

	}

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
