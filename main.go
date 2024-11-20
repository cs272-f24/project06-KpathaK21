package main

import (
    "bufio" 
    "fmt" 
    "log" 
    "os" 
)

// Declare a global variable for the chatbot instance.
var chatbot *ChatBot

func main() {
    // Retrieve the OpenAI API key from an environment variable.
    apiKey := os.Getenv("OPENAI_PROJECT_KEY")
    if apiKey == "" {
        log.Fatal("API key is missing. Please set OPENAI_PROJECT_KEY environment variable.")
    }

    // Path to the CSV file containing course information.
    csvFilePath := "Fall 2024 Class Schedule 08082024.csv"

    // Initialize the LLM client using the API key.
    llmClient := NewLLMClient(apiKey)

    // Initialize the metadata extractor using the CSV file and the LLM client.
    metadataExtractor, err := NewMetadataExtractor(csvFilePath, llmClient)
    if err != nil {
        log.Fatalf("Failed to initialize MetadataExtractor: %v", err) // Log the error and exit.
    }

    // Add course and instructor data to ChromaDB collections.
    chromaCtx, chromaClient, courseCollection, instructorCollection := Add(metadataExtractor.courses)

    // Initialize the chatbot with the required components: LLM client,  Metadata extractor ,ChromaDB context and 
    // collections for courses and instructors.
    chatbot = NewChatBot(llmClient, metadataExtractor, chromaCtx, chromaClient, courseCollection, instructorCollection)

    // Notify the user that data has been added to collections and start the chatbot.
    fmt.Println("Courses and instructors added to collections.")
    fmt.Println("Entering interactive mode. Type your questions below:")
    runInteractiveMode() // Start interactive user input handling.
}

// runInteractiveMode starts an interactive loop to process user queries.
func runInteractiveMode() {
    // Initialize a scanner to read input from the standard input (terminal).
    scanner := bufio.NewScanner(os.Stdin)

    // Print the initial prompt for the user.
    fmt.Print("\nCatalog search> ")

    // Loop to continuously read and process user input.
    for scanner.Scan() {
        // Read the user's query from input.
        question := scanner.Text()
        if question == "" {
            // If the input is empty, prompt the user to enter a valid query.
            fmt.Println("Please enter a valid query.")
            fmt.Print("\nCatalog search> ")
            continue
        }

        // Use the chatbot to process the user's question.
        answer, err := chatbot.AnswerQuestion(question)
        if err != nil {
            // Handle errors during question processing.
            fmt.Printf("Error processing your question: %v\n", err)
            fmt.Print("\nCatalog search> ")
            continue
        }

        // Print the chatbot's response to the user's question.
        fmt.Println("ChatBot:", answer)
        fmt.Print("\nCatalog search> ") // Prompt the user for the next query.
    }

    // Handle any errors encountered while reading input.
    if err := scanner.Err(); err != nil {
        log.Println("Error reading input:", err)
    }
}
