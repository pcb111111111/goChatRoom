package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

//p buffer
//flag parse
const NUM_CONN = 10
const IP = "127.0.0.1:"
const PORT = "100"
const SIZE_MESSAGE = 1024

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

func createConn(ch chan int) {
	conn, err := net.Dial("tcp", IP+PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	input := make([]byte, SIZE_MESSAGE)

	if _, err := conn.Write(input); err != nil {
		fmt.Println(err)
	}
	fmt.Println("test server")
	//读取到buffer
	buffer := make([]byte, SIZE_MESSAGE)
	//如果服务端没有把数据传递过来，那么客户端阻塞，直到服务端向其中写入了数据。
	if _, err := conn.Read(buffer); err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(buffer[:]))
	fmt.Printf("获取%s%s的tcp连接成功...\n", IP, PORT)
	ch <- 1
}
func main() {
	//net.dial 拨号 获取tcp连接
	t := time.Now()
	//semaphore block
	ch := make(chan int)
	for i := 0; i < NUM_CONN; i++ {
		go createConn(ch)
	}
	elapsed := time.Since(t)
	fmt.Println(elapsed)
	<-ch
}
