package putio

import (
	"io"
	"testing"

	"github.com/ncw/rclone/fs"
	"github.com/stretchr/testify/require"
)

type FakeClient struct {
	list        func(dirID int) (*ListFilesResponse, error)
	get         func(fileID int) (io.ReadCloser, error)
	getObjectID func(objPath string) (int, error)
}

func (c *FakeClient) List(dirID int) (*ListFilesResponse, error) {
	if c.list == nil {
		panic("list not implemented in FakeClient")
	}
	return c.list(dirID)
}

func (c *FakeClient) Get(fileID int) (io.ReadCloser, error) {
	if c.get == nil {
		panic("get not implemented in FakeClient")
	}
	return c.Get(fileID)
}

func (c *FakeClient) GetObjectID(objPath string) (int, error) {
	if c.getObjectID == nil {
		if objPath == "" || objPath == "/" {
			return 0, nil
		}
		panic("getObjectID not implemented in FakeClient")
	}
	return c.getObjectID(objPath)
}

type StubListOpts struct {
	objects  []fs.Object
	dirs     []*fs.Dir
	err      error
	finished bool
	level    int

	abort bool
}

func (s *StubListOpts) Add(obj fs.Object) bool {
	s.objects = append(s.objects, obj)
	return s.abort
}

func (s *StubListOpts) AddDir(dir *fs.Dir) bool {
	s.dirs = append(s.dirs, dir)
	return s.abort
}

func (s *StubListOpts) IncludeDirectory(remote string) bool {
	return false
}

func (s *StubListOpts) SetError(err error) {
	s.err = err
}

func (s *StubListOpts) Level() int {
	return s.level
}

func (s *StubListOpts) Buffer() int {
	return 0
}

func (s *StubListOpts) Finished() {
	s.finished = true
}

func (s *StubListOpts) IsFinished() bool {
	return s.finished
}

func TestRootList(t *testing.T) {
	f, err := newFs("putter", "/", "oauth")
	if err != nil {
		panic(err)
	}

	numRequests := 0
	fakeClient := FakeClient{
		list: func(dirID int) (*ListFilesResponse, error) {
			numRequests++
			require.Equal(t, 1, numRequests)
			require.Equal(t, 0, dirID)
			return &ListFilesResponse{
				Status: "OK",
				Files: []FileObject{
					FileObject{
						ContentType: "text/plain",
						ID:          1,
						Name:        "readme.txt",
						Size:        1024,
					},
				},
			}, nil
		},
	}
	f.client = &fakeClient

	out := StubListOpts{}

	f.List(&out, "")

	require.Equal(t, 1, numRequests)
	require.Nil(t, out.err)
	require.Empty(t, out.dirs)
	require.Equal(t, 1, len(out.objects))
	require.Equal(t, "a file at readme.txt", out.objects[0].String())
}

func TestSubdirList(t *testing.T) {
	f, err := newFs("putter", "/dir1", "oauth")
	if err != nil {
		panic(err)
	}

	numRequests := 0

	fakeClient := FakeClient{
		getObjectID: func(obj string) (int, error) {
			numRequests++
			require.Equal(t, 1, numRequests)
			require.Equal(t, "/dir1/dir2", obj)
			return 123, nil
		},
		list: func(dirID int) (*ListFilesResponse, error) {
			numRequests++
			require.Equal(t, 2, numRequests)
			require.Equal(t, 123, dirID)
			return &ListFilesResponse{
				Status: "OK",
				Files: []FileObject{
					FileObject{
						ContentType: "text/plain",
						ID:          201,
						Name:        "gorillas.txt",
						Size:        1024,
					},
				},
			}, nil
		},
	}
	f.client = &fakeClient

	out := StubListOpts{}

	f.List(&out, "dir2")

	require.Equal(t, 2, numRequests)
	require.Nil(t, out.err)
	require.Empty(t, out.dirs)
	require.Equal(t, 1, len(out.objects))
	require.Equal(t, "a file at dir2/gorillas.txt", out.objects[0].String())
	file := out.objects[0].(*File)
	require.Equal(t, 201, file.id)
}

func TestNotFoundList(t *testing.T) {
	f, err := newFs("putter", "/dir1", "oauth")
	if err != nil {
		panic(err)
	}

	numRequests := 0

	fakeClient := FakeClient{
		getObjectID: func(obj string) (int, error) {
			numRequests++
			require.Equal(t, 1, numRequests)
			require.Equal(t, "/dir1/dirRofl", obj)
			return 0, ErrNotFound
		},
	}
	f.client = &fakeClient

	out := StubListOpts{}

	f.List(&out, "dirRofl")

	require.Equal(t, 1, numRequests)
	require.Equal(t, fs.ErrorObjectNotFound, out.err)
}
