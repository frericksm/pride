// Package bundle provides a schema and resolver for bundle remote bundle management.

package bundle

import (
	"context"
//	"log"
	"fmt"
	"strings"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	//graphql "github.com/neelance/graphql-go"

	pcontext "github.com/frericksm/pride/context"	
	"github.com/frericksm/pride/utils"	
)


var Schema = `
	schema {
		query: Query
		mutation: Mutation
	}
	# The query type, represents all of the entry points into our object graph
	type Query {
                # Queries a single file from a bundle
                filenode(bundle_name: String!, path: String!): FileNode
                # Queries all bundles
                all_bundles(): [Bundle]!
                # Queries a single bundle
                bundle(name: String!): Bundle
	}

	# The mutation type, represents all updates we can make to our data
	type Mutation {
                # Creates a bundle
		create_bundle(bundle_symbolic_name: String!): Bundle
	}

	# Represents a bundle
	type Bundle {
		# A name given to this bundle
                name: String!
                # The root-file of this bundle
                root: Directory!
	}

	# Represents the common attributes of file and directory
        interface FileNode {
		# The absolute file path 
                path: String!
		# The base name of the file
                name: String!
                # Flag indicating if this file is a directory
                isDir: Boolean!
        }


	# Represents a file inside a bundle
	type File implements FileNode {
		# The absolute file path 
                path: String!
		# The base name of the file
                name: String!
                # Flag indicating if this file is a directory
                isDir: Boolean!
                # URI from where to read and write the content of the file 
                resource_uri: String!
	}

	# Represents a directory inside a bundle
	type Directory implements FileNode {
		# The absolute file path 
                path: String!
		# The base name of the file
                name: String!
                # Flag indicating if this file is a directory
                isDir: Boolean!
                # list of files in this directory (if a directory)
                children: [FileNode]             
	}

`

func checkBundleName(name string) error {
	if filepath.Base(name) !=  name {
		return errors.New("A bundle name cannot be a path. Has to be a simple name")
	}
	return nil	
}
func checkPath(path string) error {
	if filepath.Clean(path) !=  path {
		return errors.New("Only clean paths are allowed. No '..', etc")
	}
	return nil
}

type bundle struct {
	BundleDir string
	Name      string
}

type Resolver struct{}

func (r *Resolver) All_bundles(ctx context.Context) []*bundleResolver {
	var l []*bundleResolver
	
	bundle_root_dir := pcontext.BundleRootDir(ctx)
	
	fileinfos, err := ioutil.ReadDir(bundle_root_dir)
	utils.Check(err)
	
	for _, file := range fileinfos {
		name := file.Name()
		path := filepath.Join(bundle_root_dir, file.Name())
		
		l = append(l, &bundleResolver{
			&bundle{
				BundleDir: path,
				Name:      name,
			}})
	}
	return l
}

func (r *Resolver) Bundle(ctx context.Context, args struct{ Name string }) *bundleResolver {
	
	bundle_root_dir := pcontext.BundleRootDir(ctx)
	path := filepath.Join(bundle_root_dir, args.Name)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil
	}
	utils.Check(err)

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

func (r *bundleResolver) Root() *directoryResolver {
	return &directoryResolver{
		&file{
			BundlePath: r.b.BundleDir,
			Name:      "",
			Path:      "/",
		},
	}
}

func createManifest(bundle_dir, Bundle_symbolic_name string) {
	err1 := os.Mkdir(filepath.Join(bundle_dir ,"/META-INF"), 0755)
	utils.Check(err1)

	f, err2 := os.Create(filepath.Join(bundle_dir ,"/META-INF/MANIFEST.MF"))
	utils.Check(err2)
	defer f.Close()

	_, err3 := f.WriteString(fmt.Sprintf("Bundle-SymbolicName: %s\n", Bundle_symbolic_name))
	utils.Check(err3)
	
	f.Sync()
}

func (r *Resolver) Create_bundle(ctx context.Context, args *struct {Bundle_symbolic_name string}) (*bundleResolver, error) {

	error := checkBundleName(args.Bundle_symbolic_name)
	if error != nil {
		return nil, error
	}
	bundle_root_dir := pcontext.BundleRootDir(ctx)	
	bundle_dir := filepath.Join(bundle_root_dir, filepath.Clean(args.Bundle_symbolic_name))

	if error := os.Mkdir(bundle_dir, 0755); os.IsExist(error) {
		return nil, errors.New(fmt.Sprintf("A bundle with name '%s' already exists" , args.Bundle_symbolic_name))
	} else {		
		utils.Check(error)
	}

	//Create META-INF/MANIFEST.MF
        createManifest(bundle_dir, args.Bundle_symbolic_name)

	
	
	new_bundle := &bundle{
		BundleDir: bundle_dir,
		Name:      args.Bundle_symbolic_name,
	}
	return &bundleResolver{new_bundle}, nil
}


type fileNode interface {
	Name() string
	Path() string
	IsDir() bool
}

type fileNodeResolver struct {
	fileNode 
}

func (r *fileNodeResolver) ToFile() (*fileResolver, bool) {
	c, ok := r.fileNode.(*fileResolver)
	return c, ok
}

func (r *fileNodeResolver) ToDirectory() (*directoryResolver, bool) {
	c, ok := r.fileNode.(*directoryResolver)
	return c, ok
}

type file struct {
	BundlePath string
        Path       string
	Name       string
}

func (r *Resolver) Filenode(ctx context.Context, args struct{ BundleName, Path string }) (*fileNodeResolver, error) {
	
	var error error
	error = checkBundleName(args.BundleName)
	if error != nil {
		return nil, error
	}

	error = checkPath(args.Path)
	if error != nil {
		return nil, error
	}

	bundle_root_dir := pcontext.BundleRootDir(ctx)
	bundle_path := filepath.Join(bundle_root_dir, args.BundleName)
	_, err := os.Stat(bundle_path)

	if os.IsNotExist(err) {
		return nil, errors.New("Unknown bundle")
	}
	utils.Check(err)

	file_path := filepath.Join(bundle_path, args.Path)
	fileinfo, err := os.Stat(file_path)

	if os.IsNotExist(err) {
		return nil, errors.New("Unknown file")
	}
	utils.Check(err)

	if fileinfo.IsDir() {
		return &fileNodeResolver{
			&directoryResolver{
				&file{
					BundlePath: bundle_path,
					Name:      filepath.Base(args.Path),
					Path:      filepath.ToSlash(args.Path),
				},},}, nil
	} else {
		return &fileNodeResolver{
			&fileResolver{
				&file{
					BundlePath: bundle_path,
					Name:      filepath.Base(args.Path),
					Path:      filepath.ToSlash(args.Path),
				},
			},}, nil
	}
	

}


type directoryResolver struct {
	f *file
}

func (r *directoryResolver) Name() string {
	return r.f.Name
}

func (r *directoryResolver) Path() string {
	return r.f.Path
}

func (r *directoryResolver) IsDir() bool {
	fileInfo, error := os.Stat(filepath.Join(r.f.BundlePath, r.f.Path))
	utils.CheckNotExists(error)
	return fileInfo.IsDir()
}


func (r *directoryResolver) Children() (*[]*fileNodeResolver, error) {
	fp := filepath.Join(r.f.BundlePath, r.f.Path)
	fileInfo, error := os.Stat(fp)
	utils.CheckNotExists(error)
	
	if !fileInfo.IsDir() {
		return nil, errors.New("Path does not exist")
	}
	
	fileinfos, err := ioutil.ReadDir(fp)
	utils.Check(err)

	l := make([]*fileNodeResolver, 0)
	for _, f := range fileinfos {
		name := f.Name()
		path := filepath.Join(r.f.Path, f.Name())

		if !strings.HasPrefix(name , ".") && name != "" {
			if f.IsDir() {
				l = append(l, &fileNodeResolver{
					&directoryResolver{
						&file{
							BundlePath: r.f.BundlePath,
							Path:      filepath.ToSlash(path),
						Name:      name,
						}}})
			} else {
				l = append(l, &fileNodeResolver{
					&fileResolver{
						&file{
							BundlePath: r.f.BundlePath,
							Path:      filepath.ToSlash(path),
						Name:      name,
						}}})
			}
			
		}
		
	}
	return &l, nil
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
	utils.CheckNotExists(error)
	return fileInfo.IsDir()
}

func (r *fileResolver) Resource_uri() string {
	fileInfo, error := os.Stat(filepath.Join(r.f.BundlePath, r.f.Path))
	utils.Check(error)
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


