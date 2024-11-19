package main

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// MakeTool defines the tool for querying courses by various parameters using the jsonschema package.
func MakeTool() openai.FunctionDefinition {
	// Define the schema programmatically using jsonschema package.
	schema := jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"instructor": {
				Type:        jsonschema.String,
				Description: "The canonical name or alias of the instructor (e.g., Philip Peterson).",
			},
			"subject": {
				Type:        jsonschema.String,
				Description: "The subject or department code (e.g., CS, PHIL, BIO).",
			},
			"course": {
				Type:        jsonschema.String,
				Description: "The title of the course (e.g., Intro to Philosophy).",
			},
			"title": {
				Type:        jsonschema.String,
				Description: "Keywords from the course title or short description (e.g., Guitar and Bass Lessons).",
			},
		},
		Required: []string{}, // No required fields, all are optional.
	}

	// Return the function definition with the schema.
	return openai.FunctionDefinition{
		Name:        "query_courses", // Function name.
		Description: "Fetch courses based on instructor, subject, course title, or keywords.",
		Parameters:  schema,          // Pass the schema object directly.
	}
}
