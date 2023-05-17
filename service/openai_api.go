package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/errgo.v2/errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type CompletionRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float32 `json:"temperature"`
}

type CompletionResponse struct {
	Choices []struct {
		Text    string  `json:"text"`
		Index   int     `json:"index"`
		LogProb float32 `json:"logprobs"`
		Finish  string  `json:"finish_reason"`
	} `json:"choices"`
}

func OpenAiComplete(userMsg string) (string, error) {
	url := "https://api.openai.com/v1/completions"
	promptTemplate :=
		`分析以下信息并转成JSON
-------
%s
------
所需的keys包括"number", "duration", "start_time"及"people" 
其中，number代表最大参会人数（若无信息则默认为4），duration单位为小时（若无信息则默认为1），start_time参照示例格式，people用拼音（若无指定信息则[]内为空），“今天”是%s

输出示例:
{
    "number": "..",
    "duration": "..",
    "start_time": "2023-01-01 15:00:00",
    "people": ["xumengyuan", "tanchanghao", "lizhongchen"]
}
你的输出：`
	requestBody := CompletionRequest{
		Model:     "text-davinci-003",
		Prompt:    fmt.Sprintf(promptTemplate, userMsg, time.Now()),
		MaxTokens: 2000,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	fmt.Println("send to gpt", string(requestBodyBytes))
	if err != nil {
		return "", errors.Wrap(err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return "", errors.Wrap(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openAiKey))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("receive from gpt", string(bodyBytes))
	var response CompletionResponse
	if err = json.Unmarshal(bodyBytes, &response); err != nil {
		return "", errors.Wrap(err)
	}
	return response.Choices[0].Text, nil
}
