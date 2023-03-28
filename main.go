package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func openAiChatPost(oaConfig *OpenaiConfig, msgChain *Roles) (string, int32, error) {
	jsonMsg, err := json.Marshal(msgChain)
	if err != nil {
		log.Fatal(err)
	}

	requestBody := `{
		"model": "%s",
		"messages": %s
    }`
	requestBody = fmt.Sprintf(requestBody, oaConfig.ChatUseMode, string(jsonMsg))
	req, err := http.NewRequest("POST", oaConfig.ChatUrl, bytes.NewBuffer([]byte(requestBody)))
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

func openAiWhisperPost(oaConfig *OpenaiConfig, filePath string) (string, error) {
	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	io.Copy(part, file)

	part, err = writer.CreateFormField("model")
	if err != nil {
		return "", err
	}
	modelName := bytes.NewReader([]byte(oaConfig.WhisperUseMode))
	io.Copy(part, modelName)

	//Close multipart writer
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, oaConfig.WhisperUrl, &body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+oaConfig.ApiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var respBody bytes.Buffer
	respBody.ReadFrom(resp.Body)

	gptRes := oaResWhisper{}
	if err := json.Unmarshal(respBody.Bytes(), &gptRes); err != nil {
		return "", err
	}
	t := gptRes.Text
	return t, nil
}

func fileCreate(filePath string) bool {
	_, err := os.Create(filePath)

	return err == nil
}

func fileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func permCheck(filePath string) error {
	fi, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	r, w, x := judge(fi.Mode().Perm())
	fmt.Printf("r=%v\tw=%v\tx=%v\n", r, w, x)
	return nil
}

func judge(perm fs.FileMode) (bool, bool, bool) {
	var (
		r = perm&0400 == 0400
		w = perm&0200 == 0200
		x = perm&0100 == 0100
	)

	return r, w, x
}

func main() {
	oaConfig := getConfig()
	voiceFilePath := "voice.mp3"

	sys := setSystemRoleConfig()
	usr := setUserRoleConfig("準備は良いですか？")
	msgChain := Roles{}
	msgChain = append(msgChain, sys, usr)

	//会話の準備
	res, tokens, err := openAiChatPost(&oaConfig, &msgChain)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tokens)
	ast := setUAssistantRoleConfig(res)
	msgChain = append(msgChain, ast)

	//テキストで会話する場合はコメントを外す
	//reader := bufio.NewReader(os.Stdin)
	for {
		/*
			question, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			question = strings.TrimSpace(question)

			if question == "exit" {
				break
			}
		*/
		if !fileExist(voiceFilePath) {
			continue
		}
		fmt.Print("Question: ")
		_, err = os.ReadFile(voiceFilePath)
		if err != nil {
			log.Fatal(err)
		}
		question, err := openAiWhisperPost(&oaConfig, voiceFilePath)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(question)

		//読みだした音声ファイルは削除する
		err = os.Remove(voiceFilePath)
		if err != nil {
			log.Fatal(err)
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
