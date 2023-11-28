package common

import (
	"crypto/md5"
	"encoding/hex"
)

type Response struct {
	MD5Sum   string `json:"md5Sum"`
	Content  []byte `json:"content"`
	FileName string `json:"fileName"`
	FileSize int64  `json:"FileSize"`
	Start    int64  `json:"start"`
	End      int64  `json:"end"`
}

type Request struct {
	Start  int64  `json:"start"`
	End    int64  `json:"end"`
	Method string `json:"method"`
}

func Md5(s []byte) string {
	sum := md5.Sum(s)
	return hex.EncodeToString(sum[:])
}
