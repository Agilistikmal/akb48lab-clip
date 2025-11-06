package main

import (
	"os"

	"github.com/agilistikmal/akb48lab-clip/internal/clip/services"
	"github.com/agilistikmal/live-recorder/pkg/recorder"
	"github.com/agilistikmal/live-recorder/pkg/recorder/live"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	liveQuery := recorder.LiveQuery{
		Platforms:            []string{recorder.PlatformShowroom},
		StreamerUsernameLike: "*48*",
	}
	recorder := live.NewRecorder(&liveQuery)
	clipService := services.NewClipService(liveQuery, recorder, "./tmp/clips")
	clipService.SetYoutubeApiKey(os.Getenv("YOUTUBE_API_KEY"))
	clipService.WatchForClips()

	select {}
}
