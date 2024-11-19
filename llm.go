package main

import (
    "context"
    "fmt"

    openai "github.com/sashabaranov/go-openai"
)

// LLMClient wraps the OpenAI client for interaction with an LLM (Large Language Model).
type LLMClient struct {
    client *openai.Client // OpenAI client for API communication.
}

// NewLLMClient initializes a new LLMClient with the provided OpenAI API key.
// Parameters:
// - apiKey: The API key to authenticate with OpenAI's API.
// Returns:
// - A pointer to an LLMClient instance.
func NewLLMClient(apiKey string) *LLMClient {
    client := openai.NewClient(apiKey) // Create a new OpenAI client using the API key.
    return &LLMClient{client: client} // Wrap the OpenAI client in an LLMClient instance.
}

// ChatCompletion sends a user's query to the LLM and retrieves a response.
// Parameters:
// - question: The user's input question or query.
// - systemMessage: A system-level instruction to guide the LLM's behavior.
// Returns:
// - A string containing the LLM's response.
// - An error if the API call or response processing fails.
func (llm *LLMClient) ChatCompletion(question, systemMessage string) (string, error) {
    // Create a chat completion request with the given system message and user query.
    req := openai.ChatCompletionRequest{
        Model: openai.GPT4oMini, // Specify the model to use for the completion.
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleSystem, // System-level instruction to set context.
                Content: systemMessage,
            },
            {
                Role:    openai.ChatMessageRoleUser, // User's query as input for the LLM.
                Content: question,
            },
        },
    }

    // Call the OpenAI API to generate a chat completion.
    resp, err := llm.client.CreateChatCompletion(context.Background(), req)
    if err != nil {
        // Return an error if the API call fails.
        return "", fmt.Errorf("CreateChatCompletion failed: %w", err)
    }

    // Extract and return the content of the LLM's response message.
    return resp.Choices[0].Message.Content, nil
}
