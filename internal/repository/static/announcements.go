package static

import (
	"startromagnaapi/internal/model"
	"startromagnaapi/internal/notify"
	"startromagnaapi/internal/repository"
	"time"
)

func GetAnnouncements() []model.Announcement {
	var results []model.Announcement
	err := repository.DB_STATIC.Select(&results, "SELECT * FROM announcements")
	if err != nil {
		notify.ReportError("Failed to read announcements", err)
	}

	return results
}

func InsertAnnouncement(mess string, expiry time.Time) error {
	_, err := repository.DB_STATIC.Exec("INSERT INTO announcements(message, expiry) VALUES(?, ?)", mess, expiry)
	if err != nil {
		notify.ReportError("Failed to insert an announcement", err)
		return err
	}

	return nil
}

func EmptyAnnouncements() error {
	_, err := repository.DB_STATIC.Exec("DELETE FROM announcements")
	if err != nil {
		notify.ReportError("Failed to clear announcements", err)
		return err
	}

	return nil
}

func DeleteAnnouncement(id string) error {
	_, err := repository.DB_STATIC.Exec("DELETE FROM announcements WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}
