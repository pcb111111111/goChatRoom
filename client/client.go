package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
)

//p buffer
//flag parse

func main() {
	var serverAddr = flag.String("s", "127.0.0.1:1048", "Input Server IP&PORT")

	flag.Parse()

	//net.dial 拨号 获取tcp连接
	tcpAddr, _ := net.ResolveTCPAddr("tcp", *serverAddr)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	errorCheck(err)
	go receiveMessage(conn)
	//ch控制处理流程
	for {
		reader := bufio.NewReader(os.Stdin)
		data, _ := reader.ReadString('\n')
		input := string(data)
		go sendMessage(conn, input)
	}
   
}

func receiveMessage(conn *net.TCPConn) {
	for {
		reader := bufio.NewReader(conn)
		message, err := Decode(reader)
		errorCheck(err)
		fmt.Println(message)
		if message == "bye" {
			conn.Close()
			os.Exit(0)
		}
	}

}
func errorCheck(err error) {
	if err != nil {
		fmt.Println("sss")
		fmt.Println(err)
	}
}
func sendMessage(conn *net.TCPConn, input string) {
	message, err := Encode(input)
	errorCheck(err)
	if _, err := conn.Write(message); err != nil {
		fmt.Println("Write failure" + err.Error())
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
