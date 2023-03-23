package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type OpenaiConfig struct {
	Url     string
	ApiKey  string
	UseMode string
}

type Role struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Roles []Role

type oaRes struct {
	Id      string         `json:"id"`
	Object  string         `json:"object"`
	Created int32          `json:"created"`
	Usage   oaResUsage     `json:"usage"`
	Choices []oaResChoices `json:"choices"`
}

type oaResUsage struct {
	PromptTokens     int32 `json:"prompt_tokens"`
	CompletionTokens int32 `json:"completion_tokens"`
	TotalTokens      int32 `json:"total_tokens"`
}

type oaResChoices struct {
	Message      Role   `json:"message"`
	FinishReason string `json:"finish_reason"`
	Index        int32  `json:"index"`
}

func getConfig() OpenaiConfig {
	oaConfig := OpenaiConfig{}
	oaConfig.Url = "https://api.openai.com/v1/chat/completions"
	oaConfig.ApiKey = os.Getenv("OPENAI_API")
	oaConfig.UseMode = "gpt-3.5-turbo"
	return oaConfig
}

func setSystemRoleConfig() Role {
	oaRole := Role{}
	oaRole.Role = "system"
	oaRole.Content =
		`
***ここに設定を記述***

上記の設定を参考に、性格や口調や言葉の作り方を模倣してください。
`
	return oaRole
}

func setUserRoleConfig(question string) Role {
	oaRole := Role{}
	oaRole.Role = "user"
	oaRole.Content = question
	return oaRole
}

func setUAssistantRoleConfig(answer string) Role {
	oaRole := Role{}
	oaRole.Role = "assistant"
	oaRole.Content = answer
	return oaRole
}

func openAiChatPost(oaConfig *OpenaiConfig, msgChain *Roles) (string, int32, error) {
	jsonMsg, err := json.Marshal(msgChain)
	if err != nil {
		log.Fatal(err)
	}

	requestBody := `{
		"model": "%s",
		"messages": %s
    }`
	requestBody = fmt.Sprintf(requestBody, oaConfig.UseMode, string(jsonMsg))
	req, err := http.NewRequest("POST", oaConfig.Url, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return "", -1, err
	}
	req.Header.Set("Authorization", "Bearer "+oaConfig.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", -1, err
	}
	defer resp.Body.Close()

	var respBody bytes.Buffer
	respBody.ReadFrom(resp.Body)
	gptRes := oaRes{}
	if err := json.Unmarshal(respBody.Bytes(), &gptRes); err != nil {
		return "", -1, err
	}

	resMsg := ""
	if len(gptRes.Choices) > 0 {
		resMsg = gptRes.Choices[0].Message.Content
	}
	tokens := gptRes.Usage.TotalTokens

	return resMsg, tokens, nil
}

func main() {
	oaConfig := getConfig()

	sys := setSystemRoleConfig()
	usr := setUserRoleConfig("準備は良いですか？")
	msgChain := Roles{}
	msgChain = append(msgChain, sys, usr)

	//最初の会話
	res, tokens, err := openAiChatPost(&oaConfig, &msgChain)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tokens)
	ast := setUAssistantRoleConfig(res)
	msgChain = append(msgChain, ast)

	//会話を続ける
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Question: ")
		question, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		question = strings.TrimSpace(question)

		if question == "exit" {
			break
		}

		usr = setUserRoleConfig(question)
		msgChain = append(msgChain, usr)
		res, tokens, err = openAiChatPost(&oaConfig, &msgChain)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print("Answer: ")
		fmt.Println(res)
		fmt.Printf("total token: %v\n", tokens)
		ast = setUAssistantRoleConfig(res)
		msgChain = append(msgChain, ast)

		//max tokens about 4000
		if tokens > 3500 {
			msgChain = append(msgChain[:1], msgChain[3:]...) //質問・回答の1セットを削除
		}
	}
}
