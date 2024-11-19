package main

import (
    "fmt"
    "os"
    "bufio"
    "strings"
)

// MetadataExtractor loads course data and extracts metadata like instructors and departments.
type MetadataExtractor struct {
    Instructors []string
    Departments []string
    courses     []Course
    header 		string
}

// Instructor represents an instructor with a canonical name and aliases
type Instructor struct {
    CanonicalName string
    Aliases       []string
}

// NewMetadataExtractor reads course data and the header from the CSV file if needed.
func NewMetadataExtractor(csvFilePath string, client *LLMClient) (*MetadataExtractor, error) {
    // Open the CSV file containing course data.
    file, err := os.Open(csvFilePath)
    if err != nil {
        return nil, fmt.Errorf("Error opening file: %w", err)
    }
    defer file.Close()

    // Read the first line to capture the header row
    scanner := bufio.NewScanner(file)
    if !scanner.Scan() {
        return nil, fmt.Errorf("Error reading header row: %w", scanner.Err())
    }
    header := scanner.Text()

    // Read the rest of the CSV data into course records.
    courses, err := ReadCSV(file)
    if err != nil {
        return nil, fmt.Errorf("Error reading CSV: %w", err)
    }

    instructors := uniqueInstructors(courses)
    departments := uniqueSubjects(courses)

    return &MetadataExtractor{
        Instructors: instructors,
        Departments: departments,
        courses:     courses,
        header:      header,
    }, nil
}

// InitializeInstructors creates a list of instructors with canonical names
func InitializeInstructors() []Instructor {
    return []Instructor{
        {CanonicalName: "Philip Peterson", Aliases: []string{"Phil Peterson", "Philip Peterson"}},
        {CanonicalName: "Philip Choong", Aliases: []string{"Phil Choong", "Philip Choong"}},
    }
}

func findCanonicalName(inputName string, instructors []Instructor) string {
    inputName = strings.TrimSpace(inputName) // Remove leading and trailing whitespace
    if inputName == "" {
        return inputName // Return immediately if the input is empty
    }

    for _, instructor := range instructors {
        for _, alias := range instructor.Aliases {
            if strings.EqualFold(inputName, alias) {
                fmt.Printf("Substituting alias '%s' with canonical name '%s'\n", inputName, instructor.CanonicalName)
                return instructor.CanonicalName
            }
        }
    }
    fmt.Printf("No canonical substitution found for '%s'. Using original name.\n", inputName)
    return inputName // Return the original name if no match is found
}

// uniqueInstructors creates a list of unique instructor canonical names from the courses.
func uniqueInstructors(courses []Course) []string {
    instructors := InitializeInstructors() // Load the list of instructors with canonical names
    instructorSet := make(map[string]bool)
    uniqueInstructors := []string{}

    for _, course := range courses {
        fullName := course.InstructorFirstName + " " + course.InstructorLastName
        canonicalName := findCanonicalName(fullName, instructors)
        if !instructorSet[canonicalName] {
            instructorSet[canonicalName] = true
            uniqueInstructors = append(uniqueInstructors, canonicalName)
        }
    }
    return uniqueInstructors
}

// uniqueSubjects creates a list of unique department/subject names from the courses.
func uniqueSubjects(courses []Course) []string {
    subjectSet := make(map[string]bool)
    subjects := []string{}

    for _, course := range courses {
        if !subjectSet[course.Subject] {
            subjectSet[course.Subject] = true
            subjects = append(subjects, course.Subject)
        }
    }
    return subjects
}
