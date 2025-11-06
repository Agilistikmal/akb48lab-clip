package services

import (
	"github.com/agilistikmal/live-recorder/pkg/recorder"
	"github.com/agilistikmal/live-recorder/pkg/recorder/live"
	"github.com/agilistikmal/live-recorder/pkg/recorder/watch"
	"github.com/sirupsen/logrus"
)

type ClipService struct {
	liveQuery recorder.LiveQuery
	recorder  recorder.Recorder
	watcher   *watch.WatchLive
}

func NewClipService(liveQuery recorder.LiveQuery, recorder recorder.Recorder, clipDir string) *ClipService {
	return &ClipService{
		liveQuery: liveQuery,
		recorder:  live.NewRecorder(&liveQuery),
		watcher:   watch.NewWatchLive(recorder, clipDir),
	}
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

	go func() {
		for status := range statusChan {
			switch status.Status {
			case watch.StatusInProgress:
				logrus.Infof("Recording clip for %s", status.Info.Live.Streamer.Username)
			case watch.StatusCompleted:
				logrus.Infof("Clip for %s completed", status.Info.Live.Streamer.Username)
			case watch.StatusFailed:
				logrus.Errorf("Failed to record clip for %s", status.Info.Live.Streamer.Username)
			}
		}
	}()

	go s.watcher.StartWatchMode()
}
