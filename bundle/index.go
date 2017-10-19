// Package bundle provides a schema and resolver for bundle remote bundle management.
package bundle

import (
	"log"
//	"fmt"
	"os"
	"io/ioutil"
	"github.com/frericksm/pride/utils"	
	"github.com/frericksm/pride/processfile"	
	"strings"
	"path/filepath"
	"github.com/fsnotify/fsnotify"
	"crypto/sha256"
)

var e struct{}

type BundleIndex struct {
	bundle_dir string;
	bundle_name string;
	uses_processes *map[string]map[string]struct{};
	usedby_processes *map[string]map[string]struct{};
	path_contenthash *map[string][32]byte;
	contenthash_path *map[[32]byte]map[string]struct{}
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

func walk_files(bundle_dir string, uses_processes_map *map[string]map[string]struct{}, path_contenthash_map *map[string][32]byte) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		
		fi , _ := os.Stat(path)
		if fi.IsDir() {
			return nil
		}
		
		content := processfile.FileContent(path)

	        //log.Println(fmt.Sprintf("walkFile: %s", filepath.Clean(path)))

		// Calc SHA256 for all files
		p0 := filepath.Clean(path)
		(*path_contenthash_map)[p0] = sha256.Sum256(content)

		// Calc refs for all process files
		if strings.Contains(filepath.Base(path), ".process") {
			refs := make(map[string]struct{})
	//		refs := make([]string, 0)
			if len(content) != 0 {
				p := processfile.FromBytes(content)
				for _, act := range p.Activities {
					if act.Body.ImplementationType == "SUB_FLOW" {
						refs[act.Body.ImplementationRefId] = e
						//refs = append(refs, act.Body.ImplementationRefId)
					}
				}
			}
			(*uses_processes_map)[process_definition_id(bundle_dir, path)] =  refs
		}

		return nil;
	}
}

func reverse_uses_processes_map(uses_processes_map *map[string]map[string]struct{}) *map[string]map[string]struct{} {

	reverse := make(map[string]map[string]struct{})

	for k,v := range *uses_processes_map {
		for p,_ := range v {
			l:= reverse[p]
			if l == nil {
				l = make(map[string]struct{})
			}
			l[k] = e
			reverse[p] = l
		}
	}
	return &reverse
}

func reverse_path_contenthash_map(path_contenthash_map *map[string][32]byte) *map[[32]byte]map[string]struct{} {

	reverse := make(map[[32]byte]map[string]struct{})

	for k,v := range *path_contenthash_map {
		l:= reverse[v]
		if l == nil {
			l = make(map[string]struct{})
		}
		l[k] = e
		reverse[v] = l
	}
	return &reverse
}
 

func createBundleIndex(bundle_dir string, bundle_name string) *BundleIndex {

	uses_processes_map := make(map[string]map[string]struct{})
	path_contenthash_map := make(map[string][32]byte)

	filepath.Walk(bundle_dir, walk_files(bundle_dir, &uses_processes_map, &path_contenthash_map))
	
	return &BundleIndex{
		bundle_dir: bundle_dir,
		bundle_name: bundle_name,
		uses_processes: &uses_processes_map,
		usedby_processes: reverse_uses_processes_map(&uses_processes_map),
		path_contenthash: &path_contenthash_map,
		contenthash_path: reverse_path_contenthash_map(&path_contenthash_map),
	}	
}

func updateBundleIndex(bundle_dir string, bundle_name string, bundle_index *BundleIndex) *BundleIndex {
	return createBundleIndex(bundle_dir, bundle_name)
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
		//log.Println(fmt.Sprintf("bundle: %s, index: %s", path, m2bi[name]))
	}

	return &Index{
		bundle_root_dir: bundle_root_dir,
		bundle_name_2_bundle_index: m2bi,
	}	
}

// Baut einen neuen Index, der nur den BundleIndex, den Pfad 'modified_dir enthält,  neu berechnet
func updateIndex(modified_dir string, index *Index) *Index {

        //log.Println(fmt.Sprintf("updateIndex: modified_dir %s", modified_dir))
	m2bi := make(map[string]*BundleIndex)

	fileinfos, err := ioutil.ReadDir(index.bundle_root_dir)
	utils.Check(err)
	
	for _, file := range fileinfos {
		if !file.IsDir() {
			continue
		}
		name := file.Name()
		if strings.HasPrefix(name , ".") {
			continue
		}
		path := filepath.Join(index.bundle_root_dir, file.Name())

		if strings.HasPrefix(filepath.Clean(modified_dir), filepath.Clean(path)) {
			//log.Println(fmt.Sprintf("updateIndex: %s", filepath.Clean(path)))
			m2bi[name] = updateBundleIndex(path, name, index.bundle_name_2_bundle_index[name])
		} else {
			m2bi[name] = index.bundle_name_2_bundle_index[name]
		}
		//log.Println(fmt.Sprintf("bundle: %s, index: %s", path, m2bi[name]))
	}

	return &Index{
		bundle_root_dir: index.bundle_root_dir,
		bundle_name_2_bundle_index: m2bi,
	}	
}

func UpdateIndexForNewDir() Adapter {
	return func(h Handler) Handler {
		return HandlerFunc(func(watcher *fsnotify.Watcher, event *fsnotify.Event, index *Index) *Index {
			if event.Op&fsnotify.Create == fsnotify.Create {
				if fi, _ := os.Stat(event.Name); fi.IsDir() {
					log.Println("UpdateIndexForNewDir: ", event.Name)
					//filepath.Walk(event.Name, create_walkTreeFunction(watcher))
				}
			}
			return h.ServeWatcherEvent(watcher, event, index)  
		})
	}
}

func UpdateIndexForRemovedDir() Adapter {
	return func(h Handler) Handler {
		return HandlerFunc(func(watcher *fsnotify.Watcher, event *fsnotify.Event, index *Index) *Index {
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				log.Println("UpdateIndexForRemovedDir: ", event.Name)
			}
			return h.ServeWatcherEvent(watcher, event, index)  
		})
	}
}

func UpdateIndexForModifiedDir() Adapter {
	return func(h Handler) Handler {
		return HandlerFunc(func(watcher *fsnotify.Watcher, event *fsnotify.Event, index *Index) *Index {
			new_index := index
			if event.Op&fsnotify.Write == fsnotify.Write {
				// fi, er := os.Stat(event.Name)
				_, er := os.Stat(event.Name)
				if os.IsNotExist(er) {
					//nothing to do
				//} else if fi.IsDir() {
				} else  {
					new_index = updateIndex(event.Name, index) 
					//log.Println("UpdateIndexForModifiedDir: ", event.Name)
					correctErrors(index, new_index)
					//updateIndex(event.Name, index)
					//filepath.Walk(event.Name, create_walkTreeFunction(watcher))
				}
			}
			return h.ServeWatcherEvent(watcher, event, new_index)  
		})
	}
}
