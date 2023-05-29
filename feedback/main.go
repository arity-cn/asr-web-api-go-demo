package main

import (
	"arity-cn/asr-web-api-go-demo/util"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	url = "https://k8s.arity.cn/asr/http/asr/V1/feedback"
	// 对应语音识别时传的的btId
	btId = "btId(请替换为正确的btId)"
	// 对应语音识别返回的requestId
	requestId = "requestId(请替换为正确的requestId)"
	// 是否识别准确 0: 准确 1: 不准确
	exactType       = 1
	accessKey       = "accessKey(请替换为正确的accessKey)"
	accessKeySecret = "accessKey(请替换为正确的accessKeySecret)"
	appCode         = "appCode(请替换为正确的appCode)"
	channelCode     = "channelCode(请替换为正确的channelCode)"
)

// 语音识别反馈 webapi 示例代码
func main() {
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")

	signature, timestamp := util.GenerateSignature(accessKey, accessKeySecret, appCode)
	data := map[string]interface{}{
		"btId":        btId,
		"exactType":   exactType,
		"requestId":   requestId,
		"accessKey":   accessKey,
		"appCode":     appCode,
		"channelCode": channelCode,
		"timestamp":   timestamp,
		"sign":        signature,
	}

	jsonData, _ := json.Marshal(data)
	body := strings.NewReader(string(jsonData))

	httpClient := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header = headers

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var responseData map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseData)
		if err != nil {
			fmt.Println("Error decoding response body:", err)
			return
		}
		if responseData["success"].(bool) {
			fmt.Println("反馈成功")
		} else {
			fmt.Println("反馈异常: ", responseData)
		}
	} else {
		fmt.Println("请求异常, httpCode:", resp.StatusCode)
	}
}
