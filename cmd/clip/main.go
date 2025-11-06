package main

import (
	"github.com/agilistikmal/akb48lab-clip/internal/clip/services"
	"github.com/agilistikmal/live-recorder/pkg/recorder"
	"github.com/agilistikmal/live-recorder/pkg/recorder/live"
)

func main() {
	liveQuery := recorder.LiveQuery{
		Platforms:            []string{recorder.PlatformShowroom},
		StreamerUsernameLike: "*48*",
	}
	recorder := live.NewRecorder(&liveQuery)
	clipService := services.NewClipService(liveQuery, recorder, "./tmp/clips")
	clipService.WatchForClips()

	select {}
}
