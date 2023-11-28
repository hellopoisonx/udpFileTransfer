package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
	"udpFileTransfer/common"
)

const (
	Concurrency = 50
	BlockSize   = 1024
	TTL         = 300 //(ms)
)

var wg sync.WaitGroup
var FileLock sync.RWMutex
var writtenBlock int64 = 0

func sendRequest(req common.Request, server string) *common.Response {
	conn, err := net.Dial("udp", server)
	if err != nil {
		log.Fatalf("Error dialing server %s %v", server, err)
	}
	defer conn.Close()
	log.Println("Sending request")
	re, err := json.Marshal(req)
	if err != nil {
		log.Fatalf("Error marshaling json %v", err)
	}
	_, err = conn.Write(re)
	if err != nil {
		log.Fatalf("Error sending request %v", err)
	}
	log.Println("Getting response")
	buf := make([]byte, 4096)
	n, err := conn.Read(buf[:])
	if err != nil {
		log.Fatalf("Error reading from udp connection %v", err)
	}
	var resp common.Response
	err = json.Unmarshal(buf[:n], &resp)
	if err != nil {
		log.Fatalf("Error unmarshaling json %v", err)
	}
	return &resp
}
func getFileInfo(server string) (string, int64) {
	req := common.Request{
		Start:  0,
		End:    0,
		Method: "getFileInfo",
	}
	resp := sendRequest(req, server)
	fileName := resp.FileName
	fileSize := resp.FileSize
	return fileName, fileSize
}
func getBlock(region, totalBlock int64, server string, file *os.File) {
	for i := region; i < totalBlock; i += Concurrency {
		content := getBlockContent(common.Request{
			Start:  i * BlockSize,
			End:    (i + 1) * BlockSize,
			Method: "",
		}, server)
		go saveBlock(i*BlockSize, (i+1)*BlockSize, totalBlock, content, file)
	}
}
func getBlockContent(req common.Request, server string) []byte {
	respChan := make(chan *common.Response, 1)
	for {
		timer := time.NewTimer(TTL * time.Millisecond)
		go func() {
			res := sendRequest(req, server)
			respChan <- res
		}()
		select {
		case resp := <-respChan:
			content := resp.Content
			md5Sum := common.Md5(content)
			if md5Sum != resp.MD5Sum {
				log.Println("Error checking file, md5 not matched, retrying")
				return getBlockContent(req, server)
			}
			return content
		case <-timer.C:
			log.Println("Error requesting file, out of time, retrying")
			return getBlockContent(req, server)
		}
	}
}
func saveBlock(start, end, totalBlock int64, content []byte, file *os.File) {
	FileLock.Lock()
	defer FileLock.Unlock()
	file.Seek(start, 0)
	file.Write(content)
	writtenBlock++
	if writtenBlock == totalBlock {
		wg.Done()
	}
	log.Printf("Succeed saving block from %d to %d ", start, end)
}
func saveFile(pathPrefix, server string) {
	fileName, fileSize := getFileInfo(server)
	fileName = fmt.Sprintf("%s/%s", pathPrefix, fileName)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Printf("Error creating new file(%s), skip %v", fileName, err)
		return
	}
	defer file.Close()
	totalBlock := fileSize / BlockSize
	if totalBlock < 1 {
		content := getBlockContent(common.Request{
			Start:  0,
			End:    fileSize,
			Method: "",
		}, server)
		file.Write(content)
		log.Printf("Succeed saving file(%s)", fileName)
		return
	}
	size := min(fileSize, Concurrency)
	wg.Add(1)
	for i := int64(0); i < size; i++ {
		go getBlock(i, totalBlock, server, file)
	}
	wg.Wait()
	if fileSize%BlockSize != 0 {
		offset := fileSize - fileSize%BlockSize
		content := getBlockContent(common.Request{
			Start:  offset,
			End:    fileSize,
			Method: "",
		}, server)
		saveBlock(offset, fileSize, totalBlock, content, file)
	}
	log.Printf("Succeed saving file(%s)", fileName)
}
func main() {
	for {
		saveFile("./", "localhost:9992")
	}
}
