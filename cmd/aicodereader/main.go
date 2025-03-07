package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

type Config struct {
	APIKey  string
	Model   string
	BaseURL string
	Stream  bool
}

func LoadConfig() Config {
	config := Config{
		APIKey:  os.Getenv("ARK_API_KEY"),
		Model:   os.Getenv("MODEL"),
		BaseURL: os.Getenv("BASE_URL"),
		Stream:  os.Getenv("STREAM") != "",
	}

	return config
}

func test_standard_request(config Config) {
	openaiConfig := openai.DefaultConfig(config.APIKey)
	openaiConfig.BaseURL = config.BaseURL
	model := config.Model

	client := openai.NewClientWithConfig(openaiConfig)
	log.Println("----- standard request -----")
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "你是人工智能助手",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "常见的十字花科植物有哪些？",
				},
			},
		},
	)
	if err != nil {
		log.Printf("ChatCompletion error: %v\n", err)
		return
	}
	fmt.Println(resp.Choices[0].Message.Content)
}

func test_stream_request(config Config) {
	openaiConfig := openai.DefaultConfig(config.APIKey)
	openaiConfig.BaseURL = config.BaseURL
	model := config.Model

	client := openai.NewClientWithConfig(openaiConfig)

	log.Println("----- streaming request -----")
	stream, err := client.CreateChatCompletionStream(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "你是人工智能助手",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "常见的十字花科植物有哪些？",
				},
			},
		},
	)
	if err != nil {
		log.Printf("stream chat error: %v\n", err)
		return
	}
	defer stream.Close()

	for {
		recv, err := stream.Recv()
		if err == io.EOF {
			return
		}

		if err != nil {
			log.Printf("Stream chat error: %v\n", err)
			return
		}

		if len(recv.Choices) > 0 {
			fmt.Print(recv.Choices[0].Delta.Content)
		}
	}
}

func main() {
	config := LoadConfig()

	test_standard_request(config)
	test_stream_request(config)
}
