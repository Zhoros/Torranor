package main

import (
	"fmt"
	"io"
	"os"
	"log"
	"net/http"
	"time"
	"encoding/json"
	"sync"
	"context"
	"net/url"
	"strconv"
	"github.com/anacrolix/torrent"
)

//We use this map because there can be 2 people downloading the same exact torrent
//we only want to remove the file when nobody is using it anymore
var usageCount map[string]int = make(map[string]int)
//The chance of race condition is extremely small, but just in case..
var usageCountMutex sync.Mutex

func IncrementUsage(t *torrent.Torrent) {

	usageCountMutex.Lock()
    defer usageCountMutex.Unlock()
	_, exists := usageCount[t.InfoHash().HexString()]
	if exists {
		usageCount[t.InfoHash().HexString()]++
	} else {
		usageCount[t.InfoHash().HexString()] = 1
	}

}

func DecrementUsage(t *torrent.Torrent) {	

	usageCountMutex.Lock()
    defer usageCountMutex.Unlock()

	//Just being safe
	_, exists := usageCount[t.InfoHash().HexString()]
	if exists {

		usageCount[t.InfoHash().HexString()]--
		if usageCount[t.InfoHash().HexString()] <= 0 {

			time.AfterFunc(time.Duration(SeedDurationMinute)*time.Minute, func() {

				//We need this check because another user might be downloading the torrent
				//even after AfterFunc is triggered. The file will be eventually cleaned
				//by the latest DecrementUsage() -> AfterFunc() call so it's nothing to be
				//worried about.
				if usageCount[t.InfoHash().HexString()] <= 0 {

					t.Drop()
					delete(usageCount, t.InfoHash().HexString())
					err := os.RemoveAll("./data/" + t.Name())
					if err != nil {
						Log.Printf("error deleting unused torrent file: %v", err)
					}

				}
				
			})

		}

	}

}


func HandleFileList(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	//We make new one each time because this thing is not thread safe
	//expensive but nothing we can do
	client, err := NewTorrentDataFetcher()
	if err != nil {
		log.Printf("failed to create client: %v", err)
	}
	defer client.Close()

	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024*10) //Limit file size
	err = r.ParseMultipartForm(1024*1024*10) //Limit memory usage
	defer func() {

		// Before the multipart form is parsed, it will be written to a temporary folder, make sure to clean it after we are done
		if r.MultipartForm != nil {

			err := r.MultipartForm.RemoveAll()
			if err != nil {
				log.Printf("Error removing MultipartForm file: %v", err)
			}

		}

	}()

	if err != nil {
		log.Printf("Multipart form error: %v", err)
		return
	}

	files := r.MultipartForm.File["torrent"]

	magnet := ""
	if len(files) == 0 {

		magnet = r.MultipartForm.Value["magnet"][0]

	} else {

		file, err := files[0].Open()
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("INVALID"))
			return
		}
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("INVALID"))
			return
		}

		magnet, err = TorrentToMagnet(fileBytes)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("INVALID"))
			return
		}

	}

	t, err := client.AddTorrent(magnet)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("INVALID"))
		return
	}
	defer t.Drop()

	// Wait for info to be fetched
	_, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	<-t.GotInfo()

	var fileNames []string
	for _, f := range t.Files() {
		fileNames = append(fileNames, f.Path())
	}

	response := struct{
		FileNames []string `json:"fileNames"`
		Magnet string   `json:"magnet"`
	} {
		FileNames: fileNames,
		Magnet: magnet, 
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}

}

func HandleFileDownload(w http.ResponseWriter, r *http.Request, client *Torrent) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query()
	encodedMagnet := query.Get("magnet")
	magnet, err := url.QueryUnescape(encodedMagnet)
	if err != nil {

		http.Error(w, "Invalid magnet parameter", http.StatusBadRequest)
		return
	}
	
	fileIndexStr := query.Get("fileIndex")
	fileIndex, err := strconv.Atoi(fileIndexStr)
	if err != nil {

		http.Error(w, "Invalid fileIndex parameter", http.StatusBadRequest)
		return
	}

	t, err := client.AddTorrent(magnet)
	if err != nil {

		http.Error(w, fmt.Sprintf("failed to add torrent: %v", err), http.StatusInternalServerError)
		return

	}

	IncrementUsage(t)
	defer DecrementUsage(t)

	if err := client.StreamFile(w, t, fileIndex); err != nil {
		return
	}

}

func main() {

	//start fresh each time
	GenerateFreshFolder("./data")
	GenerateFreshFolder("./temp")

	InitializeConfig("./config.json")

	//To list files
	http.HandleFunc("/fileList", HandleFileList)

	//To download files
	streamer, err := NewTorrentStreamer()
	if err != nil {
		log.Fatalf("Failed to create torrent streamer: %v", err)
	}
	defer streamer.Close()
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {

		HandleFileDownload(w, r, streamer)

	})

	fs := http.FileServer(http.Dir("./public"))
    http.Handle("/", fs)

	fmt.Println("Server is listening on http://0.0.0.0:80")
	log.Fatal(http.ListenAndServe("0.0.0.0:80", nil))
}
