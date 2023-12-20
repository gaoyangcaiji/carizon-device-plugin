package main

import (
	"errors"
	"log"
	"os"
	"os/signal"

	"github.com/fsnotify/fsnotify"
	"github.com/go-resty/resty/v2"
)

func checkErr(err error) {
	if err != nil {
		log.Panicln("Fatal:", err)
	}
}

func newFSWatcher(files ...string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		err = watcher.Add(f)
		if err != nil {
			watcher.Close()
			return nil, err
		}
	}

	return watcher, nil
}

func newOSWatcher(sigs ...os.Signal) chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)

	return sigChan
}

func checkHTTPResponse(r *resty.Response, err error) (int, []byte, error) {
	code := r.StatusCode()
	if err != nil {
		return code, nil, err
	}
	if r.IsError() {
		return code, nil, errors.New(r.String())
	}
	return code, r.Body(), nil
}
