package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/fsnotify/fsnotify"
)

func main() {
	// Get the directory to watch from the environment variable
	watchDir := os.Getenv("WATCH_DIR")
	defaultPath := ""
	switch runtime.GOOS {
	case "windows":
		defaultPath = "C:\\default\\path\\file.txt" // may not need the double backslash, have to check
	case "linux":
		defaultPath = "/default/path/file.txt"
	default:
		defaultPath = "/default/path/file.txt"
	}

	if watchDir == "" {
		// Set a default directory if the environment variable is not set
		//change the watch dir to unix path based on os path separator

		watchDir = defaultPath
		log.Println("WATCH_DIR environment variable not set. Using default:", watchDir)
	}

	// Convert the directory path to the forward slash separator
	watchDir = filepath.ToSlash(watchDir)

	// Create a new watcher instance
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Add the directory to watch
	err = watcher.Add(watchDir)
	if err != nil {
		log.Fatal(err)
	}

	// Start an infinite loop to process file events
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Handle different types of events
			switch event.Op {
			case fsnotify.Create:
				log.Println("Created:", event.Name)
				uploadToS3(event.Name, "bucket-name")
			case fsnotify.Write:
				log.Println("Modified:", event.Name)
				uploadToS3(event.Name, "bucket-name")
			case fsnotify.Remove:
				log.Println("Removed:", event.Name)
			case fsnotify.Rename:
				log.Println("Renamed:", event.Name)
			case fsnotify.Chmod:
				log.Println("Permission changed:", event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Error:", err)
		}
	}
}

func uploadToS3(filePath, bucketName string) {
	// Convert the file path to the forward slash separator
	filePath = filepath.ToSlash(filePath)

	cmd := exec.Command("aws", "s3", "cp", filePath, "s3://"+bucketName)

	// Set AWS credentials as environment variables if necessary
	// cmd.Env = []string{"AWS_ACCESS_KEY_ID=YOUR_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY=YOUR_SECRET_KEY"}

	// Run the AWS CLI command and handle any errors
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
