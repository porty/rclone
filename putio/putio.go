package putio

import (
	"errors"
	"io"
	"time"

	"github.com/ncw/rclone/fs"
)

func init() {
	fs.Register(&fs.RegInfo{
		Name:        "put.io",
		Description: "put.io",
		NewFs:       NewFs,
		Options: []fs.Option{{
			Name: "putio_oauth",
			Help: "OAuth secret for put.io",
		}},
	})
}

type Fs struct {
	name string
	root string
}

// Name of the remote (as passed into NewFs)
func (f *Fs) Name() string {
	return f.name
}

// Root of the remote (as passed into NewFs)
func (f *Fs) Root() string {
	return f.root
}

// String returns a description of the FS
func (f *Fs) String() string {
	return "put.io lol"
}

// Precision of the ModTimes in this Fs
func (f *Fs) Precision() time.Duration {
	return 1 * time.Second
}

// Returns the supported hash types of the filesystem
func (f *Fs) Hashes() fs.HashSet {
	return fs.NewHashSet(fs.HashNone)
}

// List the objects and directories of the Fs starting from dir
//
// dir should be "" to start from the root, and should not
// have trailing slashes.
//
// This should return ErrDirNotFound (using out.SetError())
// if the directory isn't found.
//
// Fses must support recursion levels of fs.MaxLevel and 1.
// They may return ErrorLevelNotSupported otherwise.
func (f *Fs) List(out fs.ListOpts, dir string) {

}

// NewObject finds the Object at remote.  If it can't be found
// it returns the error ErrorObjectNotFound.
func (f *Fs) NewObject(remote string) (fs.Object, error) {
	return nil, errors.New("not implemented")
}

// Put in to the remote path with the modTime given of the given size
//
// May create the object even if it returns an error - if so
// will return the object and the error, otherwise will return
// nil and the error
func (f *Fs) Put(in io.Reader, src fs.ObjectInfo) (fs.Object, error) {
	return nil, errors.New("not implemented")
}

// Mkdir makes the directory (container, bucket)
//
// Shouldn't return an error if it already exists
func (f *Fs) Mkdir(dir string) error {
	return errors.New("not implemented")
}

// Rmdir removes the directory (container, bucket) if empty
//
// Return an error if it doesn't exist or isn't empty
func (f *Fs) Rmdir(dir string) error {
	return errors.New("not implemented")
}

// NewFs constructs an Fs from something
func NewFs(name, root string) (fs.Fs, error) {
	fs := &Fs{}
	return fs, nil
}