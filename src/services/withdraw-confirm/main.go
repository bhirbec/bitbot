package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os/user"
	"path/filepath"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var (
	tokenPath   string
	apiKeysPath = flag.String("api-keys", "ansible/secrets/gmail_api_keys.json", "Path to JSON file holding the Gmail API secret keys")
)

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("user.Current() failed: %s\n", err)
	}
	tokenPath = filepath.Join(usr.HomeDir, "gmail-token.json")
}

func main() {
	flag.Parse()

	b, err := ioutil.ReadFile(*apiKeysPath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, gmail.MailGoogleComScope)
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
