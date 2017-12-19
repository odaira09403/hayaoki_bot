package app

import (
	"time"

	"github.com/odaira09403/hayaoki_bot/handler"
)

const location = "Asia/Tokyo"

func init() {
	// Init timezone
	loc, err := time.LoadLocation(location)
	if err != nil {
		loc = time.FixedZone(location, 9*60*60)
	}
	time.Local = loc

	slashHandler := handler.NewSlashHandler()
	slashHandler.Run()
}
