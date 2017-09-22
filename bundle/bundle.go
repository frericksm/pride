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

// Das GraphqQL-Schema
var Schema = `
	schema {
		query: Query
		mutation: Mutation
	}
	# The query type, represents all of the entry points into our object graph
	type Query {
                # Queries a single file from a bundle
                filenode(bundle_symbolic_name: String!, path: String!): FileNode
                # Queries all bundles
                all_bundles(): [Bundle]!
                # Queries a single bundle
                bundle(bundle_symbolic_name: String!): Bundle
	}

	# The mutation type, represents all updates we can make to our data
	type Mutation {
                # Create bundle
		create_bundle(bundle_symbolic_name: String!): Bundle

                # Create file
		create_file(bundle_symbolic_name: String!, path: String!, name: String!): File

                # Create dir
		create_dir(bundle_symbolic_name: String!, path: String!, name: String!): Directory

                # Delete file
		delete_bundle(bundle_symbolic_name: String!): Boolean!

                # Delete file
		delete_file(bundle_symbolic_name: String!, path: String!): Boolean!

                # Delete dir
		delete_dir(bundle_symbolic_name: String!, path: String!): Boolean!

                # Move filenode
		move_filenode(bundle_symbolic_name: String!, source: String!, target: String!): Boolean!
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
                # Last modified as seconds since unix time epoch
                lastModified: Int!
        }


	# Represents a file inside a bundle
	type File implements FileNode {
		# The absolute file path 
                path: String!
		# The base name of the file
                name: String!
                # Flag indicating if this file is a directory
                isDir: Boolean!
                # Last modified as seconds since unix time epoch
                lastModified: Int!
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
                # Last modified as seconds since unix time epoch
                lastModified: Int!
                # list of files in this directory (if a directory)
                children: [FileNode]             
	}

`

func checkBundleName(name string) error {
	if len([]rune(name)) == 0 {
		return errors.New("A bundle name cannot be empty")
	} else if filepath.Base(name) !=  name {
		return errors.New("A bundle name cannot be a path. Has to be a simple name")
	}
	return nil	
}

func checkHidden(name string) error {

	if strings.HasPrefix(name, ".") {
		return errors.New("Hidden path segments are not allowed")
	}
	return nil
}


func checkPath(path string) error {
	slashPathCleaned := filepath.ToSlash(filepath.Clean(path))
	slashPath := filepath.ToSlash(path)
	if slashPathCleaned != slashPath  {
		return errors.New("Only clean paths are allowed. No '..', no ending /, etc")
	}

	for _, seg := range strings.Split(slashPath, "/") {
		error2 := checkHidden(seg)
		if error2 != nil {
			return error2
		}
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
		if !file.IsDir() {
			continue
		}
		name := file.Name()
		if strings.HasPrefix(name , ".") {
			continue
		}
		path := filepath.Join(bundle_root_dir, file.Name())
		
		l = append(l, &bundleResolver{
			&bundle{
				BundleDir: path,
				Name:      name,
			}})
	}
	return l
}

func (r *Resolver) Bundle(ctx context.Context, args struct{ BundleSymbolicName string }) (*bundleResolver, error) {
	
	bundle_root_dir := pcontext.BundleRootDir(ctx)

	error2 := checkHidden(args.BundleSymbolicName)
	if error2 != nil {
		return nil, error2
	}

	path := filepath.Join(bundle_root_dir, args.BundleSymbolicName)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}
	utils.Check(err)

	return &bundleResolver{
		&bundle{
			BundleDir: path,
			Name:      args.BundleSymbolicName,
		}, 
	}, nil
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

	_, err3 := f.WriteString(fmt.Sprintf("Bundle-SymbolicName: %s", Bundle_symbolic_name))
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
		return nil, errors.New(fmt.Sprintf("Bundle '%s' already exists" , args.Bundle_symbolic_name))
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

func (r *Resolver) Delete_bundle(ctx context.Context, args *struct {Bundle_symbolic_name string}) (bool, error) {

	error1 := checkBundleName(args.Bundle_symbolic_name)
	if error1 != nil {
		return false, error1
	}

	bundle_root_dir := pcontext.BundleRootDir(ctx)	
	bundle_dir := filepath.Join(bundle_root_dir, filepath.Clean(args.Bundle_symbolic_name))

	if _, error := os.Stat(bundle_dir); os.IsNotExist(error) {
		return false, errors.New(fmt.Sprintf("Bundle '%s' does not exist" , args.Bundle_symbolic_name))
	} 

	if error := os.RemoveAll(bundle_dir); error != nil {
		return false, errors.New(fmt.Sprintf("Bundle '%s' cannot be deleted" , args.Bundle_symbolic_name))
	} 

	return true, nil
}


func (r *Resolver) Create_file(ctx context.Context, args *struct {Bundle_symbolic_name string; Path string; Name string}) (*fileResolver, error) {

	error1 := checkBundleName(args.Bundle_symbolic_name)
	if error1 != nil {
		return nil, error1
	}

	error2 := checkPath(args.Path)
	if error2 != nil {
		return nil, error2
	}

	error3 := checkPath(args.Name)
	if error3 != nil {
		return nil, error3
	}

	bundle_root_dir := pcontext.BundleRootDir(ctx)	
	bundle_dir := filepath.Join(bundle_root_dir, filepath.Clean(args.Bundle_symbolic_name))
	rel_file_path := filepath.ToSlash(filepath.Join(args.Path, args.Name))
	filepath := filepath.Join(bundle_dir , rel_file_path)

	if _, error := os.Stat(filepath); error == nil {
		return nil, errors.New(fmt.Sprintf("A file '%s' already exists" , args.Name))
	} 

	f, error := os.Create(filepath)
	if error != nil {
		return nil, errors.New(fmt.Sprintf("File '%s' cannot be created" , args.Name))
	} 

	defer f.Close()

	new_file := &file{
		BundlePath: bundle_dir,
		Path:       rel_file_path,
		Name:       args.Name,
	}
	return &fileResolver{new_file}, nil
}

func (r *Resolver) Delete_file(ctx context.Context, args *struct {Bundle_symbolic_name string; Path string}) (bool, error) {

	error1 := checkBundleName(args.Bundle_symbolic_name)
	if error1 != nil {
		return false, error1
	}

	error2 := checkPath(args.Path)
	if error2 != nil {
		return false, error2
	}

	bundle_root_dir := pcontext.BundleRootDir(ctx)	
	bundle_dir := filepath.Join(bundle_root_dir, filepath.Clean(args.Bundle_symbolic_name))
	filepath := filepath.Join(bundle_dir , args.Path)

	if _, error := os.Stat(filepath); os.IsNotExist(error) {
		return false, errors.New(fmt.Sprintf("File '%s' does not exist" , args.Path))
	} 

	if error := os.Remove(filepath); error != nil {
		return false, errors.New(fmt.Sprintf("File '%s' cannot be deleted" , args.Path))
	} 

	return true, nil
}

func (r *Resolver) Create_dir(ctx context.Context, args *struct {Bundle_symbolic_name string; Path string; Name string}) (*directoryResolver, error) {

	error1 := checkBundleName(args.Bundle_symbolic_name)
	if error1 != nil {
		return nil, error1
	}

	error2 := checkPath(args.Path)
	if error2 != nil {
		return nil, error2
	}

	error3 := checkPath(args.Name)
	if error3 != nil {
		return nil, error3
	}

	bundle_root_dir := pcontext.BundleRootDir(ctx)	
	bundle_dir := filepath.Join(bundle_root_dir, filepath.Clean(args.Bundle_symbolic_name))
	rel_file_path := filepath.ToSlash(filepath.Join(args.Path, args.Name))
	filepath := filepath.Join(bundle_dir , rel_file_path)

	if error := os.Mkdir(filepath, 0755); os.IsExist(error) {
		return nil, errors.New(fmt.Sprintf("Directory '%s' already exists" , args.Name))
	} 

	new_dir := &file{
		BundlePath: bundle_dir,
		Path:       rel_file_path,
		Name:       args.Name,
	}
	return &directoryResolver{new_dir}, nil
}

func (r *Resolver) Delete_dir(ctx context.Context, args *struct {Bundle_symbolic_name string; Path string}) (bool, error) {

	error1 := checkBundleName(args.Bundle_symbolic_name)
	if error1 != nil {
		return false, error1
	}

	error2 := checkPath(args.Path)
	if error2 != nil {
		return false, error2
	}

	bundle_root_dir := pcontext.BundleRootDir(ctx)	
	bundle_dir := filepath.Join(bundle_root_dir, filepath.Clean(args.Bundle_symbolic_name))
	filepath := filepath.Join(bundle_dir , args.Path)

	if error := os.RemoveAll(filepath); error != nil {
		return false, errors.New(fmt.Sprintf("Directory '%s' cannot be deleted" , args.Path))
	} 

	return true, nil
}

func (r *Resolver) Move_filenode(ctx context.Context, args *struct {Bundle_symbolic_name string; Source string; Target string}) (bool, error) {

	error1 := checkBundleName(args.Bundle_symbolic_name)
	if error1 != nil {
		return false, error1
	}

	error2 := checkPath(args.Source)
	if error2 != nil {
		return false, error2
	}

	error3 := checkPath(args.Target)
	if error3 != nil {
		return false, error3
	}

	bundle_root_dir := pcontext.BundleRootDir(ctx)	
	bundle_dir := filepath.Join(bundle_root_dir, filepath.Clean(args.Bundle_symbolic_name))
	
	oldpath := filepath.Join(bundle_dir , args.Source)
	newpath := filepath.Join(bundle_dir , args.Target, filepath.Base(args.Source))

	if error := os.Rename(oldpath, newpath); error != nil {
		return false, error
	} 

	return true, nil
}

type fileNode interface {
	Name() string
	Path() string
	IsDir() bool
	LastModified() int32
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

func (r *Resolver) Filenode(ctx context.Context, args struct{ BundleSymbolicName, Path string }) (*fileNodeResolver, error) {
	
	var error error
	error = checkBundleName(args.BundleSymbolicName)
	if error != nil {
		return nil, error
	}

	error = checkPath(args.Path)
	if error != nil {
		return nil, error
	}

	bundle_root_dir := pcontext.BundleRootDir(ctx)
	bundle_path := filepath.Join(bundle_root_dir, args.BundleSymbolicName)
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

func (r *directoryResolver) LastModified() int32 {
	fileInfo, error := os.Stat(filepath.Join(r.f.BundlePath, r.f.Path))
	utils.CheckNotExists(error)
	return int32(fileInfo.ModTime().Unix())
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

func (r *fileResolver) LastModified() int32 {
	fileInfo, error := os.Stat(filepath.Join(r.f.BundlePath, r.f.Path))
	utils.CheckNotExists(error)
	return int32(fileInfo.ModTime().Unix())
}


func (r *fileResolver) Resource_uri() string {
	fileInfo, error := os.Stat(filepath.Join(r.f.BundlePath, r.f.Path))
	utils.Check(error)
	if !fileInfo.IsDir() {
	  return filepath.ToSlash(filepath.Join("/bundles" , filepath.Base(r.f.BundlePath), "resources" , r.f.Path))
	}
	return ""
}
