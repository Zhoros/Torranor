# Torranor
Torranor is a web app that allows you to download torrents **DIRECTLY** from your web browser

It's different from WebTorrent / WebRTC torrent. WebTorrent can only connect to another clients that support WebRTC. This means that the download speed will be slower than traditional torrent, especially when there is very little seeder. Torranor connects directly to the BitTorrent network and stream the download directly to your browser in real time

It also supports seeding with plenty of configurations

### <ins>**Please star this repo if you find it useful, thank you!**</ins>

## Screenshot
![image](https://github.com/user-attachments/assets/0a2b2d0a-440f-4a12-beff-4f2cf2471d5f)

## Settings
Inside `config.json` file, there are several values that you can modify:  
- `uploadKBps` = limit seeding upload speed in **KILOBYTE**/s  
- `uploadBurstSizeKB` = maximum seeding burst upload in **KILOBYTE** if there has been no upload for a while. e.g. if you set this to 10240 and your upload speed has been 0 KBps for the last 10 seconds, it will burst 10240 kb all at once to compensate, but will never exceed this limit. If you don't understand it, just set it to the same value as `UploadKBps`  
- `seedDurationMinute` = after you finish downloading a file from your browser, the server will **automatically** start seeding your file for the specified duration. After the duration ends, the file will be automatically deleted to free up space
- `listeningPort` = your torrent listening port, make sure that this port is accessible and match it with the port mapping inside your `docker-compose.yml` file

## Important
- All your files that are being seeded by the server are temporary. It will all be **DELETED** when you restart the server
- Fetching and download might take up to a minute to start, it's not broken / stuck
- The Dockerfile build can take up to 5 minutes on 1 core CPU, it's not broken / stuck

## Install with docker
1. Make sure the listening port inside `config.json` and your `docker-compose.yml` port mapping matches
2. `docker compose up -d`

## Install without docker
1. Install Go compiler
2. Make sure the listening port inside `config.json` is accessible
3. `go build .`
