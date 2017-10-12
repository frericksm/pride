// Package bundle provides a schema and resolver for bundle remote bundle management.
package bundle

import (
	"log"
	"fmt"
	"os"
	"io/ioutil"
	"github.com/frericksm/pride/utils"	
	"github.com/frericksm/pride/processfile"	
	"strings"
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

func process_definition_id(bundle_dir string, path string) string {
	p0, err := filepath.Rel(bundle_dir, path)
	utils.Check(err)
	p1 :=  strings.Replace(p0, "/" ,".", -1)
	if strings.Index(p1, "/") == 0 {
		return p1[1:strings.LastIndex(p1, ".process")]
	} else {
		return p1[0:strings.LastIndex(p1, ".process")]
	}
}

func file_path(bundle_dir string, process_definition_id string) string {
	return filepath.Join(bundle_dir, strings.Replace(process_definition_id, "." ,"/", -1) + ".process")
}

func uses_processes(bundle_dir string, uses_processes_map *map[string][]string) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		

		if strings.Contains(filepath.Base(path), ".process") {
			refs := make([]string, 0)
			content := processfile.FileContent(path)
			if len(content) != 0 {
				p := processfile.FromBytes(content)
				for _, act := range p.Activities {
					if act.Body.ImplementationType == "SUB_FLOW" {
						refs = append(refs, act.Body.ImplementationRefId)
					}
				}
			}
			(*uses_processes_map)[process_definition_id(bundle_dir, path)] =  refs
		}

		log.Println(fmt.Sprintf("bundle: %s, uses_processes_map: %s", bundle_dir, uses_processes_map))

		
		return nil;
	}
}

 
func createBundleIndex(bundle_dir string, bundle_name string) *BundleIndex {

	uses_processes_map := make(map[string][]string)

	filepath.Walk(bundle_dir, uses_processes(bundle_dir, &uses_processes_map))
	
	return &BundleIndex{
		bundle_name: bundle_name,
		uses_processes: uses_processes_map,
	}	

}

func createIndex(bundle_root_dir string) *Index {

	m2bi := make(map[string]*BundleIndex)

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

func UpdateIndexForNewDir(index *Index) Adapter {
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

func UpdateIndexForRemovedDir(index *Index) Adapter {
	return func(h Handler) Handler {
		return HandlerFunc(func(watcher *fsnotify.Watcher, event *fsnotify.Event) {
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				log.Println("UpdateIndexForRemovedDir: ", event.Name)
			}
			h.ServeWatcherEvent(watcher, event)  
		})
	}
}

func UpdateIndexForModifiedDir(index *Index) Adapter {
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
