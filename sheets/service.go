package sheets

import (
	"context"
	"fmt"

	sheetsv4 "google.golang.org/api/sheets/v4"
)

func AppendRow(
	ctx context.Context,
	service *sheetsv4.Service,
	spreadsheetID string,
	sheetName string,
	values []interface{},
) error {
	if spreadsheetID == "" {
		return fmt.Errorf("spreadsheet ID is empty")
	}
	
	rangeData := fmt.Sprintf("%s!A:G", sheetName)
	valueRange := &sheetsv4.ValueRange{
		Values: [][]interface{}{values},
	}

	_, err := service.Spreadsheets.Values.Append(
		spreadsheetID,
		rangeData,
		valueRange,
	).ValueInputOption("USER_ENTERED").Do()
	
	return err
}