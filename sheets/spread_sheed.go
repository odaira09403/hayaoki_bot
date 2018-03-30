package sheets

import (
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

const (
	// SpreadSheetID is spread sheet ID.
	SpreadSheetID = "1zbByclh5LJ9Dxa0qGglDSt54lBboiPKynOP_KDcPnRs"
)

// SpreadSheet manages spread sheet
type SpreadSheet struct {
	Hayaoki *HayaokiSheet
	Kiken   *KikenSheet
}

// NewSpreadSheet creates SpreadSheet instance.
func NewSpreadSheet(ctx context.Context) (*SpreadSheet, error) {
	client, err := google.DefaultClient(ctx, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, err
	}

	srv, err := sheets.New(client)
	if err != nil {
		return nil, err
	}

	return &SpreadSheet{Hayaoki: &HayaokiSheet{Sheets: srv.Spreadsheets}, Kiken: &KikenSheet{Sheets: srv.Spreadsheets}}, nil
}
