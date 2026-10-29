package static

import (
	"database/sql"
	"fmt"

	"startromagnaapi/internal/repository"
)

func GetCalDatesRange(basin string) (string, string) {
	var min, max sql.NullTime

	err := repository.DB_STATIC.QueryRow("SELECT MIN(date), MAX(date) FROM calendar_dates WHERE basin = ?", basin).Scan(&min, &max)

	if err != nil {
		fmt.Println("GetCalDatesRange error:", err)
		return "", ""
	}

	if !min.Valid || !max.Valid {
		return "", ""
	}
	return min.Time.Format("2006-01-02"), max.Time.Format("2006-01-02")
}
