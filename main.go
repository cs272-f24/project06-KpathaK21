package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"

    chroma "github.com/amikos-tech/chroma-go"
)

var chatbot *ChatBot 

func main() {
    // Retrieve OpenAI API key from environment variable.
    apiKey := os.Getenv("OPENAI_PROJECT_KEY")
    if apiKey == "" {
        log.Fatal("API key is missing. Please set OPENAI_PROJECT_KEY environment variable.")
    }
    csvFilePath := "Fall 2024 Class Schedule 08082024.csv" // Path to the CSV file.

    llmClient := NewLLMClient(apiKey) // Initialize LLM client.

    // Initialize metadata extractor using the CSV file.
    metadataExtractor, err := NewMetadataExtractor(csvFilePath, llmClient)
    if err != nil {
        log.Fatalf("Failed to initialize MetadataExtractor: %v", err)
    }

    // Add course and instructor data to ChromaDB collections.
    chromaCtx, chromaClient, courseCollection, instructorCollection := Add(metadataExtractor.courses)

    // Initialize the chatbot with the LLM client, metadata, and ChromaDB collections.
    chatbot = NewChatBot(llmClient, metadataExtractor, chromaCtx, chromaClient, courseCollection, instructorCollection)

    fmt.Println("Courses and instructors added to collections.")
    fmt.Println("Entering interactive mode. Type your questions below:")
    runInteractiveMode(chromaCtx, chromaClient, courseCollection) // Start interactive mode.
}

// runInteractiveMode reads user queries from stdin and responds with chatbot answers.
func runInteractiveMode(ctx context.Context, client *chroma.Client, collection *chroma.Collection) {
    scanner := bufio.NewScanner(os.Stdin) // Read user input from standard input.
    fmt.Print("\nCatalog search> ") // Prompt for user input.
    for scanner.Scan() {
        question := scanner.Text() // Read the user's query.
        if question == "" {
            fmt.Println("Please enter a valid query.")
            continue
        }

        // Process the user's question using the chatbot.
        answer, err := chatbot.AnswerQuestion(question)
        if err != nil {
            fmt.Printf("Error processing your question: %v\n", err)
            continue
        }

        fmt.Println(answer) // Print the chatbot's answer.
        fmt.Print("\nCatalog search> ") // Prompt for the next query.
    }

    // Handle any errors encountered while reading input.
    if err := scanner.Err(); err != nil {
        log.Println("Error reading input:", err)
    }
}
