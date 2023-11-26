package common

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
}
