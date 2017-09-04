// Package bundle provides a schema and resolver for bundle remote bundle management.

package bundle

import (
//	"log"
	"strings"
	"io/ioutil"
	"os"
	"path/filepath"
	//graphql "github.com/neelance/graphql-go"
)

const BUNDLE_ROOT_DIR = "c:/5555/install"


func check(e error) {
    if e != nil {
        panic(e)
    }
}


var Schema = `
	schema {
		query: Query
		mutation: Mutation
	}
	# The query type, represents all of the entry points into our object graph
	type Query {
                file(bundle_name: String!, path: String!): File
                all_bundles(): [Bundle]!
                bundle(name: String!): Bundle
	}
	# The mutation type, represents all updates we can make to our data
	type Mutation {
		create_bundle(name: String!): Bundle
	}
	# Represents a bundle
	type Bundle {
		# A name given to this bundle
                name: String!
                # The root-file of this bundle
                root: File!
	}
	# Represents a file or directory of a bundle
	type File {
		# The absolute file path 
                path: String!
		# The base name of the file
                name: String!
                # Flag indicating if this file is a directory
                isDir: Boolean!
                # URI from where to read and write the contents of the file 
                resource_uri: String!
                # list of files in this directory (if a directory)
                children: [File]             
	}
`

type bundle struct {
	BundleDir string
	Name      string
}


type Resolver struct{}

func (r *Resolver) All_bundles() []*bundleResolver {
	var l []*bundleResolver

	fileinfos, err := ioutil.ReadDir(BUNDLE_ROOT_DIR)
	check(err)

	for _, file := range fileinfos {
		name := file.Name()
		path := filepath.Join(BUNDLE_ROOT_DIR, file.Name())
		
		l = append(l, &bundleResolver{
			&bundle{
				BundleDir: path,
				Name:      name,
			}})
	}
	return l
}

func (r *Resolver) Bundle(args struct{ Name string }) *bundleResolver {
	
	path := filepath.Join(BUNDLE_ROOT_DIR, args.Name)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	check(err)

	return &bundleResolver{
		&bundle{
			BundleDir: path,
			Name:      args.Name,
		},
	}
}

type bundleResolver struct {
	b *bundle
}

func (r *bundleResolver) Name() string {
	return r.b.Name
}

func (r *bundleResolver) Root() *fileResolver {
	return &fileResolver{
		&file{
			BundlePath: r.b.BundleDir,
			Name:      "",
			Path:      "/",
		},
	}
}

func (r *Resolver) Create_bundle(args *struct {Name string}) *bundleResolver {
	bundle_dir := filepath.Join(BUNDLE_ROOT_DIR, args.Name)
	error := os.Mkdir(bundle_dir, 0755)
	check(error)
	new_bundle := &bundle{
		BundleDir: bundle_dir,
		Name:      args.Name,
	}
	return &bundleResolver{new_bundle}
}


type file struct {
	BundlePath string
        Path       string
	Name       string
}


func (r *Resolver) File(args struct{ BundleName, Path string }) *fileResolver {
	
	bundle_path := filepath.Join(BUNDLE_ROOT_DIR, args.BundleName)
	_, err := os.Stat(bundle_path)
	if os.IsNotExist(err) {
		return nil
	}
	check(err)

	return &fileResolver{
		&file{
			BundlePath: bundle_path,
			Name:      filepath.Base(args.Path),
			Path:      filepath.ToSlash(args.Path),
		},
	}
}


type fileResolver struct {
	f *file
}

func (r *fileResolver) Name() string {
	return r.f.Name
}

func (r *fileResolver) Path() string {
	return r.f.Path
}

func (r *fileResolver) IsDir() bool {
	fileInfo, error := os.Stat(filepath.Join(r.f.BundlePath, r.f.Path))
	check(error)
	return fileInfo.IsDir()
}

func (r *fileResolver) Resource_uri() string {
	fileInfo, error := os.Stat(filepath.Join(r.f.BundlePath, r.f.Path))
	check(error)
	if !fileInfo.IsDir() {
	  return filepath.ToSlash(filepath.Join("/bundles" , filepath.Base(r.f.BundlePath), "resources" , r.f.Path))
	}
	return ""
}


func Filter(vs []string, f func(string) bool) []string {
    vsf := make([]string, 0)
    for _, v := range vs {
        if f(v) {
            vsf = append(vsf, v)
        }
    }
    return vsf
}

func (r *fileResolver) Children() *[]*fileResolver {
	fp := filepath.Join(r.f.BundlePath, r.f.Path)
	fileInfo, error := os.Stat(fp)
	check(error)
	
	if !fileInfo.IsDir() {
		return nil
	}
	
	fileinfos, err := ioutil.ReadDir(fp)
	check(err)

	l := make([]*fileResolver, 0)
	for _, f := range fileinfos {
		name := f.Name()
		path := filepath.Join(r.f.Path, f.Name())

		if !strings.HasPrefix(name , ".") && name != "" {
		
			l = append(l, &fileResolver{
				&file{
					BundlePath: r.f.BundlePath,
					Path:      filepath.ToSlash(path),
					Name:      name,
				}})
		}
		
	}
	return &l
}
