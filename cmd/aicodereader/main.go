package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/sashabaranov/go-openai"
)

// flags for cli
var (
	filename = flag.String("f", "", "path to the file to read")
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
	fmt.Println("----- 推理过程  -----")
	fmt.Println(resp.Choices[0].Message.ReasoningContent)

	fmt.Println("----- 最终回答 -----")
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
			Temperature: 0.7,
			Stream:      true,
		},
	)
	if err != nil {
		log.Printf("stream chat error: %v\n", err)
		return
	}
	defer stream.Close()

	isThinking := false

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
			if recv.Choices[0].Delta.ToolCalls != nil ||
				(recv.Choices[0].Delta.Role == "assistant" && !isThinking) {
				if !isThinking {
					fmt.Println("----- 模型思考过程 -----")
					isThinking = true
				}

				if recv.Choices[0].Delta.ToolCalls != nil {
					for _, toolCall := range recv.Choices[0].Delta.ToolCalls {
						if toolCall.Function.Arguments != "" {
							fmt.Print(toolCall.Function.Arguments)
						}
					}
				}
			} else if recv.Choices[0].Delta.Content != "" {
				if isThinking {
					log.Println("----- 模型最终回答 -----")
					isThinking = false
				}

				fmt.Print(recv.Choices[0].Delta.Content)
			}
		}
	}
}

func main() {
	flag.Parse()

	if *filename == "" {
		fmt.Println("filename is required")
		flag.Usage()
		return
	}

	content, err := os.ReadFile(*filename)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
		return
	}

	fmt.Println(string(content))

	config := LoadConfig()

	test_standard_request(config)
	// test_stream_request(config)
}
