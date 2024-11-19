package main

import (
    "context"
    "fmt"
    "strings"
	"log"
	"encoding/json"
	
    chroma "github.com/amikos-tech/chroma-go"
    openai "github.com/sashabaranov/go-openai"
)

// ChatBot uses LLMClient and MetadataExtractor to answer questions
type ChatBot struct {
    llmClient            *LLMClient
    metadata             *MetadataExtractor
    chromaCtx            context.Context
    chromaClient         *chroma.Client
    courseCollection     *chroma.Collection
    instructorCollection *chroma.Collection
}


// NewChatBot initializes a ChatBot with an LLM client, metadata extractor, and ChromaDB context
func NewChatBot(llmClient *LLMClient, metadata *MetadataExtractor, chromaCtx context.Context, chromaClient *chroma.Client, courseCollection, instructorCollection *chroma.Collection) *ChatBot {
    return &ChatBot{
        llmClient:         llmClient,
        metadata:          metadata,
        chromaCtx:         chromaCtx,
        chromaClient:      chromaClient,
        courseCollection:  courseCollection,
        instructorCollection: instructorCollection,
    }
}

func (bot *ChatBot) QueryCourses(term, queryType string) string {
    instructors := InitializeInstructors()
    var queryTerms []string

    // Handle query type
    switch queryType {
    case "instructor":
        // Resolve the canonical name for the instructor
        term = findCanonicalName(term, instructors)
        if term == "" {
            return fmt.Sprintf("No valid instructor found for '%s'.", term)
        }
        fmt.Printf("Canonical name to search for instructor: '%s'\n", term)
        queryTerms = append(queryTerms, term)

    case "subject":
        fmt.Printf("Searching for courses in subject: '%s'\n", term)
        queryTerms = append(queryTerms, term)

    case "title":
        fmt.Printf("Searching for courses with title keywords: '%s'\n", term)
        queryTerms = append(queryTerms, term)

    case "combined":
        fmt.Printf("Searching for combined criteria: '%s'\n", term)
        queryTerms = append(queryTerms, term)

    default:
        return "Invalid query type. Please specify 'instructor', 'subject', 'title', or 'combined'."
    }

    // Query the collection
    queryResults, err := bot.courseCollection.Query(bot.chromaCtx, queryTerms, 10, nil, nil, nil)
    if err != nil {
        log.Printf("Error querying collection: %v", err)
        return "An error occurred while searching for courses."
    }

    // Check if results are empty
    if len(queryResults.Documents) == 0 {
        return fmt.Sprintf("No courses found for '%s'.", term)
    }

    // Format the results
    var result strings.Builder
    switch queryType {
    case "instructor":
        result.WriteString(fmt.Sprintf("Here are the courses taught by '%s':\n", term))
    case "subject":
        result.WriteString(fmt.Sprintf("Here are the courses in the subject '%s':\n", term))
    case "title":
        result.WriteString(fmt.Sprintf("Here are the courses with titles containing '%s':\n", term))
    case "combined":
        result.WriteString(fmt.Sprintf("Here are the courses matching the criteria '%s':\n", term))
    }

    // Iterate through documents and append to results
    for _, doc := range queryResults.Documents {
        // Convert the document array to a single string for readability
        result.WriteString(fmt.Sprintf("- %s\n", strings.Join(doc, " ")))
    }

    return result.String()
}

func (bot *ChatBot) AnswerQuestion(question string) (string, error) {
    queryTool := MakeTool()

    // Updated system prompt to encourage correct argument inference
    dialogue := []openai.ChatCompletionMessage{
        {
            Role: openai.ChatMessageRoleSystem,
            Content: "You are a course assistant. Your role is to help users find course information. " +
                "If the user asks about courses, invoke the function 'query_courses' with the following fields: " +
                "'instructor', 'subject', 'title', or 'course'. Infer these fields from the user's query.",
        },
        {
            Role: openai.ChatMessageRoleUser,
            Content: question,
        },
    }

    resp, err := bot.llmClient.client.CreateChatCompletion(context.Background(),
        openai.ChatCompletionRequest{
            Model:     openai.GPT4oMini,
            Messages:  dialogue,
            Functions: []openai.FunctionDefinition{queryTool},
        },
    )
    if err != nil {
        return "", fmt.Errorf("ChatCompletion failed: %w", err)
    }

    var args struct {
        Instructor string `json:"instructor"`
        Subject    string `json:"subject"`
        Course     string `json:"course"`
        Title      string `json:"title"`
    }

    // Handle tool invocation
    if resp.Choices[0].Message.FunctionCall != nil {
        functionCall := resp.Choices[0].Message.FunctionCall

        // Parse arguments passed to the tool
        if err := json.Unmarshal([]byte(functionCall.Arguments), &args); err != nil {
            return "", fmt.Errorf("Failed to parse function arguments: %w", err)
        }

        fmt.Printf("Parsed arguments: Instructor: '%s', Subject: '%s', Course: '%s', Title: '%s'\n",
            args.Instructor, args.Subject, args.Course, args.Title)

        // Call QueryCourses based on parsed arguments
        var toolResponse string
        if args.Instructor != "" && args.Subject != "" {
            combinedTerm := fmt.Sprintf("%s %s", args.Instructor, args.Subject)
            toolResponse = bot.QueryCourses(combinedTerm, "combined")
        } else if args.Instructor != "" {
            toolResponse = bot.QueryCourses(args.Instructor, "instructor")
        } else if args.Subject != "" {
            toolResponse = bot.QueryCourses(args.Subject, "subject")
        } else if args.Title != "" {
            toolResponse = bot.QueryCourses(args.Title, "title")
        } else if args.Course != "" {
            toolResponse = bot.QueryCourses(args.Course, "course")
        } else {
            toolResponse = "No valid query parameters provided."
        }

        // Append the tool response to the dialogue
        dialogue = append(dialogue, openai.ChatCompletionMessage{
            Role:    openai.ChatMessageRoleFunction,
            Name:    "query_courses",
            Content: toolResponse,
        })

        // Generate a final response from the LLM based on the tool's output
        resp, err = bot.llmClient.client.CreateChatCompletion(context.Background(),
            openai.ChatCompletionRequest{
                Model:    openai.GPT4oMini,
                Messages: dialogue,
            },
        )
        if err != nil {
            return "", fmt.Errorf("ChatCompletion failed: %w", err)
        }
        return resp.Choices[0].Message.Content, nil
    }

    return "No relevant tool was invoked for the question.", nil
}
