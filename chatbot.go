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

// ChatBot struct combines LLM client, metadata extraction, and ChromaDB collections
// to answer user queries about courses.
type ChatBot struct {
    llmClient            *LLMClient          
    metadata             *MetadataExtractor   
    chromaCtx            context.Context     
    chromaClient         *chroma.Client       
    courseCollection     *chroma.Collection  
    instructorCollection *chroma.Collection  
    context              []openai.ChatCompletionMessage 
}

// NewChatBot initializes and returns a new ChatBot instance.
func NewChatBot(llmClient *LLMClient, metadata *MetadataExtractor, chromaCtx context.Context, chromaClient *chroma.Client, courseCollection, instructorCollection *chroma.Collection) *ChatBot {
    return &ChatBot{
        llmClient:         llmClient,
        metadata:          metadata,
        chromaCtx:         chromaCtx,
        chromaClient:      chromaClient,
        courseCollection:  courseCollection,
        instructorCollection: instructorCollection,
        context: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleSystem, // System message for LLM initialization
                Content: "You are a course assistant. Your role is to help users find course information. " +
                          "If the user asks about courses, invoke the function 'query_courses' with the following fields: " +
                          "'instructor', 'subject', 'title', or 'course'. Infer these fields from the user's query.",
            },
        },
    }
}

// QueryCourses searches the ChromaDB collection for courses based on the query type and term.
func (bot *ChatBot) QueryCourses(term, queryType string) string {
    instructors := InitializeInstructors() // Load instructor aliases
    var queryTerms []string

    // Determine the type of query and prepare search terms
    switch queryType {
    case "instructor":
        // Convert instructor name to canonical form
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

    // Query ChromaDB for matching courses
    queryResults, err := bot.courseCollection.Query(bot.chromaCtx, queryTerms, 10, nil, nil, nil)
    if err != nil {
        log.Printf("Error querying collection: %v", err)
        return "An error occurred while searching for courses."
    }

    // Handle case when no results are found
    if len(queryResults.Documents) == 0 {
        return fmt.Sprintf("No courses found for '%s'.", term)
    }

    // Format results for output
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

    // Append each document to the results
    for _, doc := range queryResults.Documents {
        result.WriteString(fmt.Sprintf("- %s\n", strings.Join(doc, " "))) // Join document fields for readability
    }

    return result.String()
}

// AnswerQuestion processes user questions using LLM and invokes relevant tools for queries.
func (bot *ChatBot) AnswerQuestion(question string) (string, error) {
    queryTool := MakeTool() // Define the query tool for LLM function invocation

    // Add user question to the conversation context
    bot.context = append(bot.context, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleUser,
        Content: question,
    })

    // Generate a response using OpenAI's ChatCompletion API
    resp, err := bot.llmClient.client.CreateChatCompletion(context.Background(),
        openai.ChatCompletionRequest{
            Model:     openai.GPT4oMini, 
            Messages:  bot.context,
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

    // Check if the LLM invoked a tool
    if resp.Choices[0].Message.FunctionCall != nil {
        functionCall := resp.Choices[0].Message.FunctionCall

        // Parse the tool arguments from JSON
        if err := json.Unmarshal([]byte(functionCall.Arguments), &args); err != nil {
            return "", fmt.Errorf("Failed to parse function arguments: %w", err)
        }

        fmt.Printf("Parsed arguments: Instructor: '%s', Subject: '%s', Course: '%s', Title: '%s'\n",
            args.Instructor, args.Subject, args.Course, args.Title)

        // Determine the appropriate query type based on parsed arguments
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

        // Add the tool response to the conversation context
        bot.context = append(bot.context, openai.ChatCompletionMessage{
            Role:    openai.ChatMessageRoleFunction,
            Name:    "query_courses",
            Content: toolResponse,
        })

        // Generate a final LLM response
        resp, err = bot.llmClient.client.CreateChatCompletion(context.Background(),
            openai.ChatCompletionRequest{
                Model:    openai.GPT4oMini, 
                Messages: bot.context,
            },
        )
        if err != nil {
            return "", fmt.Errorf("ChatCompletion failed: %w", err)
        }

        // Extract and return the assistant's final response
        assistantResponse := resp.Choices[0].Message.Content
        bot.context = append(bot.context, openai.ChatCompletionMessage{
            Role:    openai.ChatMessageRoleAssistant,
            Content: assistantResponse,
        })

        return assistantResponse, nil
    }

    // If no tool was invoked, return the LLM's direct response
    assistantResponse := resp.Choices[0].Message.Content
    bot.context = append(bot.context, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleAssistant,
        Content: assistantResponse,
    })

    return assistantResponse, nil
}
