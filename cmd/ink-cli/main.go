package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/samdammers/go-ink/ink"
)

func main() {
	storyPath := flag.String("story", "", "Path to the .ink.json file")
	flag.Parse()

	if *storyPath == "" {
		fmt.Println("Please provide a story file using -story")
		return
	}

	jsonBytes, err := os.ReadFile(*storyPath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	story, err := ink.NewStory(string(jsonBytes))
	if err != nil {
		log.Fatalf("Failed to load story: %v", err)
	}

	fmt.Println("Loaded story successfully.")

	for story.CanContinue() {
		text, err := story.Continue()
		if err != nil {
			log.Fatalf("Runtime error: %v", err)
		}
		fmt.Print(text)
	}

	fmt.Println("\n--- End of Story ---")
}
