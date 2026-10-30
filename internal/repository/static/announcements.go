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

func InsertAnnouncement(mess string) error {
	_, err := repository.DB_STATIC.Exec("INSERT INTO announcements(message) VALUES(?)", mess)
	if err != nil {
		fmt.Println("InsertAnnouncements error:", err)
		return err
	}

	return nil
}
