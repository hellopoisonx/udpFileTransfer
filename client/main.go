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
	"sync"
	"time"
)

const FileBlockSize = 1024
const Concurrency = 20

var fileLock sync.Mutex
var wg sync.WaitGroup
var totalWrote = int64(0)

func sendRequest(fileReq common.FileRequest, server string) common.FileResponse {
	conn, err := net.Dial("udp", server)
	if err != nil {
		log.Fatalf("Error dialing UDP: %v", err)
	}
	defer conn.Close()

	req, err := json.Marshal(fileReq)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	_, err = conn.Write(req)
	if err != nil {
		log.Fatalf("Error writing to UDP connection: %v", err)
	}

	var buf [4096]byte
	n, err := conn.Read(buf[0:])
	if err != nil {
		log.Fatalf("Error reading from UDP connection: %v", err)
	}

	var fileResp common.FileResponse
	err = json.Unmarshal(buf[0:n], &fileResp)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	return fileResp

}

func getFileInfo(server string) (string, int64) {
	log.Println("Requesting file size")
	fileReq := common.FileRequest{Start: common.MaxLen, End: common.MaxLen}
	response := make(chan common.FileResponse, 1)
	for {
		timer := time.NewTimer(5 * time.Second)
		go func() {
			data := sendRequest(fileReq, server)
			response <- data
		}()

		select {
		case <-timer.C:
			continue
		case result := <-response:
			log.Println("接收成功")
			return string(result.Content), result.Start
		}
	}
}

func getFileBlock(region, totalBlock int64, file *os.File, server string) {
	for blockId := region; blockId < totalBlock; blockId += Concurrency {
		startFile, endFile := blockId*FileBlockSize, (blockId+1)*FileBlockSize
		data := getFileContent(startFile, endFile, server)
		go saveFileBlock(startFile, endFile, totalBlock, data, file)
		log.Printf("Requested file block finished: %d\n", blockId)
	}
}

func getFileContent(start, end int64, server string) []byte {
	fileReq := common.FileRequest{Start: start, End: end}
	responseChan := make(chan common.FileResponse, 1)
	for {
		timer := time.NewTimer(time.Duration(time.Millisecond * 500))
		go func() {
			data := sendRequest(fileReq, server)
			responseChan <- data
		}()
		select {
		case res := <-responseChan:
			hasher := md5.New()
			io.WriteString(hasher, string(res.Content))
			md5hash := fmt.Sprintf("%x", hasher.Sum(nil))
			if md5hash != res.MD5Hash {
                log.Println("md5sum校验失败")
				continue
			}
			return res.Content
		case <-timer.C:
			// return getFileContent(start, end, server)
			continue
		}
	}
}
func saveFileBlock(start, end, totalBlock int64, content []byte, f *os.File) {
	fileLock.Lock()
	defer fileLock.Unlock()
	f.Seek(start, 0)
	f.Write(content)
	totalWrote++
	if totalWrote == totalBlock {
		wg.Done()
	}
	log.Printf("保存数据 %d 到 %d\n", start, end)
}
func saveFile(savePath, server string) {
	for {
		fileName, fileSize := getFileInfo(server)
		log.Println(fileName, " ", fileSize)
		f, err := os.OpenFile(savePath+fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		log.Println("打开文件成功")
		if err != nil {
			log.Printf("创建文件失败, 跳过接收%s\n", fileName)
			continue
		}
		totalBlock := fileSize / FileBlockSize
		wg.Add(1)
		blockAmount := min(totalBlock, Concurrency)
		log.Println("")
		if blockAmount != 0 {
			for i := int64(0); i < blockAmount; i++ {
				log.Println("请求块文件")
				go getFileBlock(i, totalBlock, f, server)
			}
			wg.Wait()
		}
		if remains := fileSize % FileBlockSize; remains != 0 {
			offset := fileSize - remains
			content := getFileContent(offset, fileSize, server)
			saveFileBlock(offset, fileSize, totalBlock, content, f)
		}
		log.Printf("保存文件%s成功\n", fileName)
		break
	}
}
func waitForServer(server string) {
	for {
		_, err := net.Dial("udp", server)
		if err == nil {
			log.Println("服务器已启动")
			return
		}
		log.Println("等待服务器启动...")
		time.Sleep(1 * time.Second)
	}
}
func main() {
	args := os.Args
	filePath := ""
	server := ""
	switch len(args) {
	case 1:
		log.Fatalln("参数错误 ./client xxx xxx.xxx.xxx.xxx:xxxx")
	case 2:
		server = args[1]
	case 3:
		filePath = args[1] + "/"
		server = args[2]
	}
	waitForServer(server)
	saveFile(filePath, server)
}
