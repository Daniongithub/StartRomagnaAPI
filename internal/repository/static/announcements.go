package static

import (
	"fmt"
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/repository"
	"time"
)

func GetAnnouncements() []model.Announcement {
	var results []model.Announcement
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM announcements")
	if err != nil {
		fmt.Println("GetAnnouncements error:", err)
	}

	return results
}

func InsertAnnouncement(mess string, expiry time.Time) error {
	_, err := repository.DB_STATIC.Exec("INSERT INTO announcements(message, expiry) VALUES(?, ?)", mess, expiry)
	if err != nil {
		fmt.Println("InsertAnnouncements error:", err)
		return err
	}

	return nil
}

func EmptyAnnouncements() error {
	_, err := repository.DB_STATIC.Exec("DELETE FROM announcements")
	if err != nil {
		fmt.Println("EmptyAnnouncements error:", err)
		return err
	}

	return nil
}

func DeleteAnnouncement(id string) error {
	_, err := repository.DB_STATIC.Exec("DELETE FROM announcements WHERE id = ?", id)
	if err != nil {
		fmt.Println("DeleteAnnouncement error:", err)
		return err
	}

	return nil
}