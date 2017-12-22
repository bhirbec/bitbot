package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
)

const (
	// gmail user
	me = "me"
	// kraken withdran confirmation URL
	confirmWithdrawURL = "https://www.kraken.com/withdrawal-approve"
	// Label was created manually in gmail UI.
	label = "bitbot-confirmed"
	// "bitbot-confirmed" label id. TODO: would be nice to automate label setup
	labelId = "Label_9"
	// ticker duration in minutes
	dur = 1
)

type searchOptions struct {
	key   string
	value string
}

func (s searchOptions) Get() (key, value string) {
	return s.key, s.value
}

func fetch(config *oauth2.Config) {
	fmt.Println("Starting...")
	c := time.Tick(dur * time.Minute)

	for _ = range c {
		work(config)
		fmt.Println("Waiting...")
	}
}

func work(config *oauth2.Config) {
	ctx := context.Background()
	client, err := getClient(ctx, config)
	if err != nil {
		log.Fatalf("getClient() failed: %s", err)
	}

	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client: %s", err)
	}

	s := searchOptions{"q", confirmWithdrawURL + " -label:" + label}
	r, err := srv.Users.Messages.List(me).Do(s)
	if err != nil {
		log.Fatalf("srv.Users.Messages.List(): %s", err)
	}

	for _, m := range r.Messages {
		fmt.Printf("Confirming Kraken withdraw (email id: %s)...\n", m.Id)

		msg, err := fetchMessage(srv, m.Id)
		if err != nil {
			fmt.Printf("fetchMessage() failed - %s\n", err)
			continue
		}

		url := extractURL(msg)
		if url == "" {
			fmt.Printf("extractURL() returned empty string\n")
			continue
		}

		err = confirmWithdraw(url)
		if err != nil {
			fmt.Printf("confirmWithdraw() failed - %s\n", err)
			continue
		}

		params := &gmail.ModifyMessageRequest{
			AddLabelIds:     []string{labelId},
			ForceSendFields: []string{"AddLabelIds"},
		}

		_, err = srv.Users.Messages.Modify(me, m.Id, params).Do()
		if err != nil {
			fmt.Printf("srv.Users.Messages.Modify() failed - %s\n", err)
			continue
		}

		fmt.Printf("Message %s completed \n", m.Id)
	}
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	f, err := os.Open(tokenPath)
	if err != nil {
		return nil, err
	}

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	return config.Client(ctx, tok), nil
}

func fetchMessage(srv *gmail.Service, id string) (string, error) {
	m, err := srv.Users.Messages.Get(me, id).Do()
	if err != nil {
		return "", fmt.Errorf("srv.Users.Messages.Get() failed: %s", err)
	}

	str := m.Payload.Parts[0].Body.Data
	data, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		return "", fmt.Errorf("base64.URLEncoding.DecodeString failed - %s", err)
	}

	return string(data), nil
}

func extractURL(data string) string {
	lines := strings.Split(data, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, confirmWithdrawURL) {
			return line
		}
	}
	return ""
}

func confirmWithdraw(rawurl string) error {
	u, err := url.Parse(rawurl)
	if err != nil {
		return fmt.Errorf("url.Parse() failed - %s\n", err)
	}

	data := strings.NewReader(u.Query().Encode())
	req, err := http.NewRequest("POST", confirmWithdrawURL, data)
	if err != nil {
		return fmt.Errorf("http.NewRequest() failed - %s\n", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("POST failed - %s\n", err)
	}

	if resp.StatusCode != http.StatusOK {
		status := http.StatusText(resp.StatusCode)
		return fmt.Errorf("POST return error - %s\n", status)
	}

	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll failed() - %s\n", err)
	}

	return nil
}
