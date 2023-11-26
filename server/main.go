package main

import (
	"crypto/md5"
	"encoding/json"
	"fileTransfer/common"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func main() {
	args := os.Args
	if len(args) <= 1 {
		log.Fatalln("参数错误,请重试:./server xxx.xxx.xxx.xxx:xxxx xxxx.img xxx.txt ...")
	}
	server := args[1]
	filePath := args[2:]
	log.Println(server)
	addr, err := net.ResolveUDPAddr("udp", server)
	if err != nil {
		log.Fatalf("解析%s失败 %v", server, err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("监听%s失败 %v", server, err)
	}
	f, _ := os.OpenFile(filePath[0], os.O_RDONLY, os.ModePerm)
	for {
		var buf = make([]byte, 1024)
		// 读取客户端请求
		n, addr, err := conn.ReadFromUDP(buf[0:])
		go handleRequest(f, conn, addr, buf, n, err, filePath[0])
	}
}

func handleRequest(f *os.File, conn *net.UDPConn, addr *net.UDPAddr, buf []byte, n int, err error, name string) {
	if err != nil {
		log.Printf("发送%s失败,跳过 %s\n", name, err)
		return
	}
	fileInfo, err := f.Stat()
	if err != nil {
		log.Printf("获取文件大小失败 %v\n", err)
		return
	}
	fileSize := fileInfo.Size()
	fileName := fileInfo.Name()
	// 处理请求
	var fileReq common.FileRequest
	err = json.Unmarshal(buf[:n], &fileReq)
	if err != nil {
		log.Printf("解析请求至json失败 %v\n", err)
		return
	}

	fileResp := &common.FileResponse{}
	if fileReq.Start == common.MaxLen { // 当start == maxlen时代表客户端请求文件大小和名称
		fileResp = &common.FileResponse{
			Content: []byte(fileName),
			Start:   fileSize,
			End:     fileSize,
		}
	} else { // 客户端请求文件内容
		f.Seek(fileReq.Start, 0) //从start开始读取
		content := make([]byte, fileReq.End-fileReq.Start)
		_, err = f.Read(content[:])
		hasher := md5.New()
		io.WriteString(hasher, string(content))
		md5hash := fmt.Sprintf("%x", hasher.Sum(nil))
		if err != nil {
			log.Printf("读取文件失败 %v\n", err)
			return
		}
		fileResp = &common.FileResponse{
			Content: content,
            MD5Hash: md5hash,
			Start:   fileReq.Start,
			End:     fileReq.End,
		}
	}
	resp, err := json.Marshal(fileResp)
	if err != nil {
		log.Printf("解析响应至json失败 %v\n", err)
		return
	}
	log.Println(string(resp))
	_, err = conn.WriteToUDP(resp, addr)
	if err != nil {
		log.Printf("发送响应至udp连接失败 %v\n", err)
		return
	}
	log.Printf("发送成功 start:%d end:%d", fileResp.Start, fileResp.End)
}
