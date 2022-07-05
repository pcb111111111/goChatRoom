package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

//session interface 流程 用例，即实现场景
//发不全的情况，回车换行判断读取结束
//尽量别用全局变量，for test
//forword service + saas service 环境
//fake  forward service 用于测速，支持多个同时跑

type user struct {
	ip   string
	conn *net.TCPConn
}

func main() {

	var onlineConns = make(map[int]user)
	var msgCh = make(chan string, 1000) // all incoming client msgCh
	var broadCastCh = make(chan string)

	var a = flag.String("server", "127.0.0.1:1048", "Input Server IP&PORT")
	var maxCapacity = flag.Int("capacity", 1000, "max capacity of chatroom")
	flag.Parse()
	addr, err := net.ResolveTCPAddr("tcp", *a)
	if err != nil {
		panic(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(err)
	}
	defer l.Close()
	fmt.Printf("tcp服务端开始监听 %s...\n", addr)

	go handleMessage(msgCh, broadCastCh, &onlineConns)

	for {
		//Listener.Accept()
		conn, err := l.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			return
		}

		var u user
		id := rand.Int() % *maxCapacity
		addr := fmt.Sprintf("%s", conn.RemoteAddr())
		u.ip = addr
		u.conn = conn
		onlineConns[id] = u
		for i, _ := range onlineConns {
			fmt.Println(i, "is online")
		}
		go handleConnection(conn, &msgCh, &broadCastCh, &onlineConns, id)
	}
}

//多用入参
func handleConnection(conn *net.TCPConn, msgCh *chan string, broadcastCh *chan string, onlineConns *map[int]user, id int) {
	//一些代码逻辑...
	fmt.Println("tcp服务端开始处理请求...")

	//delete connection info and print online client
	defer func(conn net.TCPConn) {
		fmt.Println(id, conn.RemoteAddr(), "has left")
		delete(*onlineConns, id)
		conn.Close()
	}(*conn)
	//logrus
	for {
		//reader handle mistake
		reader := bufio.NewReader(conn)
		message, err := Decode(reader)
		errorCheck(err)
		if strings.ToUpper(message) == "LIST" {
			var ips string = ""
			for i := range *onlineConns {
				ips = ips + strconv.Itoa(i) + ":" + (*onlineConns)[i].ip + "|"
			}
			input, err := Encode(ips)
			errorCheck(err)
			if _, err := conn.Write(input); err != nil {
				fmt.Println("online info send failure")
			}
		} else if strings.ToUpper(message) == "QUIT" {
			input, err := Encode("bye")
			errorCheck(err)
			if _, err := conn.Write(input); err != nil {
				fmt.Println("online info send failure")
			}
			delete(*onlineConns, id)
			break
		} else if str := strings.ToUpper(message); str != "" && str[0:4] == "@ALL" {
			*broadcastCh <- message
		} else if str := strings.ToUpper(message); str != "" && str[0] == '@' {
			*msgCh <- message
		} else {
			input, err := Encode("wrong action,please input again!")
			errorCheck(err)
			if _, err := conn.Write(input); err != nil {
				fmt.Println("online info send failure")
			}
		}

	}

	//write to channel
	//msgCh

	fmt.Println("tcp服务端开始处理请求完毕...")

}
func handleMessage(msgCh chan string, broadcastCh chan string, onlineConns *map[int]user) {
	for {
		select {
		case message := <-msgCh:
			doProcessMessage(message, onlineConns)
		case message := <-broadcastCh:
			doBroadcast(message, onlineConns)
		}
	}
}

func doBroadcast(message string, onlineconns *map[int]user) {
	contents := strings.Split(message, ":")
	if len(contents) > 1 {
		for i := range *onlineconns {
			input, err := Encode(contents[1])
			errorCheck(err)
			if _, err := (*onlineconns)[i].conn.Write(input); err != nil {
				fmt.Println("online info send failure")
			}
		}
	}
}
func doProcessMessage(message string, onlineConns *map[int]user) {
	//id 替代ip
	//# means communication，* means action
	contents := strings.Split(message, ":")
	//带有用户名的命令
	if len(contents) > 1 {
		tmp := strings.Split(contents[0], "@")
		desId, _ := strconv.Atoi(tmp[1])
		conn := (*onlineConns)[desId].conn
		input, err := Encode(contents[1])
		errorCheck(err)
		if _, err := conn.Write(input); err != nil {
			fmt.Println("online info send failure")
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
