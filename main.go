package main

import (
	"bytes"
	"encoding/json"
	"github.com/alexrocco/file-watch/watch"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {
	log := logrus.New()

	watchPath := os.Getenv("WATCH_PATH")

	watcher := watch.NewInotifyWatcher(log)

	fileModified := make(chan watch.FileModified)
	watchStarted := make(chan bool)

	go func() {
		err := watcher.Watches(watchPath, fileModified, watchStarted)
		if err != nil {
			log.Panicf("error on Watcher: %q, \nexiting...", err.Error())
		}
	}()

	for {
		select {
		// Wait for the watcher to start
		case <-watchStarted:
			for {
				select {
				case fileModified := <-fileModified:
					log.Debug(fileModified)

					operation := ""
					switch fileModified.FileOperation {
					case watch.Create:
						operation = "create"
					case watch.Remove:
						operation = "remove"
					case watch.Write:
						operation = "write"
					case watch.Rename:
						operation = "rename"
					}

					msg := map[string]string{
						"path":      fileModified.Path,
						"operation": operation,
					}

					body, err := json.Marshal(msg)
					if err != nil {
						log.Errorf("error parsing file modified with path %q and operation %q", fileModified.Path, fileModified.FileOperation)
					}

					reader:= bytes.NewReader(body)

					resp, err := http.Post("http://localhost:8080/add", "application/json", reader)
					if err != nil || resp.StatusCode != http.StatusOK {
						log.Errorf("error adding file modified with path %q and operation %q to the queue", fileModified.Path, fileModified.FileOperation)
					}
				}
			}
		}
	}
}
