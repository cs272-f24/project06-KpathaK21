package main

import (
	"fmt"
	"log"
	"os"
	//"strings"
	"testing"
	
)

func RealChatBot() *ChatBot {
    apiKey := os.Getenv("OPENAI_PROJECT_KEY")
    if apiKey == "" {
        log.Fatal("API key is missing. Please set OPENAI_PROJECT_KEY environment variable.")
    }

    llmClient := NewLLMClient(apiKey)

    // Open and parse the CSV file
    csvFilePath := "Fall 2024 Class Schedule 08082024.csv"
    csvFile, err := os.Open(csvFilePath)
    if err != nil {
        log.Fatalf("Failed to open CSV file: %v", err)
    }
    defer csvFile.Close()

    courses, err := ReadCSV(csvFile)
    if err != nil {
        log.Fatalf("Failed to read CSV file: %v", err)
    }
    fmt.Printf("Loaded %d courses from CSV.\n", len(courses))

    // Initialize MetadataExtractor
    metadataExtractor := &MetadataExtractor{courses: courses}

    // Add courses and instructors to ChromaDB
    chromaCtx, chromaClient, courseCollection, instructorCollection := Add(metadataExtractor.courses)

    // Return the chatbot
    return NewChatBot(llmClient, metadataExtractor, chromaCtx, chromaClient, courseCollection, instructorCollection)
}



func TestCanonicalName(t *testing.T) {
    instructors := InitializeInstructors()
    name := findCanonicalName("Phil Peterson", instructors)
    if name != "Philip Peterson" {
        t.Errorf("Expected 'Philip Peterson', got '%s'", name)
    }
}

func TestContext(t *testing.T) {
	chatbot := RealChatBot()

	// First question: Who is teaching CS 272?
	question1 := "Who is teaching CS 272?"
	answer1, _:= chatbot.AnswerQuestion(question1)
	fmt.Printf("Answer for question '%s':\n%s\n", question1, answer1)

	// Second question: What's his email address?
	question2 := "What's his email address?"
	answer2, _ := chatbot.AnswerQuestion(question2)
	fmt.Printf("Answer for question '%s':\n%s\n", question2, answer2)
}

func TestMultiple(t *testing.T) {
	chatbot := RealChatBot()

	fmt.Printf("What CS courses are Phil Peterson and Greg Benson teaching?\n")
	// Question: What CS courses are Phil Peterson and Greg Benson teaching?
	question1 := "What CS courses is Phil Peterson teaching?"
	answer1, _ := chatbot.AnswerQuestion(question1)

	question2 := "What CS courses is Greg Benson teaching?"
	answer2, _ := chatbot.AnswerQuestion(question2)

	// Print the response
	fmt.Printf("Answer for question '%s':\n%s\n", answer1, answer2)

}

// TestLocation tests queries with location constraints
func TestLocation(t *testing.T) {
	chatbot := RealChatBot()

	question := "What CS course is Phil Peterson teaching in LS G12?"
	answer, _  := chatbot.AnswerQuestion(question)

	fmt.Printf("Answer for question '%s':\n%s\n", question, answer)
}
