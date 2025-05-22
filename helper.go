package main

import (
	"os"
	"log"
)

func GenerateFreshFolder(name string) {

	if _, err := os.Stat(name); err == nil {
		err = os.RemoveAll(name)
		if err != nil {
			log.Fatalf("Failed to delete data folder: %v", err)
		}
		err := os.MkdirAll(name, 0755)
		if err != nil {
			log.Fatalf("Failed to create data folder: %v", err)
		}
	} else if !os.IsNotExist(err) {
		log.Fatalf("Folder Stat() error: %v", err)
	}

}

