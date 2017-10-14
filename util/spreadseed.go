package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	sheets "google.golang.org/api/sheets/v4"
)

const (
	// SpreadSheetID is spread sheet ID.
	SpreadSheetID = "1zbByclh5LJ9Dxa0qGglDSt54lBboiPKynOP_KDcPnRs"
)

// SpreadSheet manages spread sheet
type SpreadSheet struct {
	Seets *sheets.SpreadsheetsService
}

// NewSpreadSheet creates SpreadSheet instance.
func NewSpreadSheet(secretPath string) (*SpreadSheet, error) {
	ctx := context.Background()
	b, err := ioutil.ReadFile(secretPath)
	if err != nil {
		return nil, err
	}

	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return nil, err
	}
	client, err := getClient(ctx, config)
	if err != nil {
		return nil, err
	}

	srv, err := sheets.New(client)
	if err != nil {
		return nil, err
	}

	return &SpreadSheet{Seets: srv.Spreadsheets}, nil
}

// GetLastDate gets last date of spread sheet.
func (s *SpreadSheet) GetLastDate() (*time.Time, error) {
	ret, err := s.Seets.Values.Get(SpreadSheetID, "hayaoki!A2").Do()
	if err != nil {
		return nil, err
	}
	dateStr := ""
	fmt.Println(ret.Values)
	if len(ret.Values) > 0 {
		if len(ret.Values[0]) > 0 {
			dateStr = ret.Values[0][0].(string)
		}
	}
	if dateStr == "" {
		return nil, errors.New("Read cell error")
	}
	date, err := time.Parse("2006/1/2", dateStr)
	if err != nil {
		return nil, err
	}
	return &date, nil
}

// AddNewDate adds new date to hayaoki sheet.
func (s *SpreadSheet) AddNewDate() error {
	// dateStr := [][]interface{}{[]interface{}{time.Now().Format("2006/01/02")}}

	insertRequest := &sheets.Request{InsertDimension: &sheets.InsertDimensionRequest{
		InheritFromBefore: false,
		Range:             &sheets.DimensionRange{Dimension: "ROWS", StartIndex: 1, EndIndex: 2}}}
	_, err := s.Seets.BatchUpdate(SpreadSheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{insertRequest}}).Do()
	if err != nil {
		return err
	}

	today := time.Now().Format("2006/1/2")
	_, err = s.Seets.Values.Update(SpreadSheetID, "hayaoki!A2", &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         [][]interface{}{[]interface{}{today}},
	}).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}

	return nil
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		return nil, err
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		err = saveToken(cacheFile, tok)
		if err != nil {
			return nil, err
		}
	}
	return config.Client(ctx, tok), nil
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, err
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("sheets.googleapis.go-hayaoki-bot.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
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

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
	return nil
}
