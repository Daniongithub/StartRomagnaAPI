package utils

import (
	"time"
	_ "time/tzdata"
)

var rome, _ = time.LoadLocation("Europe/Rome")

func RomeNow() time.Time { return time.Now().In(rome) }
