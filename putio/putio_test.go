package putio

import (
	"errors"
	"io"
	"testing"

	"github.com/ncw/rclone/fs"
	"github.com/stretchr/testify/require"
)

type ListerFunc func(dirID int) (*ListFilesResponse, error)

func (f ListerFunc) List(dirID int) (*ListFilesResponse, error) {
	return f(dirID)
}

func (f ListerFunc) Get(fileID int) (io.ReadCloser, int64, error) {
	return nil, 0, errors.New("not implemented")
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
	// ???
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

	clientRequests := 0

	f.client = ListerFunc(func(dirID int) (*ListFilesResponse, error) {
		clientRequests++
		switch clientRequests {
		case 1:
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

		default:
			t.Fatalf("Wasn't expecting client.List to be called %d times", clientRequests)
			return nil, errors.New("wat")
		}

	})

	out := StubListOpts{}

	f.List(&out, "")

	require.Nil(t, out.err)
	require.Empty(t, out.dirs)
	require.Equal(t, 1, len(out.objects))
	require.Equal(t, "a file called readme.txt", out.objects[0].String())
}

func TestSubdirList(t *testing.T) {
	f, err := newFs("putter", "/dir1", "oauth")
	if err != nil {
		panic(err)
	}

	clientRequests := 0

	f.client = ListerFunc(func(dirID int) (*ListFilesResponse, error) {
		clientRequests++
		switch clientRequests {
		case 1:
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
					FileObject{
						ContentType: "application/x-directory",
						ID:          100,
						Name:        "dir1",
					},
				},
			}, nil

		case 2:
			require.Equal(t, 100, dirID)
			return &ListFilesResponse{
				Status: "OK",
				Files: []FileObject{
					FileObject{
						ContentType: "application/x-directory",
						ID:          200,
						Name:        "dir2",
					},
				},
			}, nil

		case 3:
			require.Equal(t, 200, dirID)
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

		default:
			t.Fatalf("Wasn't expecting client.List to be called %d times", clientRequests)
			return nil, errors.New("wat")
		}

	})

	out := StubListOpts{}

	f.List(&out, "dir2")

	require.Equal(t, 3, clientRequests)
	require.Nil(t, out.err)
	require.Empty(t, out.dirs)
	require.Equal(t, 1, len(out.objects))
	require.Equal(t, "a file called gorillas.txt", out.objects[0].String())
	file := out.objects[0].(*File)
	require.Equal(t, 201, file.id)
}

func TestNotFoundList(t *testing.T) {
	f, err := newFs("putter", "/dir1", "oauth")
	if err != nil {
		panic(err)
	}

	clientRequests := 0

	f.client = ListerFunc(func(dirID int) (*ListFilesResponse, error) {
		clientRequests++
		switch clientRequests {
		case 1:
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
					FileObject{
						ContentType: "application/x-directory",
						ID:          100,
						Name:        "dir1",
					},
				},
			}, nil

		case 2:
			require.Equal(t, 100, dirID)
			return &ListFilesResponse{
				Status: "OK",
				Files: []FileObject{
					FileObject{
						ContentType: "application/x-directory",
						ID:          200,
						Name:        "dir2",
					},
				},
			}, nil

		default:
			t.Fatalf("Wasn't expecting client.List to be called %d times", clientRequests)
			return nil, errors.New("wat")
		}

	})

	out := StubListOpts{}

	f.List(&out, "dirRofl")

	require.Equal(t, 2, clientRequests)
	require.Equal(t, fs.ErrorObjectNotFound, out.err)
}
