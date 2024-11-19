package main

import (
    "io"
    "github.com/gocarina/gocsv"
    "log"
    "encoding/csv"
)


// Course represents a record in the CSV file with csv tags for each field
type Course struct {
    Subject                   string `csv:"SUBJ"`
    CourseNumber              string `csv:"CRSE NUM"`
    Section                   string `csv:"SEC"`
    CRN                       string `csv:"CRN"`
    ScheduleTypeCode          string `csv:"Schedule Type Code"`
    CampusCode                string `csv:"Campus Code"`
    Title                     string `csv:"Title Short Desc"`
    InstructionModeDesc       string `csv:"Instruction Mode Desc"`
    MeetingTypeCodes          string `csv:"Meeting Type Codes"`
    MeetDays                  string `csv:"Meet Days"`
    BeginTime                 string `csv:"Begin Time"`
    EndTime                   string `csv:"End Time"`
    MeetStart                 string `csv:"Meet Start"`
    MeetEnd                   string `csv:"Meet End"`
    Building                  string `csv:"BLDG"`
    Room                      string `csv:"RM"`
    ActualEnrollment          string `csv:"Actual Enrollment"`
    InstructorFirstName       string `csv:"Primary Instructor First Name"`
    InstructorLastName        string `csv:"Primary Instructor Last Name"`
    InstructorEmail           string `csv:"Primary Instructor Email"`
    College                   string `csv:"College"`
}

func ReadCSV(file io.Reader) ([]Course, error) {
    // Set up gocsv to use a custom CSV reader with LazyQuotes enabled
    gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
        reader := csv.NewReader(in)
        reader.Comma = '\t'          // Set to tab-delimited
        reader.LazyQuotes = true     // Enable lazy quotes to handle unescaped quotes
        reader.FieldsPerRecord = -1  // Allow variable field counts if needed
        return reader
    })

    var courses []Course
    if err := gocsv.Unmarshal(file, &courses); err != nil {
        log.Printf("Failed to unmarshal CSV file: %v", err)
        return nil, err
    }

    log.Printf("Successfully read %d records from CSV.", len(courses))
    return courses, nil
}
