// Package bundle provides a schema and resolver for bundle remote bundle management.
package bundle

import (
//	"context"
	"log"
//	"fmt"
//	"strings"
//	"errors"
//	"io/ioutil"
	"os"
	"path/filepath"
	//graphql "github.com/neelance/graphql-go"

	//pcontext "github.com/frericksm/pride/context"	
	//"github.com/frericksm/pride/utils"	
	"github.com/fsnotify/fsnotify"
)

type Index struct {
	bundle_name string;
	hash_of_content_2_processes map[string][]string;
	process_2_hash_of_content map[string][]string;
	process_2_used_processes map[string][]string;
	process_2_used_by_processes map[string][]string;
}

func create_index(bundleRootDir string) *Index {
	var index = new(Index)
	
	index.hash_of_content_2_processes = make(map[string][]string)
	index.process_2_hash_of_content = make(map[string][]string)
	index.process_2_used_processes = make(map[string][]string)
	index.process_2_used_by_processes = make(map[string][]string)
	
	return index
}

func create_walkTreeFunction(watcher *fsnotify.Watcher) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		
		err = watcher.Add(path)
		if err != nil {
			log.Fatal(err)
			return err
		}
		
		return nil;
	}
}


type Handler interface {
        ServeWatcherEvent(*fsnotify.Watcher, *fsnotify.Event)
}

type Adapter func(Handler) Handler


func Adapt(h Handler, adapters ...Adapter) Handler {
	for i := len(adapters); i > 0; i-- {
		h = adapters[i - 1](h)
	}
	return h
}

type HandlerFunc func(*fsnotify.Watcher, *fsnotify.Event)

func (f HandlerFunc) ServeWatcherEvent(watcher *fsnotify.Watcher, event *fsnotify.Event) {
  	f(watcher, event)
}

func UpdateWatcher() Adapter {
	return func(h Handler) Handler {
		return HandlerFunc(func(watcher *fsnotify.Watcher, event *fsnotify.Event) {
			if event.Op&fsnotify.Create == fsnotify.Create {
				if fi, _ := os.Stat(event.Name); fi.IsDir() {
					log.Println("UpdateWatcher: add ", event.Name)
					filepath.Walk(event.Name, create_walkTreeFunction(watcher))
				}
			} else if event.Op&fsnotify.Rename == fsnotify.Rename {
				log.Println("UpdateWatcher: remove ", event.Name)
				watcher.Remove(event.Name)
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				log.Println("UpdateWatcher: remove ", event.Name)
				watcher.Remove(event.Name)
			}
			
			
			h.ServeWatcherEvent(watcher, event)  
		})
	}
}

func UpdateIndexForNewDir() Adapter {
	return func(h Handler) Handler {
		return HandlerFunc(func(watcher *fsnotify.Watcher, event *fsnotify.Event) {
			if event.Op&fsnotify.Create == fsnotify.Create {
				if fi, _ := os.Stat(event.Name); fi.IsDir() {
					log.Println("UpdateIndexForNewDir: ", event.Name)
					//filepath.Walk(event.Name, create_walkTreeFunction(watcher))
				}
			}
			h.ServeWatcherEvent(watcher, event)  
		})
	}
}

func UpdateIndexForRemovedDir() Adapter {
	return func(h Handler) Handler {
		return HandlerFunc(func(watcher *fsnotify.Watcher, event *fsnotify.Event) {
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				log.Println("UpdateIndexForRemovedDir: ", event.Name)
			}
			h.ServeWatcherEvent(watcher, event)  
		})
	}
}

func UpdateIndexForModifiedDir() Adapter {
	return func(h Handler) Handler {
		return HandlerFunc(func(watcher *fsnotify.Watcher, event *fsnotify.Event) {
			if event.Op&fsnotify.Write == fsnotify.Write {
				fi, er := os.Stat(event.Name)
				if os.IsNotExist(er) {
					//nothing to do
				} else if fi.IsDir() {
					log.Println("UpdateIndexForModifiedDir: ", event.Name)
					//filepath.Walk(event.Name, create_walkTreeFunction(watcher))
				}
			}
			h.ServeWatcherEvent(watcher, event)  
		})
	}
}

func LogEvent() Adapter {
	return func(h Handler) Handler {
		return HandlerFunc(func(watcher *fsnotify.Watcher, event *fsnotify.Event) {
			log.Println("Event: ", event)
			h.ServeWatcherEvent(watcher, event)  
		})
	}
}


type NoopHandler struct {}

func (h *NoopHandler) ServeWatcherEvent(watcher *fsnotify.Watcher, event *fsnotify.Event) {
	
}


func StartWatching(watcher *fsnotify.Watcher, bundleRootDir string) {
	handler := Adapt(&NoopHandler{},
		// LogEvent(),
		UpdateWatcher(), 
		UpdateIndexForNewDir(), 
		UpdateIndexForRemovedDir(), 
		UpdateIndexForModifiedDir() )
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				handler.ServeWatcherEvent(watcher, &event)
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()
	filepath.Walk(bundleRootDir, create_walkTreeFunction(watcher))
}
