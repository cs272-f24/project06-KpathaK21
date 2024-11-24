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
    context              []openai.ChatCompletionMessage
}


// NewChatBot initializes a ChatBot with an LLM client, metadata extractor, and ChromaDB context
func NewChatBot(llmClient *LLMClient, metadata *MetadataExtractor, chromaCtx context.Context, chromaClient *chroma.Client, courseCollection, instructorCollection *chroma.Collection) *ChatBot {
    return &ChatBot{
        llmClient:            llmClient,
        metadata:             metadata,
        chromaCtx:            chromaCtx,
        chromaClient:         chromaClient,
        courseCollection:     courseCollection,
        instructorCollection: instructorCollection,
        context: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleSystem,
                Content: "You are a course assistant. Help users find course information.",
            },
        },
    }
}


func (bot *ChatBot) QueryCourses(term string) string {
    // Find the canonical name for the given term
    instructors := InitializeInstructors()
    canonicalName := findCanonicalName(term, instructors)
    fmt.Printf("Canonical name to search for: %s\n", canonicalName)

    // If the canonical name is empty, return a fallback message
    if canonicalName == "" {
        return fmt.Sprintf("No valid instructor found for '%s'.", term)
    }
	
    // Query the collection using the canonical name
    queryResults, err := bot.courseCollection.Query(bot.chromaCtx, []string{canonicalName}, 5, nil, nil, nil)
    if err != nil {
        log.Printf("Error querying collection: %v", err)
        return "An error occurred while searching for courses."
    }

    // Check if results are empty
    if len(queryResults.Documents) == 0 {
        return fmt.Sprintf("No courses found for %s.", canonicalName)
    }

    // Format the results
    var result strings.Builder
    result.WriteString(fmt.Sprintf("Here are the courses taught by %s:\n", canonicalName))
    for _, doc := range queryResults.Documents {
        result.WriteString(fmt.Sprintf("- %s\n", doc))
    }

    return result.String()
}

func (bot *ChatBot) AnswerQuestion(question string) (string, error) {
    

    // Check if the question is a web search query
    if strings.Contains(strings.ToLower(question), "take me to the web page") {
        // Extract the search query
        query := strings.Replace(strings.ToLower(question), "take me to the web page where i can", "", 1)
        query = strings.TrimSpace(query)

        // Perform web search using the LLM
        webSearchPrompt := fmt.Sprintf("Search the web and provide a list of links for the query: '%s'", query)
        systemMessage := "You are a web search assistant. Generate a list of clickable links for the query provided."

        webSearchResponse, err := bot.llmClient.ChatCompletion(webSearchPrompt, systemMessage)
        if err != nil {
            return "", fmt.Errorf("web search failed: %w", err)
        }

        return webSearchResponse, nil
    }

    // Add the user's question to the context for course-related queries
    bot.context = append(bot.context, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleUser,
        Content: question,
    })

    // Handle course-related queries as usual
    instructors := InitializeInstructors()
    for _, instructor := range instructors {
        for _, alias := range instructor.Aliases {
            if strings.Contains(strings.ToLower(question), strings.ToLower(alias)) {
                question = strings.ReplaceAll(question, alias, instructor.CanonicalName)
            }
        }
    }

    var collectionToQuery *chroma.Collection
    if strings.Contains(strings.ToLower(question), "instructor") {
        collectionToQuery = bot.instructorCollection
    } else {
        collectionToQuery = bot.courseCollection
    }

    documents := Query(bot.chromaCtx, bot.chromaClient, collectionToQuery, question)

    var preamble string
    if len(documents) > 0 {
        preamble = "Based on the available information, here are the relevant matches:\n\n"
        for _, doc := range documents {
            preamble += fmt.Sprintf("- %s\n", strings.Join(doc, " "))
        }
        preamble += "\nPlease use this information to answer the user's question."
    } else {
        preamble = "Provide accurate information based on the context of university courses and instructors."
    }

    bot.context = append(bot.context, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleAssistant,
        Content: preamble,
    })

    req := openai.ChatCompletionRequest{
        Model:    openai.GPT4oMini,
        Messages: bot.context,
    }

    response, err := bot.llmClient.client.CreateChatCompletion(context.Background(), req)
    if err != nil {
        return "", fmt.Errorf("ChatCompletion failed: %w", err)
    }

    if len(response.Choices) > 0 {
        reply := response.Choices[0].Message.Content
        bot.context = append(bot.context, openai.ChatCompletionMessage{
            Role:    openai.ChatMessageRoleAssistant,
            Content: reply,
        })
        return reply, nil
    }

    return "", fmt.Errorf("no response from LLM")
}
