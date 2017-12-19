package sheets

import (
	"errors"
	"time"

	sheets "google.golang.org/api/sheets/v4"
)

// HayaokiSheet manages the sheet which have a hayaoki logs.
type HayaokiSheet struct {
	Sheets *sheets.SpreadsheetsService
}

// GetLastDate gets last date of spread sheet.
func (s *HayaokiSheet) GetLastDate() (*time.Time, error) {
	ret, err := s.Sheets.Values.Get(SpreadSheetID, "hayaoki!A2").Do()
	if err != nil {
		return nil, err
	}
	dateStr := ""
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
func (s *HayaokiSheet) AddNewDate() error {
	insertRequest := &sheets.Request{InsertDimension: &sheets.InsertDimensionRequest{
		InheritFromBefore: false,
		Range:             &sheets.DimensionRange{Dimension: "ROWS", StartIndex: 1, EndIndex: 2}}}
	_, err := s.Sheets.BatchUpdate(SpreadSheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{insertRequest}}).Do()
	if err != nil {
		return err
	}

	today := time.Now().Format("2006/1/2")
	_, err = s.Sheets.Values.Update(SpreadSheetID, "hayaoki!A2", &sheets.ValueRange{
		Values: [][]interface{}{{today}},
	}).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}

	return nil
}

// UserExists gets user index of column from hayaoki sheet.
func (s *HayaokiSheet) UserExists(userName string) (bool, error) {
	ret, err := s.Sheets.Values.Get(SpreadSheetID, "hayaoki!B1:1").Do()
	if err != nil {
		return false, err
	}
	if len(ret.Values) == 0 {
		return false, nil
	}
	users := ret.Values[0]
	for _, user := range users {
		if user.(string) == userName {
			return true, nil
		}
	}
	return false, nil
}

// AddNewUser adds new user.
func (s *HayaokiSheet) AddNewUser(userName string) error {
	insertRequest := &sheets.Request{InsertDimension: &sheets.InsertDimensionRequest{
		InheritFromBefore: false,
		Range:             &sheets.DimensionRange{Dimension: "COLUMNS", StartIndex: 1, EndIndex: 2}}}
	_, err := s.Sheets.BatchUpdate(SpreadSheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{insertRequest}}).Do()
	if err != nil {
		return err
	}

	_, err = s.Sheets.Values.Update(SpreadSheetID, "hayaoki!B1", &sheets.ValueRange{
		Values: [][]interface{}{{userName}},
	}).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}
	return nil
}

// SetHayaokiFlag sets hayaoki flag of the spesicied user.
func (s *HayaokiSheet) SetHayaokiFlag(now time.Time, userName string) error {
	ret, err := s.Sheets.Values.Get(SpreadSheetID, "hayaoki!B1:1").Do()
	if err != nil {
		return err
	}
	if len(ret.Values) == 0 {
		return errors.New("User not found")
	}
	users := ret.Values[0]
	for i, user := range users {
		if user.(string) == userName {
			_, err = s.Sheets.Values.Update(SpreadSheetID, "hayaoki!"+string('B'+i)+"2", &sheets.ValueRange{
				Values: [][]interface{}{{now.Format("15:04")}},
			}).ValueInputOption("USER_ENTERED").Do()
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("User not found")
}
