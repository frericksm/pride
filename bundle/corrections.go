// Package bundle provides a schema and resolver for bundle remote bundle management.
package bundle

import (
	"log"
	"fmt"
//	"os"
//	"io/ioutil"
//	"github.com/frericksm/pride/utils"	
//	"github.com/frericksm/pride/processfile"	
//	"strings"
	"path/filepath"
//	"github.com/fsnotify/fsnotify"
//	"crypto/sha256"
)

func findMovedTo(old_path string, content_hash [32]byte, old_bundle_index *BundleIndex, new_bundle_index *BundleIndex) (string, bool)  {

	// Alle neuen Pfade zum content_hash ...
	for new_path := range (*new_bundle_index.contenthash_path)[content_hash] {
		// ... die nicht als alte Pfade zum content_hash existieren ...
		if _, ok := (*old_bundle_index.contenthash_path)[content_hash][new_path]; !ok {
			//... und deren Basename mit dem des old_path übereinstimmt
			if filepath.Base(old_path) ==  filepath.Base(new_path) {
				return new_path, true
			}
		}
	}

	// Danach:
	// Alle neuen Pfade zum content_hash ...
	for new_path := range (*new_bundle_index.contenthash_path)[content_hash] {
		// ... die nicht als alte Pfade zum content_hash existieren ...
		if _, ok := (*old_bundle_index.contenthash_path)[content_hash][new_path]; !ok {
			return new_path, true
		}
	}
	return "", false
}

func findMovedFrom(new_path string, content_hash [32]byte, old_bundle_index *BundleIndex, new_bundle_index *BundleIndex) (string, bool)  {

	// Alle alten Pfade zum content_hash ...
	for old_path := range (*old_bundle_index.contenthash_path)[content_hash] {
		// ... die nicht als neue Pfade zum content_hash existieren ...
		if _, ok := (*new_bundle_index.contenthash_path)[content_hash][old_path]; !ok {
			//... und deren Basename mit dem des new_path übereinstimmt
			if filepath.Base(old_path) ==  filepath.Base(new_path) {
				return old_path, true
			}
		}
	}

	// Danach:
	// Alle alten Pfade zum content_hash ...
	for old_path := range (*old_bundle_index.contenthash_path)[content_hash] {
		// ... die nicht als neue Pfade zum content_hash existieren ...
		if _, ok := (*new_bundle_index.contenthash_path)[content_hash][old_path]; !ok {
			return old_path, true
		}
	}

	return "", false
}

func correctBundleErrors(old_bundle_index *BundleIndex, new_bundle_index *BundleIndex) {

	for old_path, old_content_hash := range *old_bundle_index.path_contenthash {
		
		// Deleted or moved
		if _, present := (*new_bundle_index.path_contenthash)[old_path]; !present {
                        movedTo, moved := findMovedTo(old_path, old_content_hash, old_bundle_index, new_bundle_index)
			if moved {
				log.Println(fmt.Sprintf("correctBundleErrors: Datei verschoben von %s nach %s" , old_path, movedTo))
			} else {
				log.Println(fmt.Sprintf("correctBundleErrors: Datei gelöscht %s", old_path))
			}
		}
		
		// Maybe Changed 
		if new_content_hash, present := (*new_bundle_index.path_contenthash)[old_path]; present {
			if old_content_hash != new_content_hash {
				log.Println(fmt.Sprintf("correctBundleErrors: Datei geändert %s", old_path))
			}
		}
	}
			
	// Added or moved
	for new_path, new_content_hash := range *new_bundle_index.path_contenthash {
		if _, present := (*old_bundle_index.path_contenthash)[new_path]; !present {
			_, moved := findMovedFrom(new_path, new_content_hash, old_bundle_index, new_bundle_index)
			if !moved {
				log.Println(fmt.Sprintf("correctBundleErrors: Datei hinzugefügt %s", new_path))
			}
		}
	}
}

func correctErrors(old_index *Index, new_index *Index) {

	for name, old_bundle_index := range old_index.bundle_name_2_bundle_index {
		// Deleted bundles
		if _, present := new_index.bundle_name_2_bundle_index[name]; !present {
			log.Println(fmt.Sprintf("correctErrors: Bundle gelöscht %s", name))
		}
		
		// Maybe Changed bundle
		if new_bundle_index, present := new_index.bundle_name_2_bundle_index[name]; present {
			correctBundleErrors(old_bundle_index, new_bundle_index)
		}
	}
	
	// Added bundles
	for name, _ := range new_index.bundle_name_2_bundle_index {
		if _, present := old_index.bundle_name_2_bundle_index[name]; !present {
			log.Println(fmt.Sprintf("correctErrors: Bundle hinzugefügt %s", name))
			
		}
	}
}
