package sheets

import (
	"errors"
	"strconv"
	"time"

	sheets "google.golang.org/api/sheets/v4"
)

// KikenSheet manages the sheet which have a kiken logs.
type KikenSheet struct {
	Sheets *sheets.SpreadsheetsService
}

// UserExists gets user index of column from hayaoki sheet.
func (s *KikenSheet) UserExists(userName string) (bool, error) {
	ret, err := s.Sheets.Values.Get(SpreadSheetID, "kiken!A2:A").MajorDimension("COLUMNS").Do()
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
func (s *KikenSheet) AddNewUser(userName string) error {
	res, err := s.Sheets.Get(SpreadSheetID).Ranges("kiken").Do()
	if err != nil {
		return err
	}
	if len(res.Sheets) == 0 {
		return errors.New("Sheet not found")
	}
	sheetID := res.Sheets[0].Properties.SheetId
	insertRequest := &sheets.Request{InsertDimension: &sheets.InsertDimensionRequest{
		InheritFromBefore: false,
		Range:             &sheets.DimensionRange{SheetId: sheetID, Dimension: "ROWS", StartIndex: 1, EndIndex: 2}}}
	_, err = s.Sheets.BatchUpdate(SpreadSheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{insertRequest}}).Do()
	if err != nil {
		return err
	}

	_, err = s.Sheets.Values.Update(SpreadSheetID, "kiken!A2", &sheets.ValueRange{
		Values: [][]interface{}{[]interface{}{userName}},
	}).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}
	return nil
}

// AddDate adds new date of kiken.
func (s *KikenSheet) AddDate(userName string, dates []time.Time) error {
	appendDate := ""
	if len(dates) == 1 {
		appendDate = dates[0].Format("2006/01/02")
	} else if len(dates) == 2 {
		appendDate = dates[0].Format("2006/01/02") + "-" + dates[1].Format("2006/01/02")
	} else {
		return errors.New("Invalid length of dates")
	}
	ret, err := s.Sheets.Values.Get(SpreadSheetID, "kiken!A2:B").MajorDimension("COLUMNS").Do()
	if err != nil {
		return err
	}
	if len(ret.Values) < 1 {
		return errors.New("User not found")
	}
	users := ret.Values[0]
	userDates := []interface{}{}
	if len(ret.Values) >= 2 {
		userDates = ret.Values[1]
	}
	for i, user := range users {
		if user.(string) == userName {
			writeDates := appendDate
			if len(userDates) > i {
				if userDates[i].(string) != "" {
					writeDates = userDates[i].(string) + "," + appendDate
				}
			}
			_, err = s.Sheets.Values.Update(SpreadSheetID, "kiken!B"+strconv.Itoa(i+2), &sheets.ValueRange{
				Values: [][]interface{}{[]interface{}{writeDates}},
			}).ValueInputOption("USER_ENTERED").Do()
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("User not found")
}
