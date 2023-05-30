package main

import (
	"arity-cn/asr-web-api-go-demo/util"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	url = "https://k8s.arity.cn/asr/http/asr/toText"
	// 业务方唯一标识id，最高128位，建议不要重复，这里只是模拟
	btId            = "123"
	accessKey       = "accessKey(请替换为正确的accessKey)"
	accessKeySecret = "accessKey(请替换为正确的accessKeySecret)"
	appCode         = "appCode(请替换为正确的appCode)"
	channelCode     = "channelCode(请替换为正确的channelCode)"
	audioFilePath   = "audio/ARITY2023S001W0001.wav"
)

// ResponseData http 返回实体
type ResponseData struct {
	Success bool `json:"success"`
	Data    struct {
		AudioText string `json:"audioText"`
	} `json:"data"`
}

// 读取文件并转为 base64
func fileToBase64(filename string) (string, error) {
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(fileContent), nil
}

// http application/json webapi 示例代码
func main() {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// 生成签名
	signature, timestamp := util.GenerateSignature(accessKey, accessKeySecret, appCode)

	audioContent, err := fileToBase64(audioFilePath)
	if err != nil {
		fmt.Printf("Failed to read audio file: %v\n", err)
		return
	}

	data := map[string]interface{}{
		"btId":        btId,
		"accessKey":   accessKey,
		"appCode":     appCode,
		"channelCode": channelCode,
		"contentType": "RAW",
		"formatInfo":  "WAV",
		"content":     audioContent,
		"timestamp":   timestamp,
		"sign":        signature,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Failed to marshal JSON data: %v\n", err)
		return
	}

	requestBody := bytes.NewBuffer(jsonData)

	client := http.Client{Timeout: 10 * time.Second}
	response, err := client.Post(url, headers["Content-Type"], requestBody)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v\n", err)
		return
	}

	if response.StatusCode == http.StatusOK {
		responseData := ResponseData{}
		err = json.Unmarshal(responseBody, &responseData)
		if err != nil {
			fmt.Printf("Failed to unmarshal response data: %v\n", err)
			return
		}

		if responseData.Success {
			fmt.Printf("语音识别结果: %s\n", responseData.Data.AudioText)
		} else {
			fmt.Printf("请求异常: %s\n", string(responseBody))
		}
	} else {
		fmt.Printf("请求异常, httpCode: %d\n", response.StatusCode)
	}
}
