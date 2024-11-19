package main

import (
	"os"
	"testing"
	"strings"
)

// Test Course data to be used in tests
var testCourses = []Course{
	{
		Subject:              "AAS",
		CourseNumber:         "100",
		Section:              "01",
		CRN:                  "42180",
		ScheduleTypeCode:     "SEM",
		CampusCode:           "M",
		Title:                "Black Activists & Visionaries",
		InstructionModeDesc:  "In-Person",       
		MeetingTypeCodes:     "IP",
		MeetDays:             "MW",              
		BeginTime:            "1645",
		EndTime:              "1825",
		MeetStart:            "8/20/24",         
		MeetEnd:              "12/4/24",          
		Building:             "LM",
		Room:                 "140",
		ActualEnrollment:      "30",
		InstructorFirstName:  "Sheryl",
		InstructorLastName:   "Davis",
		InstructorEmail:      "sedavis2@usfca.edu",
		College:              "LA",
	},
	{
		Subject:              "BAT",
		CourseNumber:         "101",
		Section:              "01",
		CRN:                  "99999",
		ScheduleTypeCode:     "LEC",
		CampusCode:           "G",
		Title:                "The Dark Knight's Tactics",
		InstructionModeDesc:  "In-Person",        
		MeetingTypeCodes:     "LEC",
		MeetDays:             "TR",               
		BeginTime:            "1800",
		EndTime:              "2000",
		MeetStart:            "8/20/24",          
		MeetEnd:              "12/4/24",          
		Building:             "Wayne Tower",
		Room:                 "Gotham",
		ActualEnrollment:     "20",
		InstructorFirstName:  "Bruce",
		InstructorLastName:   "Wayne",
		InstructorEmail:      "bwayne@gotham.edu",
		College:              "Justice",
	},
}


// TestAdd verifies that courses are added to the ChromaDB collection without errors
func TestAdd(t *testing.T) {
    // Check if the required environment variable is set
    if os.Getenv("OPENAI_PROJECT_KEY") == "" {
        t.Fatalf("Environment variable OPENAI_PROJECT_KEY is not set")
    }

    // Call Add function with test courses to add them to the collection
    ctx, client, courseCollection, instructorCollection := Add(testCourses)

    // Verify that none of the returned values are nil
    if ctx == nil || client == nil || courseCollection == nil || instructorCollection == nil {
        t.Fatalf("Add function returned nil for context, client, or collections")
    }

    t.Log("Courses and instructors successfully added to collections") // Log successful addition
}


// TestQuery verifies that querying by course title returns expected results
func TestQuery(t *testing.T) {
    // Ensure the environment variable for API key is set
    if os.Getenv("OPENAI_PROJECT_KEY") == "" {
        t.Fatalf("Environment variable OPENAI_PROJECT_KEY is not set")
    }

    // Initialize client and collections with test data
    ctx, client, courseCollection, _ := Add(testCourses) // Ignore instructorCollection for this test

    // Define the course title to search for
    queryTitle := "Black Activists & Visionaries"
    results := Query(ctx, client, courseCollection, queryTitle)

    // Check if the results contain the expected course title
    found := false
    for _, doc := range results {
        for _, field := range doc {
            if strings.Contains(field, queryTitle) {
                found = true
                break
            }
        }
        if found {
            break
        }
    }

    if !found {
        t.Errorf("Expected query result to contain title '%s', but it was not found", queryTitle)
    } else {
        t.Logf("Query returned result(s) for title '%s'", queryTitle) // Log success
    }
}
