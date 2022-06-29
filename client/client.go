package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

//p buffer
//flag parse
const (
	NUM_CONN     = 10
	SIZE_MESSAGE = 1024
)

func main() {
	t := time.Now()
	elapsed := time.Since(t)
	//var clientAddr = flag.String("s", "127.0.0.0:1000", "Input Server IP&PORT")

	//flag.Parse()

	//net.dial 拨号 获取tcp连接
	conn, err := net.Dial("tcp", "127.0.0.1:1000")
	checkError(err)
	defer conn.Close()

	sendMessage(conn)

	buf := make([]byte, 1024)
	for {
		_, err := conn.Read(buf)
		checkError(err)
		fmt.Println("receive from server", string(buf))
	}

	fmt.Println("Online Time is ", elapsed)

}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
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

func sendMessage(conn net.Conn) {
	for {
		reader := bufio.NewReader(os.Stdin)
		data, _, _ := reader.ReadLine()
		input := string(data)

		if strings.ToUpper(input) == "EXIT" {
			conn.Close()
			os.Exit(-1)
			break
		}

		_, err := conn.Write([]byte(input))
		if err != nil {
			conn.Close()
			fmt.Println("Write failure" + err.Error())
		}
	}

}
