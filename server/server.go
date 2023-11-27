package main

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"udpFileTransfer/common"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatalln("Error reading files")
	}
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:9992")
	if err != nil {
		log.Fatalf("Error resolving address %s %v", "0.0.0.0:9992", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Error listening address %s %v", "0.0.0.0:9992", err)
	}
	for _, fileName := range args[1:] {
		file, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
		if err != nil {
			log.Printf("Error opening %s, skip %v", fileName, err)
			continue
		}
		for {
			var buf [4096]byte
			n, addr, err := conn.ReadFromUDP(buf[:])
			if err != nil {
				log.Printf("Error reading from udp %v", err)
				break
			}
			handleRequest(n, addr, conn, buf[:], file)
		}
	}
}

func handleRequest(n int, addr *net.UDPAddr, conn *net.UDPConn, buf []byte, file *os.File) {
	var req common.Request
	var err error
	err = json.Unmarshal(buf[:n], &req)
	if err != nil {
		log.Printf("Error unmarshaling json, skip %v", err)
		return
	}
	resp := &common.Response{}
	if req.Method == "getFileInfo" {
		fileInfo, err1 := file.Stat()
		if err != nil {
			log.Printf("Error getting file info, skip %v", err1)
			return
		}
		fileName := fileInfo.Name()
		fileSize := fileInfo.Size()
		resp = &common.Response{
			MD5Sum:   "",
			Content:  "",
			FileName: fileName,
			FileSize: fileSize,
			Start:    0,
			End:      0,
		}
	} else {
		file.Seek(req.Start, 0)
		data := make([]byte, req.End-req.Start)
		n, err1 := file.Read(data[:])
		if err1 != nil {
			log.Printf("Error reading from file, skip %v", err1)
		}
		resp = &common.Response{
			MD5Sum:   common.Md5(data[:n]),
			Content:  string(data[:n]),
			FileName: "",
			FileSize: 0,
			Start:    req.Start,
			End:      req.End,
		}
	}
	res, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshaling json %v", err)
		return
	}
	_, err = conn.WriteToUDP(res, addr)
	if err != nil {
		log.Printf("Error responding %v", err)
	}
}
