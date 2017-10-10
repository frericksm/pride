// Package bundle provides a schema and resolver for bundle remote bundle management.
package bundle

import (
	"log"
	"os"
	"path/filepath"
	"github.com/fsnotify/fsnotify"
)

type BundleIndex struct {
	bundle_name string;
	//hash_of_content_2_processes map[string][]string;
	//process_2_hash_of_content map[string][]string;
	uses_processes map[string][]string;
	//process_2_used_by_processes map[string][]string;
}

type Index struct {
	bundle_root_dir string;
	bundle_name_2_bundle_index map[string]*BundleIndex
}

func uses_processes(uses_processes_map *map[string][]string) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		
		if ... {
			content := processfile.FileContent(path)
			p := processfile.FromBytes(content)
			
			
		}
		
		
		err = watcher.Add(path)
		if err != nil {
			log.Fatal(err)
			return err
		}
		
		return nil;
	}
}


func createBundleIndex(path string, name sting) *BundleIndex {

	uses_processes_map := make(map[string][]string)
	filepath.Walk(bundleRootDir, uses_processes(uses_processes_map))
	
	return &BundleIndex{
		bundle_name: name,
		uses_processes: uses_processes_map,
	}	
}

func CreateIndex(bundle_root_dir string) *Index {

	m2bi := make(map[string]BundleIndex)
	fileinfos, err := ioutil.ReadDir(bundle_root_dir)
	utils.Check(err)
	
	for _, file := range fileinfos {
		if !file.IsDir() {
			continue
		}
		name := file.Name()
		if strings.HasPrefix(name , ".") {
			continue
		}
		path := filepath.Join(bundle_root_dir, file.Name())

		m2bi[name] = createBundleIndex(path, name)
	}
	return &Index{
		bundle_root_dir: bundle_root_dir,
		bundle_name_2_bundle_index: m2bi,
	}	
}

/*
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
*/

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
