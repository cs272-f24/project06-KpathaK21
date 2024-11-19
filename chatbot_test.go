package main

import (
	"fmt"
	"log"
	"os"
	"strings"
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

func TestPhil(t *testing.T) {
	chatbot := RealChatBot()
	question := "What CS classes is Phil Peterson teaching?"


	// Get the answer for the question
	answer, err := chatbot.AnswerQuestion(question)
	fmt.Printf("Answer for question '%s':\n%s\n", question, answer)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

}


// TestPHIL tests the ChatBot response for philosophy courses
func TestPHIL(t *testing.T) {
	chatbot := RealChatBot()
	question := "Which philosophy courses are offered this semester?"
	answer, err := chatbot.AnswerQuestion(question)
	fmt.Printf("Answer for question '%s':\n%s\n", question, answer)
	if err != nil || !strings.Contains(answer, "Great Philosophical Questions") {
		t.Errorf("Expected answer to mention 'Great Philosophical Questions', got: %v", answer)
	}
}

// TestBio tests the ChatBot response for Bioinformatics location
func TestBio(t *testing.T) {
	chatbot := RealChatBot()
	question := "Where does Bioinformatics meet?"
	answer, err := chatbot.AnswerQuestion(question)
	fmt.Printf("Answer for question '%s':\n%s\n", question, answer)
	if err != nil || !strings.Contains(answer, "KA") || !strings.Contains(answer, "311") {
		t.Errorf("Expected answer to mention 'Building KA, Room 311', got: %v", answer)
	}
}

func TestGuitar(t *testing.T) {
	chatbot := RealChatBot()
	question := "Can I learn guitar this semester?"
	answer, err := chatbot.AnswerQuestion(question)
	fmt.Printf("Answer for question '%s':\n%s\n", question, answer)
	if err != nil || !strings.Contains(answer, "Guitar and Bass Lessons") {
		t.Errorf("Expected answer to mention 'Guitar and Bass Lessons', got: %v", answer)
	}
}

func TestMultiple(t *testing.T) {
	chatbot := RealChatBot()
	question := "I would like to take a Rhetoric course from Phil Choong. What can I take?"
	answer, _ := chatbot.AnswerQuestion(question)
	fmt.Printf("Answer for question '%s':\n%s\n", question, answer)

}
