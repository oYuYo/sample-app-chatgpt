package main

import "os"

func setSystemRole() string {
	content :=
		`
***ここに設定を記述***
上記の設定を参考に、性格や口調や言葉の作り方を模倣してください。
`
	return content
}

func getConfig() OpenaiConfig {
	oaConfig := OpenaiConfig{}
	oaConfig.ChatUrl = "https://api.openai.com/v1/chat/completions"
	oaConfig.WhisperUrl = "https://api.openai.com/v1/audio/transcriptions"
	oaConfig.ApiKey = os.Getenv("OPENAI_API")
	oaConfig.ChatUseMode = "gpt-3.5-turbo"
	oaConfig.WhisperUseMode = "whisper-1"
	return oaConfig
}

func setSystemRoleConfig() Role {
	oaRole := Role{}
	oaRole.Role = "system"
	oaRole.Content = setSystemRole()
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
