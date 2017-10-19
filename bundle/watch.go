// Package bundle provides a schema and resolver for bundle remote bundle management.
package bundle

import (
	"log"
	"os"
	"path/filepath"
	"github.com/fsnotify/fsnotify"
)

func watcherWalkTreeFunction(watcher *fsnotify.Watcher) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		
		if err != nil {
			return err
		}
		err = watcher.Add(path)
		if err != nil {
			log.Fatal(err)
			return err
		}
		
		return nil;
	}
}


type Handler interface {
        ServeWatcherEvent(*fsnotify.Watcher, *fsnotify.Event, *Index) *Index
}

type Adapter func(Handler) Handler


func Adapt(h Handler, adapters ...Adapter) Handler {
	for i := len(adapters); i > 0; i-- {
		h = adapters[i - 1](h)
	}
	return h
}

type HandlerFunc func(*fsnotify.Watcher, *fsnotify.Event, *Index) *Index

func (f HandlerFunc) ServeWatcherEvent(watcher *fsnotify.Watcher, event *fsnotify.Event, index *Index) *Index {
  	return f(watcher, event, index)
}

func UpdateWatcher() Adapter {
	return func(h Handler) Handler {
		return HandlerFunc(func(watcher *fsnotify.Watcher, event *fsnotify.Event, index *Index) *Index {
			if event.Op&fsnotify.Create == fsnotify.Create {
				log.Println("UpdateWatcher: add ", event.Name)

				fi, err := os.Stat(event.Name)
				if !os.IsNotExist(err) && fi.IsDir() {
					//log.Println("UpdateWatcher: add ", event.Name)
					filepath.Walk(event.Name, watcherWalkTreeFunction(watcher))
				}
			} else if event.Op&fsnotify.Rename == fsnotify.Rename {
				//log.Println("UpdateWatcher: remove ", event.Name)
				watcher.Remove(event.Name)
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				//log.Println("UpdateWatcher: remove ", event.Name)
				watcher.Remove(event.Name)
			}
			
			
			return h.ServeWatcherEvent(watcher, event, index)  
		})
	}
}


func LogEvent() Adapter {
	return func(h Handler) Handler {
		return HandlerFunc(func(watcher *fsnotify.Watcher, event *fsnotify.Event, index *Index) *Index {
			log.Println("Event: ", event)
			return h.ServeWatcherEvent(watcher, event, index)  
		})
	}
}


type NoopHandler struct {}

func (h *NoopHandler) ServeWatcherEvent(watcher *fsnotify.Watcher, event *fsnotify.Event, index *Index) *Index {
	return index
}


func StartWatching(watcher *fsnotify.Watcher, bundleRootDir string) {

	index := createIndex(bundleRootDir)


	handler := Adapt(&NoopHandler{},
//		LogEvent(),
		UpdateWatcher(), 
//		UpdateIndexForNewDir(), 
//		UpdateIndexForRemovedDir(), 
		UpdateIndexForModifiedDir(), )

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				//log.Println("event:", event)
				index = handler.ServeWatcherEvent(watcher, &event, index)
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	filepath.Walk(bundleRootDir, watcherWalkTreeFunction(watcher))
}
