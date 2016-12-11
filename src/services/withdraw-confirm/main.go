package main

import (
	"flag"
	"io/ioutil"
	"log"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const (
	tokenPath = "/tmp/gmail-token.json"
)

var (
	apiKeysPath = flag.String("api-keys", "ansible/secrets/gmail_api_keys.json", "Path to JSON file holding the Gmail API secret keys")
)

func main() {
	flag.Parse()

	b, err := ioutil.ReadFile(*apiKeysPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope, gmail.GmailModifyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	switch cmd := flag.Arg(0); cmd {
	case "authorize":
		authorize(config)
	case "fetch":
		fetch(config)
	default:
		log.Fatal("first argument must be `authorize` or `fetch`")
	}
}
