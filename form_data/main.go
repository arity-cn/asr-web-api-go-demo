package main

import (
	"arity-cn/asr-web-api-go-demo/util"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

const (
	url = "https://k8s.arity.cn/asr/http/asr/toTextBinary"
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

// http form-data webapi 示例代码
func main() {

	// 生成签名
	signature, timestamp := util.GenerateSignature(accessKey, accessKeySecret, appCode)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	f, err := os.Open(audioFilePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	part, err := writer.CreateFormFile("file", "ARITY2023S001W0001.wav")
	if err != nil {
		fmt.Println(err)
		return
	}
	io.Copy(part, f)

	writer.WriteField("btId", btId)
	writer.WriteField("accessKey", accessKey)
	writer.WriteField("appCode", appCode)
	writer.WriteField("channelCode", channelCode)
	writer.WriteField("timestamp", strconv.Itoa(int(timestamp)))
	writer.WriteField("sign", signature)
	writer.WriteField("sampleRateEnum", "SAMPLE_RATE_16K")

	contentType := writer.FormDataContentType()
	writer.Close()

	cli := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", contentType)
	response, err := cli.Do(req)
	if err != nil {
		fmt.Println(err)
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
