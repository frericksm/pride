package resource

import (
	"os"
	"net/http"
//	"fmt"
	"io"
	"io/ioutil"
//	"bufio"
//	"strings"
	"path/filepath"
	"regexp"

	"github.com/frericksm/pride/bundle"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

type Handler struct {}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//fmt.Printf("ServeHTTP: %s" , r.RequestURI)
	slashed_path := filepath.ToSlash(r.RequestURI)
	re , _ := regexp.Compile("bundles/(.+?)/resources/(.*)")
	groups := re.FindStringSubmatch(slashed_path)
	bundle_name := groups[1]
	path := groups[2]

	filename :=filepath.Join(bundle.BUNDLE_ROOT_DIR, bundle_name, path)

	if  r.Method == http.MethodGet {
		content, err := ioutil.ReadFile(filename)
		check(err)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(content)
	} else if  r.Method == http.MethodPut {

		f, err := os.Create(filename)
		check(err)
		defer f.Close()

		//fw := bufio.NewWriter(f)
		_, err = io.Copy(f, r.Body)
		check(err)
		f.Sync()
		//fw.Flush()

	} else if  r.Method == http.MethodPost {

		f, err := os.Create(filename)
		check(err)
		defer f.Close()

		//fw := bufio.NewWriter(f)
		_, err = io.Copy(f, r.Body)
		check(err)
		f.Sync()
		//fw.Flush()

	}

}
