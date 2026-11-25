package static

import (
	"database/sql"

	"startromagnaapi/internal/notify"
	"startromagnaapi/internal/repository"
)

func GetCalDatesRange(basin string) (string, string) {
	var min, max sql.NullTime

	err := repository.DB_STATIC.QueryRow("SELECT MIN(date), MAX(date) FROM calendar_dates WHERE basin = ?", basin).Scan(&min, &max)

	if err != nil {
		notify.ReportError("Failed to read a basin's calendar range", err)
		return "", ""
	}

	if !min.Valid || !max.Valid {
		return "", ""
	}
	return min.Time.Format("2006-01-02"), max.Time.Format("2006-01-02")
}
