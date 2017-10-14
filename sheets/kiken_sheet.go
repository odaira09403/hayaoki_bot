package sheets

import sheets "google.golang.org/api/sheets/v4"

// KikenSheet manages the sheet which have a kiken logs.
type KikenSheet struct {
	Sheets *sheets.SpreadsheetsService
}
