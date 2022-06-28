package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const IP = "127.0.0.1:"
const PORT = "100"

//地址变成arg
//格式化输出，更改传输数据的尺寸，
//
//goroutinuecheck模式 ?
//连接错误处理
//连到公司服务器
//
// Decode 解码消息
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
func main() {
	// tcp 监听并接受端口
	l, err := net.Listen("tcp", IP+PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	//最后关闭
	defer l.Close()
	fmt.Printf("tcp服务端开始监听 %s 端口...", PORT)
	// 使用循环一直接受连接
	for {
		fmt.Println("loop test")
		//Listener.Accept() 接受连接
		//conn 是双方的。和长度为1的channel有些类似。
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		//处理tcp请求
		go handleConnection(c)
	}
}
func handleConnection(c net.Conn) {
	//一些代码逻辑...
	fmt.Println("tcp服务端开始处理请求...")
	//读取
	//reader = bufio.NewReader(c)
	buffer := make([]byte, 1024)
	//如果客户端无数据则会阻塞，服务端阻塞，直到等待客户端传递数据。
	if _, err := c.Read(buffer); err != nil {
		fmt.Println(err)
	}
	fmt.Println(c.RemoteAddr())
	//服务端成功从阻塞状态走出，读取客户端的数据，并根据自身的接口输出buffer
	//mistake handle
	if _, err := c.Write(buffer); err != nil {
		fmt.Println(err)
	}
	fmt.Println("tcp服务端开始处理请求完毕...")

}
