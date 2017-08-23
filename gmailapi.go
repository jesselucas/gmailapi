// Package gmailapi provides helpers for gmail's Go api.
// It is based on the Go Quickstart https://developers.google.com/gmail/api/quickstart/go
package gmailapi

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// ConfigFromJSON is a helper function to take a JSONPath and scope that returns an *oauth2.Config
func ConfigFromJSON(jsonPath, scope string) (*oauth2.Config, error) {
	b, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	config, err := google.ConfigFromJSON(b, scope)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}

	return config, nil
}

// DefaultDirectory is a helper function to return a .credentials
// folder in the current users home directory
func DefaultDirectory() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, ".credentials"), nil
}

// CreateTokenFile helper function that creates a file for the token cache
// based on the supplied directory and filename
func CreateTokenFile(directory, filename string) (string, error) {
	tokenCacheDir := directory
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape(filename)), nil
}

// Helper struct for the gmail api
// Typical usage is to use the helper functions CreateTokenFile and ConfigFromJSON
type Helper struct {
	Ctx       context.Context
	TokenFile string
	Config    *oauth2.Config
}

// NewService returns a *gmail.Service based on the properties of the Helper struct
func (h *Helper) NewService() (*gmail.Service, error) {
	client, err := h.httpClient()
	if err != nil {
		return nil, err
	}

	srv, err := gmail.New(client)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve gmail Client %v", err)
	}

	return srv, nil
}

// httpClient creates or returns an *http.Client to be used with the gmail.New() method
func (h *Helper) httpClient() (*http.Client, error) {
	token, err := tokenFromFile(h.TokenFile)
	if err != nil {
		token, err = tokenFromWeb(h.Config)
		if err != nil {
			return nil, err
		}
		saveToken(h.TokenFile, token)
	}
	return h.Config.Client(h.Ctx, token), nil
}

// tokenFromFile retrieves a token
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// tokenFromWeb uses *oauth2.Config to request a token from the
// web and displays a prompt in the command line
func tokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	io.WriteString(os.Stdout, "Enter code> ")

	bs := bufio.NewScanner(os.Stdin)
	if !bs.Scan() {
		return nil, errors.New("Unable to read authorization code")
	}
	code := strings.TrimSpace(bs.Text())

	token, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token from web %v", err)
	}
	return token, nil
}

// saveToken saves an *oauth2.Token as json to a file
func saveToken(file string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)

	return nil
}
