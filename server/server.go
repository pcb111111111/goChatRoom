package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"net"
	"os"
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
	var maxCapacity = flag.Int("capacity", 1000, "Max Capacity Of Chatroom")
	var logDirectionary = flag.String("logP", "./server/log/chatLog.log", "Log")

	flag.Parse()
	file := InitLog(*logDirectionary)
	defer file.Close()
	addr, err := net.ResolveTCPAddr("tcp", *a)
	if err != nil {
		panic(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		doLogErr(err)
	}
	defer l.Close()
	fmt.Printf("tcp服务端开始监听 %s...\n", addr)

	go handleMessage(msgCh, broadCastCh, &onlineConns)

	for {
		//Listener.Accept()
		conn, err := l.AcceptTCP()
		if err != nil {
			doLogErr(err)
			return
		}

		var u user
		id := rand.Int() % *maxCapacity
		addr := fmt.Sprintf("%s", conn.RemoteAddr())
		u.ip = addr
		u.conn = conn
		onlineConns[id] = u

		//record online info in logfile
		var msg string = ""
		msg = msg + strconv.Itoa(id) + "|" + addr + " " + "is online"

		doLogInfo(msg)
		broadCastCh <- msg

		go handleConnection(conn, &msgCh, &broadCastCh, &onlineConns, id)
	}
}

//多用入参
func handleConnection(conn *net.TCPConn, msgCh *chan string, broadcastCh *chan string, onlineConns *map[int]user, id int) {
	//一些代码逻辑...
	fmt.Println("tcp服务端开始处理请求...")

	defer func(conn net.TCPConn) {
		addr := fmt.Sprintf("%s", conn.RemoteAddr())
		var msg string = ""
		msg = msg + strconv.Itoa(id) + ":" + addr + " " + "has left"
		//Log
		doLogInfo(msg)
		conn.Close()
		//broadcast leaving to all online user
		*broadcastCh <- msg
	}(*conn)
	//logrus
	var msg string = ""
	for {
		//reader handle mistake
		reader := bufio.NewReader(conn)
		message, err := Decode(reader)
		doLogErr(err)
		msg += message
		if msg[len(msg)-1:len(msg)] != "\n" {
			continue
		}
		if strings.ToUpper(message) == "LIST" {
			var ips string = ""
			for i := range *onlineConns {
				ips = ips + strconv.Itoa(i) + ":" + (*onlineConns)[i].ip + "|"
			}
			input, err := Encode(ips)
			doLogErr(err)
			if _, err := conn.Write(input); err != nil {
				doLogErr(err)
			}
			doLogInfo("LIST" + ":" + ips)
		} else if strings.ToUpper(message) == "QUIT" {
			input, err := Encode("bye")
			doLogErr(err)
			if _, err := conn.Write(input); err != nil {
				doLogErr(err)
			}

			delete(*onlineConns, id)

			break
		} else if str := strings.ToUpper(message); str != "" && str[0:4] == "@ALL" {
			*broadcastCh <- message[4:]
		} else if str := strings.ToUpper(message); str != "" && str[0] == '@' {
			*msgCh <- message
		} else {
			input, err := Encode("wrong action,please input again!")
			doLogErr(err)
			if _, err := conn.Write(input); err != nil {
				doLogErr(err)
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
	if len(message) > 1 {
		for i := range *onlineconns {
			input, err := Encode(message)
			doLogErr(err)
			if _, err := (*onlineconns)[i].conn.Write(input); err != nil {
				doLogErr(err)
			}
		}
	}
	doLogInfo("BroadCast" + message)
}
func doProcessMessage(message string, onlineConns *map[int]user) {
	//id 替代ip
	//# means communication，* means action
	contents := strings.Split(message, " ")
	//带有用户名的命令
	if len(contents) > 1 {
		tmp := strings.Split(contents[0], "@")
		desId, _ := strconv.Atoi(tmp[1])
		conn := (*onlineConns)[desId].conn
		input, err := Encode(contents[1])
		doLogErr(err)
		if _, err := conn.Write(input); err != nil {
			doLogErr(err)
		}
		dstIp := fmt.Sprintf("%s", conn.RemoteAddr())
		doLogInfo("1v1 msg to" + contents[0] + ":" + dstIp)
	}

}
func InitLog(logPath string) *os.File {
	//设置输出样式，自带的只有两种样式logrus.JSONFormatter{}和logrus.TextFormatter{}
	log.SetFormatter(&log.TextFormatter{})
	//设置output,默认为stderr,可以为任何io.Writer，比如文件*os.File
	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	//同时写文件和屏幕
	fileAndStdoutWriter := io.MultiWriter(file)
	if err == nil {
		log.SetOutput(fileAndStdoutWriter)
	} else {
		log.Info("failed to log to file.")
	}
	return file
}

//错误行数
func doLogErr(err error) {
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("A Error appears")
	}
}
func doLogInfo(msg string) {
	log.WithFields(log.Fields{
		"Info": msg,
	}).Info("A Info appears")

}
func Encode(message string) ([]byte, error) {
	// 读取消息的长度，转换成int32类型（占4个字节）
	var length = int32(len(message))
	var pkg = new(bytes.Buffer)
	// 写入消息头
	err := binary.Write(pkg, binary.LittleEndian, length)
	if err != nil {
		doLogErr(err)
	}
	// 写入消息实体
	err = binary.Write(pkg, binary.LittleEndian, []byte(message))
	if err != nil {
		doLogErr(err)
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
		doLogErr(err)
	}
	// Buffered返回缓冲中现有的可读取的字节数。
	if int32(reader.Buffered()) < length+4 {
		return "", err
	}

	// 读取真正的消息数据
	pack := make([]byte, int(4+length))
	_, err = reader.Read(pack)
	if err != nil {
		doLogErr(err)
	}
	return string(pack[4:]), nil
}
