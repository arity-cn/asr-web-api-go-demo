package main

import (
	"arity-cn/asr-web-api-go-demo/util"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// 识别结果
var result string

const (
	wsUrl = "wss://k8s.arity.cn/asr/ws"
	// 业务方唯一标识id，最高128位，建议不要重复，这里只是模拟
	btId            = "123"
	accessKey       = "accessKey(请替换为正确的accessKey)"
	accessKeySecret = "accessKey(请替换为正确的accessKeySecret)"
	appCode         = "appCode(请替换为正确的appCode)"
	channelCode     = "channelCode(请替换为正确的channelCode)"
	audioFilePath   = "audio/ARITY2023S001W0001.wav"
)

// websocket webapi 示例代码
func main() {

	completeURL := getWebSocketURL(wsUrl, btId, accessKey, accessKeySecret, appCode, channelCode)
	fmt.Printf("构建参数后的url: %s\n", completeURL)

	dialer := websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
	}
	conn, _, err := dialer.Dial(completeURL, nil)
	if err != nil {
		fmt.Println("Error connecting to websocket:", err)
		return
	}
	defer conn.Close()

	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading websocket message:", err)
			break
		}
		handleWebSocketMessage(conn, messageType, data)
	}
}

// 构建请求参数，获取携带参数的ws url
func getWebSocketURL(websocketURL, btID, accessKey, accessKeySecret, appCode, channelCode string) string {
	signature, timestamp := util.GenerateSignature(accessKey, accessKeySecret, appCode)
	params := url.Values{}
	params.Set("accessKey", accessKey)
	params.Set("appCode", appCode)
	params.Set("channelCode", channelCode)
	params.Set("btId", btID)
	params.Set("sign", signature)
	params.Set("timestamp", strconv.Itoa(int(timestamp)))

	return fmt.Sprintf("%s?%s", websocketURL, params.Encode())
}

// 构建开始报文
func buildStartFrame() string {
	business := map[string]interface{}{
		"vadEos": 5000,
	}
	data := map[string]interface{}{
		"audioFormatInfo": "WAV",
		"sampleRate":      "SAMPLE_RATE_16K",
	}
	frame := map[string]interface{}{
		"signal":   "start",
		"business": business,
		"data":     data,
	}
	frameJSON, _ := json.Marshal(frame)
	return string(frameJSON)
}

// 构建结束报文
func buildEndFrame() string {
	frame := map[string]interface{}{
		"signal": "end",
	}
	frameJSON, _ := json.Marshal(frame)
	return string(frameJSON)
}

// websocket消息处理
func handleWebSocketMessage(conn *websocket.Conn, messageType int, data []byte) {
	fmt.Printf("接收到消息：%s\n", data)
	var messageObj map[string]interface{}
	json.Unmarshal(data, &messageObj)

	// 消息分发
	switch messageObj["type"].(string) {
	case "verify":
		afterProcessVerify(conn, messageObj)
	case "server_ready":
		afterProcessServerReady(conn, messageObj)
	case "partial_result":
		afterProcessPartialResult(conn, messageObj)
	case "final_result":
		afterProcessFinalResult(conn, messageObj)
	case "speech_end":
		afterProcessSpeechEnd(conn, messageObj)
	}
}

// 处理验证结果报文
func afterProcessVerify(conn *websocket.Conn, messageObj map[string]interface{}) {
	if messageObj["status"] == "ok" {
		fmt.Printf("校验通过，requestId: %s, code: %s\n", messageObj["requestId"], messageObj["code"])
		conn.WriteMessage(websocket.TextMessage, []byte(buildStartFrame()))
	} else {
		fmt.Printf("校验失败，code: %s, message: %s\n", messageObj["code"], messageObj["message"])
	}
}

// 处理准备好进行语音识别报文
func afterProcessServerReady(conn *websocket.Conn, messageObj map[string]interface{}) {
	// 处理服务端准备好进行语音识别报文，发送二进制报文，发送完二进制报文后发送结束识别报文
	fmt.Println("处理服务端准备好进行语音识别报文")
	if messageObj["status"] == "ok" {
		file, err := os.Open(audioFilePath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()

		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println("Error reading file:", err)
				return
			}
			// 发送二进制报文
			conn.WriteMessage(websocket.BinaryMessage, buf[:n])
		}
		// 发送结束识别报文
		conn.WriteMessage(websocket.TextMessage, []byte(buildEndFrame()))
	} else {
		fmt.Printf("服务器准备失败, 报文: %v\n", messageObj)
	}
}

// 处理中间识别结果报文
func afterProcessPartialResult(conn *websocket.Conn, messageObj map[string]interface{}) {
	fmt.Println("处理中间结果报文")
	var nbest []map[string]interface{}
	json.Unmarshal([]byte(messageObj["nbest"].(string)), &nbest)
	sentence := nbest[0]["sentence"]
	if len(sentence.(string)) == 0 {
		fmt.Println("没有识别出结果，跳过此次中间结果报文处理")
		return
	}
	if len(result) > 0 {
		fmt.Printf("当前语音识别结果：%s，%s\n", result, sentence)
	} else {
		fmt.Printf("当前语音识别结果：%s\n", sentence)
	}
}

// 处理最终识别结果报文
func afterProcessFinalResult(conn *websocket.Conn, messageObj map[string]interface{}) {
	fmt.Println("处理最终结果报文")
	var nbest []map[string]interface{}
	json.Unmarshal([]byte(messageObj["nbest"].(string)), &nbest)
	sentence := nbest[0]["sentence"]
	if len(sentence.(string)) == 0 {
		fmt.Println("没有识别出结果，跳过此次最终结果报文处理")
		return
	}
	if len(result) > 0 {
		result += "，"
		result += sentence.(string)
		fmt.Printf("当前语音识别结果：%s\n", result)
	} else {
		result += sentence.(string)
		fmt.Printf("当前语音识别结果：%s\n", result)
	}
}

// 处理识别结束报文
func afterProcessSpeechEnd(conn *websocket.Conn, messageObj map[string]interface{}) {
	fmt.Println("收到识别结束报文")
	if len(result) > 0 {
		result += "。"
	}
	fmt.Printf("最终语音识别结果：%s\n", result)
	conn.Close()
}
