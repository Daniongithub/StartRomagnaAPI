package realtime

import (
	"fmt"

	"github.com/MobilityData/gtfs-realtime-bindings/golang/gtfs"
)

func SaveAlerts(feeds map[string]*gtfs.FeedMessage) {
	for _, val := range feeds {
		for _, val2 := range val.Entity {
			fmt.Println(val2)
			fmt.Println("------")
		}
	}
}
