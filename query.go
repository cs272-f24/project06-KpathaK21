
package main

import(
	"context"
	"encoding/json"
	"strconv"
	"os"
	"log"
	"fmt"
	"time"
	
	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
)



// Add adds a list of Course objects to the ChromaDB collection
func Add(courses []Course) (context.Context, *chroma.Client, *chroma.Collection, *chroma.Collection) {
    openaikey := os.Getenv("OPENAI_PROJECT_KEY")
    if openaikey == "" {
        log.Fatalf("OPENAI_PROJECT_KEY not set in environment variables")
    }

    ctx := context.TODO()
    client, err := chroma.NewClient("http://localhost:8000")
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    openaiEf, err := openai.NewOpenAIEmbeddingFunction(openaikey)
    if err != nil {
        log.Fatalf("Error creating OpenAI embedding function: %s", err)
    }

    // Get or create the courses collection
    coursesCollection, err := client.GetCollection(ctx, "courses-collection", openaiEf)
    if err != nil {
        log.Fatalf("Failed to get courses collection: %v", err)
    }

    // Get or create the instructors collection
    instructorsCollection, err := client.GetCollection(ctx, "instructors-collection", openaiEf)
    if err != nil {
        log.Fatalf("Failed to get instructors collection: %v", err)
    }

    // Check if there are existing documents in the courses collection
    testQueryResults, err := coursesCollection.Query(ctx, []string{"test"}, 1, nil, nil, nil)
    if err == nil && len(testQueryResults.Documents) > 0 {
        fmt.Println("Courses already loaded in ChromaDB, skipping addition.")
        return ctx, client, coursesCollection, instructorsCollection
    }

    instructors := InitializeInstructors()
    uniqueInstructorNames := make(map[string]struct{}) // Track unique instructors

    fmt.Printf("Adding %d courses to the collection...\n", len(courses))
    for i, course := range courses {
        fullName := course.InstructorFirstName + " " + course.InstructorLastName
        canonicalName := findCanonicalName(fullName, instructors)

        // Track unique instructor canonical names
        uniqueInstructorNames[canonicalName] = struct{}{}

        metadata := map[string]interface{}{
            "instructor_canonical_name": canonicalName,
        }
		
        jsonData, err := json.Marshal(course)
        if err != nil {
            log.Printf("Failed to marshal course to JSON: %v", err)
            continue
        }

        // Unique document ID for the course
        documentID := strconv.Itoa(i)

        // Use retry mechanism to add the course
        fmt.Printf("Processing course %d of %d: %s\n", i+1, len(courses), course.Title)
        addCourseWithRetry(ctx, coursesCollection, []map[string]interface{}{metadata}, []string{string(jsonData)}, []string{documentID})
    }

    fmt.Printf("Adding %d unique instructors to the collection...\n", len(uniqueInstructorNames))
    for name := range uniqueInstructorNames {
        // Unique document ID for the instructor
        instructorID := name

        // Add instructor name to the instructors collection
        addCourseWithRetry(ctx, instructorsCollection, nil, []string{name}, []string{instructorID})
    }

    fmt.Println("Finished adding courses and instructors to the collections.")
    return ctx, client, coursesCollection, instructorsCollection
}

// addCourseWithRetry handles adding a document to the ChromaDB collection with retries.
func addCourseWithRetry(ctx context.Context, collection *chroma.Collection, metadata []map[string]interface{}, documents []string, ids []string) {
    retries := 3 // Maximum number of retries
    var err error

    for i := 0; i < retries; i++ {
        _, err = collection.Add(ctx, nil, metadata, documents, ids)
        if err == nil {
            fmt.Printf("Successfully added document with ID: %s\n", ids[0])
            return
        }
        log.Printf("Retry %d: Failed to add document with ID %s: %v", i+1, ids[0], err)
        time.Sleep(time.Second * time.Duration(i+1)) // Exponential backoff
    }

    // If all retries fail, log the final error
    log.Printf("Failed to add document with ID %s after %d retries: %v", ids[0], retries, err)
}

// Query searches the ChromaDB collection for a term and retrieves matching documents
func Query(ctx context.Context, client *chroma.Client, collection *chroma.Collection, term string) [][]string {
    terms := []string{term}
    fmt.Printf("Querying for term: %s\n", term)

    queryResults, err := collection.Query(ctx, terms, 5, nil, nil, nil)
    if err != nil {
        log.Fatalf("failed to query collection: %v", err)
    }

    documents := queryResults.Documents

    return documents
}


