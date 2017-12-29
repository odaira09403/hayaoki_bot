package sheets

import (
	"errors"
	"strconv"
	"strings"
	"time"

	sheets "google.golang.org/api/sheets/v4"
)

// KikenSheet manages the sheet which have a kiken logs.
type KikenSheet struct {
	Sheets *sheets.SpreadsheetsService
}

// GetKikenList gets list of kiken dates.
func (s *KikenSheet) GetKikenList() (map[string]string, error) {
	ret, err := s.Sheets.Values.Get(SpreadSheetID, "kiken!A2:B").Do()
	if err != nil {
		return nil, err
	}
	retMap := map[string]string{}
	for _, row := range ret.Values {
		if len(row) < 2 {
			break
		}
		retMap[row[0].(string)] = row[1].(string)
	}
	return retMap, nil
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
		Values: [][]interface{}{{userName}},
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
					contains, err := s.containsDates(userDates[i].(string), dates)
					if err != nil {
						return err
					}
					if contains {
						return errors.New("Date contains previous application date")
					}
					writeDates = userDates[i].(string) + "," + appendDate
				}
			}
			_, err = s.Sheets.Values.Update(SpreadSheetID, "kiken!B"+strconv.Itoa(i+2), &sheets.ValueRange{
				Values: [][]interface{}{{writeDates}},
			}).ValueInputOption("USER_ENTERED").Do()
			if err != nil {
				return err
			}
			return nil
		}
	}
	return errors.New("User not found")
}

func (s *KikenSheet) containsDates(dateStr string, targetDates []time.Time) (bool, error) {
	for _, targetDate := range targetDates {
		contains, err := s.ContainsDate(dateStr, targetDate)
		if err != nil {
			return false, err
		}
		if contains {
			return true, nil
		}
	}
	return false, nil
}

func (s *KikenSheet) ContainsDate(dateStr string, targetDate time.Time) (bool, error) {
	dateList := strings.Split(dateStr, ",")
	for _, date := range dateList {
		dates := strings.Split(date, "-")
		if len(dates) == 1 {
			day, err := time.Parse("2006/1/2", dates[0])
			if err != nil {
				return false, err
			}
			begin := day.Unix()
			end := day.Add(24*time.Hour).Unix()
			if targetDate.Unix() > begin && targetDate.Unix() < end {
				return true, nil
			}
		} else if len(dates) == 2 {
			day1, err := time.Parse("2006/1/2", dates[0])
			if err != nil {
				return false, err
			}
			day2, err := time.Parse("2006/1/2", dates[1])
			if err != nil {
				return false, err
			}
			begin := day1.Unix()
			end := day2.Add(24*time.Hour).Unix()
			if targetDate.Unix() > begin && targetDate.Unix() < end {
				return true, nil
			}
		}
	}
	return false, nil
}
