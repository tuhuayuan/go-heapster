package models

import (
	"encoding/json"
	"math/rand"
	"time"
)

// SerialNumber 序列号
type SerialNumber string

// NewSerialNumber 创建一个唯一序列号
func NewSerialNumber() SerialNumber {
	return SerialNumber(randStringBytesMaskImprSrc(12))
}

// MarshalJSON json编码实现
func (sn SerialNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(sn))
}

// UnmarshalJSON json解码实现
func (sn *SerialNumber) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return nil
	}
	*sn = SerialNumber(str)
	return nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}
