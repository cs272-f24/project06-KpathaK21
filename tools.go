package main

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// WebSearchTool defines a tool for performing web searches
func WebSearchTool() openai.FunctionDefinition {
	schema := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"query": {
				Type:        jsonschema.String,
				Description: "The search query to find relevant web pages (e.g., 'latest AI news').",
			},
		},
		Required: []string{"query"}, // Query is required for a search.
	}

	return openai.FunctionDefinition{
		Name:        "web_search",
		Description: "Perform a web search and return links to relevant pages.",
		Parameters:  schema,
	}
}
