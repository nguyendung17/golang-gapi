package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}

func main() {
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved client_secret.json.
	config, err := google.ConfigFromJSON(b, gmail.MailGoogleComScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	srv, err := gmail.New(getClient(config))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}
	sendMail(*srv, "@gmail.com", "Dung Nguyen",
		"@gmail.com", "subject", "<h1 style='color:red'>body</h1>")

}
func sendMail(srv gmail.Service, from, sender, to, subject, body string) {
	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	header["Sender"] = fmt.Sprintf("%s <%s>", sender, from)
	header["Content-Type"] = "text/html; charset=\"utf-8\""
	header["Subject"] = subject
	var msg string
	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += fmt.Sprintf("\n%s\n", body)
	message := gmail.Message{
		Raw: encodeWeb64String([]byte(msg)),
	}
	res, err := srv.Users.Messages.Send(from, &message).Do()
	fmt.Println(toJSON(res))
	fmt.Println(toJSON(err))

}

func encodeWeb64String(b []byte) string {
	s := base64.URLEncoding.EncodeToString(b)
	var i = len(s) - 1
	for s[i] == '=' {
		i--
	}
	return s[0 : i+1]
}

func toJSON(needMarshal interface{}) (string, error) {
	if needMarshal == nil || needMarshal == "" {
		return "", nil
	}
	retByte, err := json.Marshal(needMarshal)
	if err != nil || retByte == nil {
		return "", errors.New("marshal fail")
	}
	if string(retByte) == "null" {
		v := reflect.ValueOf(needMarshal)
		switch v.Kind() {
		case reflect.Struct, reflect.Map:
			return "{}", nil
		case reflect.Slice:
			return "[]", nil
		default:
			return "", nil
		}
	}
	return string(retByte), nil
}
