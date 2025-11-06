package services

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/agilistikmal/live-recorder/pkg/recorder"
	"github.com/agilistikmal/live-recorder/pkg/recorder/live"
	"github.com/agilistikmal/live-recorder/pkg/recorder/watch"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type ClipService struct {
	liveQuery      recorder.LiveQuery
	recorder       recorder.Recorder
	watcher        *watch.WatchLive
	youtubeService *youtube.Service
}

func NewClipService(liveQuery recorder.LiveQuery, recorder recorder.Recorder, clipDir string) *ClipService {
	return &ClipService{
		liveQuery: liveQuery,
		recorder:  live.NewRecorder(&liveQuery),
		watcher:   watch.NewWatchLive(recorder, clipDir),
	}
}

func (s *ClipService) SetYoutubeApiKey(apiKey string) {
	youtubeService, err := youtube.NewService(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		logrus.Fatalf("Failed to create YouTube client: %v", err)
	}
	s.youtubeService = youtubeService
}

func (s *ClipService) WatchForClips() {
	liveChan := make(chan *recorder.Live, 100)
	statusChan := make(chan *watch.StatusUpdate, 100)
	s.watcher.SetLiveChannel(liveChan)
	s.watcher.SetStatusChannel(statusChan)

	go func() {
		for live := range liveChan {
			logrus.Infof("New live detected: %s", live.Streamer.Username)
		}
	}()

	uploadQueueChan := make(chan *watch.RecordingInfo, 100)
	go func() {
		for status := range statusChan {
			switch status.Status {
			case watch.StatusInProgress:
				logrus.Infof("Recording clip for %s", status.Info.Live.Streamer.Username)
			case watch.StatusCompleted:
				uploadQueueChan <- status.Info
				logrus.Infof("Clip for %s completed", status.Info.Live.Streamer.Username)
			case watch.StatusFailed:
				logrus.Errorf("Failed to record clip for %s", status.Info.Live.Streamer.Username)
			}
		}
	}()

	go s.watcher.StartWatchMode()

	go s.UploadClips(uploadQueueChan)
}

func (s *ClipService) UploadClips(uploadQueueChan chan *watch.RecordingInfo) error {
	if s.youtubeService == nil {
		logrus.Warn("YouTube service not initialized. Skipping upload.")
		return errors.New("YouTube service not initialized")
	}

	for info := range uploadQueueChan {
		video, err := os.Open(info.FilePath)
		if err != nil {
			return err
		}
		defer video.Close()

		startedAt := info.Live.StartedAt.Format("2006-01-02 15:04")
		duration := info.CompletedAt.Sub(*info.Live.StartedAt).Round(time.Second).String()
		title := fmt.Sprintf("[SUB EN/JP/ID] %s - %s", info.Live.Streamer.Name, startedAt)
		description := fmt.Sprintf(
			"%s (%s) %s live\n%s\n\nStarted at: %s\nDuration: %s",
			info.Live.Streamer.Name, info.Live.Streamer.Username, info.Live.Platform, info.Live.Title, startedAt, duration,
		)

		parts := []string{"snippet", "status"}
		youtubeVideo := &youtube.Video{
			Snippet: &youtube.VideoSnippet{
				Title:       title,
				Description: description,
				Tags:        []string{info.Live.Streamer.Username, info.Live.Platform, info.Live.Streamer.Name},
				Thumbnails: &youtube.ThumbnailDetails{
					Default: &youtube.Thumbnail{
						Url: info.Live.ImageUrl,
					},
				},
				CategoryId: "22",
			},
		}
		res, err := s.youtubeService.Videos.Insert(parts, youtubeVideo).Do()
		if err != nil {
			return err
		}

		logrus.Infof("Uploaded video: %s", res.Id)
	}
	return nil
}
