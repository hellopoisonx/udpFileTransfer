package main

import (
	"encoding/json"
	"log"
	"net"
	"udpFileTransfer/common"
)

func sendRequest(req common.Request, server string) *common.Response {
	conn, err := net.Dial("udp", server)
	if err != nil {
		log.Fatalf("Error dialing server %s %v", server, err)
	}
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

func main() {
	// server := "localhost:9992"
	// _, size := getFileInfo(server)
	// resp := sendRequest(common.Request{
	// 	Start:  0,
	// 	End:    size,
	// 	Method: "",
	// }, server)
	// content := resp.Content
	// log.Println(content)
}
