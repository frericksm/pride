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

	"github.com/frericksm/pride/utils"
	pcontext "github.com/frericksm/pride/context"	
)


type Handler struct {}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//fmt.Printf("ServeHTTP: %s" , r.RequestURI)
	slashed_path := filepath.ToSlash(r.RequestURI)
	re , _ := regexp.Compile("bundles/(.+?)/resources/(.*)")
	groups := re.FindStringSubmatch(slashed_path)
	bundle_name := groups[1]
	path := groups[2]

	bundle_root_dir := pcontext.BundleRootDir(r.Context())

	filename :=filepath.Join(bundle_root_dir, bundle_name, path)

	if  r.Method == http.MethodGet {
		content, err := ioutil.ReadFile(filename)
		utils.Check(err)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(content)
	} else if  r.Method == http.MethodPut {

		f, err := os.Create(filename)
		utils.Check(err)
		defer f.Close()

		//fw := bufio.NewWriter(f)
		_, err = io.Copy(f, r.Body)
		utils.Check(err)
		f.Sync()
		//fw.Flush()

	} else if  r.Method == http.MethodPost {

		f, err := os.Create(filename)
		utils.Check(err)
		defer f.Close()

		//fw := bufio.NewWriter(f)
		_, err = io.Copy(f, r.Body)
		utils.Check(err)
		f.Sync()
		//fw.Flush()

	}

}
