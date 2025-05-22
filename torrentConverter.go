package main

import (

	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/anacrolix/torrent/bencode"
)

type TorrentFile struct {
	Info         map[string]interface{} `bencode:"info"`
	Announce     string                 `bencode:"announce"`
	AnnounceList [][]string             `bencode:"announce-list"`
	CreationDate int64                  `bencode:"creation date"`
	Comment      string                 `bencode:"comment"`
	CreatedBy    string                 `bencode:"created by"`
	Encoding     string                 `bencode:"encoding"`
}


func TorrentToMagnet(torrentData []byte) (string, error) {

	var torrent TorrentFile
	err := bencode.Unmarshal(torrentData, &torrent)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal torrent data: %w", err)
	}

	infoData, err := bencode.Marshal(torrent.Info)
	if err != nil {
		return "", fmt.Errorf("failed to marshal info dictionary: %w", err)
	}

	hash := sha1.Sum(infoData)
	infoHash := hex.EncodeToString(hash[:])

	magnet := fmt.Sprintf("magnet:?xt=urn:btih:%s", infoHash)

	if name, ok := torrent.Info["name"].(string); ok {
		magnet += fmt.Sprintf("&dn=%s", name)
	}

	if torrent.Announce != "" {
		magnet += fmt.Sprintf("&tr=%s", torrent.Announce)
	}

	for _, tier := range torrent.AnnounceList {
		for _, tracker := range tier {
			magnet += fmt.Sprintf("&tr=%s", tracker)
		}
	}

	return magnet, nil
}

