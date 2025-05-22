package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"
	"golang.org/x/time/rate"
	"github.com/anacrolix/torrent"
)

type Torrent struct {
	client *torrent.Client
}

func NewTorrentStreamer() (*Torrent, error) {

	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = "./data"
	cfg.UploadRateLimiter = rate.NewLimiter(rate.Limit(UploadKBps * 1024), int(UploadBurstSizeKB * 1024))
	cfg.ListenPort = int(ListeningPort)
	cfg.Seed = true

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create torrent client: %v", err)
	}

	return &Torrent{client: client}, nil
}

func NewTorrentDataFetcher() (*Torrent, error) {

	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = "./temp"
	cfg.NoUpload = true
	cfg.Seed = false
	cfg.ListenPort = 0

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create torrent client: %v", err)
	}

	return &Torrent{client: client}, nil

}

func (ts *Torrent) AddTorrent(magnetURI string) (*torrent.Torrent, error) {
	t, err := ts.client.AddMagnet(magnetURI)
	if err != nil {
		return nil, fmt.Errorf("failed to add magnet URI: %v", err)
	}
	select {
		case <-t.GotInfo():
			return t, nil
		case <-time.After(time.Duration(1) * time.Minute):
			return nil, errors.New("timeout waiting for torrent info")
	}
}

func (ts *Torrent) StreamFile(w http.ResponseWriter, t *torrent.Torrent, fileIndex int) error {
	if len(t.Files()) == 0 {
		return errors.New("no files in torrent")
	}

	if fileIndex < 0 || fileIndex >= len(t.Files()) {
		return fmt.Errorf("invalid file index %d (available: 0-%d)", fileIndex, len(t.Files())-1)
	}

	file := t.Files()[fileIndex]
	reader := file.NewReader()
	defer reader.Close()

	file.SetPriority(torrent.PiecePriorityNormal)
	reader.SetReadahead(5 << 20) //5MB

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(file.Path())))
	w.Header().Set("Content-Type", "application/octet-stream")

	_, err := io.Copy(w, reader)
	return err
}

func (ts *Torrent) Close() {
	ts.client.Close()
}
