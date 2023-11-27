package common

<<<<<<< HEAD
import "math"

const MaxLen = math.MaxInt64

type FileRequest struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

type FileResponse struct {
	Content []byte `json:"content"`
	MD5Hash string `json:"md5hash"`
	Start   int64  `json:"start"`
	End     int64  `json:"end"`
=======
import (
	"crypto/md5"
	"encoding/hex"
)

type Response struct {
	MD5Sum   string `json:"md5Sum"`
	Content  string `json:"content"`
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
>>>>>>> 45123f3 (first commit)
}
