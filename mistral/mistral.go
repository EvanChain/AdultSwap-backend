package mistral

import (
	"awesomeProject2/global"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type MistralRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type MistralResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

func PostMistral(text string) {
	apiKey := global.MistralApiKeyConfig.ApiKey // API Key
	url := global.MistralApiKeyConfig.ApiUrl
	question := "分析以下数据，判断当前最适合质押的价格(请使用中文回复！)\n" + text
	requestBody := MistralRequest{
		Model: "mistral-tiny", // 或 mistral-small, mistral-medium
		Messages: []ChatMessage{
			{Role: "user", Content: question},
		},
	}

	jsonData, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey) // 确保 Bearer 后有空格

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("HTTP请求失败:", err)
		return
	}
	defer resp.Body.Close()

	// 打印API返回的原始状态码和错误信息
	fmt.Println("API状态码:", resp.Status)
	if resp.StatusCode == 401 {
		fmt.Println("错误原因: API Key 无效或未正确设置")
	}

	// 尝试解析响应体
	var response MistralResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		fmt.Println("响应解析失败:", err)
		return
	}

	if len(response.Choices) > 0 {
		fmt.Println("Mistral回复:", response.Choices[0].Message.Content)
	}
}
