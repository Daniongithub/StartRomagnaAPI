package static

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
)

func GetAnnouncements() []model.Announcement {
	var results []model.Announcement
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM announcements")
	if err != nil {
		fmt.Println("GetAnnouncements error:", err)
	}

	return results
}