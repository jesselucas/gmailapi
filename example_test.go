package gmailapi

import (
	"fmt"
	"log"

	"golang.org/x/net/context"
	gmail "google.golang.org/api/gmail/v1"
)

func listLabels() {
	// Create config from a json file
	config, err := ConfigFromJSON("client_secret.json", gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatal(err)
	}

	// Using the default directory of ~/.credentials
	defaultDir, err := DefaultDirectory()
	if err != nil {
		log.Fatal(err)
	}

	// Create file to store token
	tokenFile, err := CreateTokenFile(defaultDir, "gmail-token.json")
	if err != nil {
		log.Fatal(err)
	}

	// Use Helper struct to create a new gmail api *Service
	gh := Helper{
		Ctx:       context.Background(),
		Config:    config,
		TokenFile: tokenFile,
	}
	srv, err := gh.NewService()
	if err != nil {
		log.Fatalf("Unable to retrieve gmail service %v", err)
	}

	// Test to retrieve labels from gmail
	user := "me"
	r, err := srv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels. %v", err)
	}
	if len(r.Labels) > 0 {
		fmt.Print("Labels:\n")
		for _, l := range r.Labels {
			fmt.Printf("- %s\n", l.Name)
		}
	} else {
		fmt.Print("No labels found.")
	}
}
