package putio

import (
	"errors"
	"io"
	"time"

	"github.com/ncw/rclone/fs"
)

type File struct {
	remote    string
	createdAt time.Time
	size      int64
	id        int
	fs        *Fs
}

// String returns a description of the Object
func (f *File) String() string {
	return "a file at " + f.remote
}

// Remote returns the remote path
func (f *File) Remote() string {
	return f.remote
}

// ModTime returns the modification date of the file
// It should return a best guess if one isn't available
func (f *File) ModTime() time.Time {
	return f.createdAt
}

// Size returns the size of the file
func (f *File) Size() int64 {
	return f.size
}

// Fs returns read only access to the Fs that this object is part of
func (f *File) Fs() fs.Info {
	return f.fs
}

// Hash returns the selected checksum of the file
// If no checksum is available it returns ""
func (f *File) Hash(fs.HashType) (string, error) {
	return "", errors.New("No hashing is available")
}

// Storable says whether this object can be stored
func (f *File) Storable() bool {
	return false
}

// SetModTime sets the metadata on the object to set the modification date
func (f *File) SetModTime(time.Time) error {
	return errors.New("Not implemented")
}

// Open opens the file for read.  Call Close() on the returned io.ReadCloser
func (f *File) Open(options ...fs.OpenOption) (io.ReadCloser, error) {
	return f.fs.client.Get(f.id)
}

// Update in to the object with the modTime given of the given size
func (f *File) Update(in io.Reader, src fs.ObjectInfo) error {
	return errors.New("Not implemented")
}

// Remove this object
func (f *File) Remove() error {
	return errors.New("Not implemented")
}
