package util

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GenerateSignature 生成API签名
func GenerateSignature(accessKey string, accessKeySecret string, appCode string) (string, int64) {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// 将签名参数按照字典序排序
	params := map[string]string{
		"accessKey":       accessKey,
		"accessKeySecret": accessKeySecret,
		"appCode":         appCode,
		"timestamp":       strconv.FormatInt(timestamp, 10),
	}
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接排序后的参数
	var signStr string
	for _, k := range keys {
		signStr += k + params[k]
	}

	// 计算签名
	md5Bytes := md5.Sum([]byte(signStr))
	sign := hex.EncodeToString(md5Bytes[:])
	sign = strings.ToUpper(sign)

	return sign, timestamp
}
