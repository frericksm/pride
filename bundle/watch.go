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


func StartWatching(watcher *fsnotify.Watcher, bundleRootDir string) {
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				//	if event.Op&fsnotify.Write == fsnotify.Write {
				//		log.Println("modified file:", event.Name)
				//	}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()
	filepath.Walk(bundleRootDir, create_walkTreeFunction(watcher))
}
