package putio

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type DoerFunc func(req *http.Request) (*http.Response, error)

func (d DoerFunc) Do(req *http.Request) (*http.Response, error) {
	return d(req)
}

func newResponse(code int, body string) *http.Response {
	return &http.Response{
		StatusCode: code,
		Body:       ioutil.NopCloser(strings.NewReader(body)),
		Header:     http.Header{},
	}
}

func TestEmptyList(t *testing.T) {
	c := newClient("oauth")
	c.client = DoerFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://api.put.io/v2/files/list?parent_id=0&oauth_token=oauth", req.URL.String())
		return newResponse(http.StatusOK, `{"status": "OK", "files": []}`), nil
	})

	resp, err := c.List(0)

	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 0, len(resp.Files))
}

func TestList(t *testing.T) {
	c := newClient("oauth")
	c.client = DoerFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "https://api.put.io/v2/files/list?parent_id=0&oauth_token=oauth", req.URL.String())
		files := `{"id": 123, "name": "fred.txt", "size": 1234}`
		return newResponse(http.StatusOK, `{"status": "OK", "files": [`+files+`]}`), nil
	})

	resp, err := c.List(0)

	require.Nil(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, len(resp.Files))
	require.Equal(t, 123, resp.Files[0].ID)
	require.Equal(t, "fred.txt", resp.Files[0].Name)
	require.Equal(t, int64(1234), resp.Files[0].Size)

}

func TestGet(t *testing.T) {
	reqCount := 0
	c := newClient("oauth")
	c.client = DoerFunc(func(req *http.Request) (*http.Response, error) {
		reqCount++
		switch reqCount {
		case 1:
			require.Equal(t, "https://api.put.io/v2/files/1/download?oauth_token=oauth", req.URL.String())
			resp := newResponse(http.StatusFound, "redirect text")
			resp.Header.Add("Location", "https://some-cdn-server/some-file.txt")
			return resp, nil
		case 2:
			require.Equal(t, "https://some-cdn-server/some-file.txt", req.URL.String())
			return newResponse(http.StatusOK, "I am a file"), nil
		}
		require.FailNow(t, "Should not get more than 2 requests")
		return nil, nil
	})

	resp, err := c.Get(1)
	require.Equal(t, 2, reqCount)
	require.Nil(t, err)
	require.NotNil(t, resp)
	body, err := ioutil.ReadAll(resp)
	if err != nil {
		panic(err)
	}
	require.Equal(t, "I am a file", string(body))
}

func TestGetObjectID(t *testing.T) {
	reqCount := 0
	c := newClient("oauth")
	c.client = DoerFunc(func(req *http.Request) (*http.Response, error) {
		reqCount++
		switch reqCount {
		case 1:
			// list the root directory
			require.Equal(t, "https://api.put.io/v2/files/list?parent_id=0&oauth_token=oauth", req.URL.String())

			resp := ListFilesResponse{
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
						Name:        "one",
					},
				},
			}
			body, err := json.Marshal(resp)
			if err != nil {
				panic(err)
			}

			return newResponse(http.StatusOK, string(body)), nil
		case 2:
			// list the "one" directory
			require.Equal(t, "https://api.put.io/v2/files/list?parent_id=100&oauth_token=oauth", req.URL.String())

			resp := ListFilesResponse{
				Status: "OK",
				Files: []FileObject{
					FileObject{
						ContentType: "text/plain",
						ID:          101,
						Name:        "readme.txt",
						Size:        1024,
					},
					FileObject{
						ContentType: "application/x-directory",
						ID:          200,
						Name:        "two",
					},
				},
			}
			body, err := json.Marshal(resp)
			if err != nil {
				panic(err)
			}

			return newResponse(http.StatusOK, string(body)), nil

		case 3:
			// list the "two" directory
			require.Equal(t, "https://api.put.io/v2/files/list?parent_id=200&oauth_token=oauth", req.URL.String())

			resp := ListFilesResponse{
				Status: "OK",
				Files: []FileObject{
					FileObject{
						ContentType: "text/plain",
						ID:          201,
						Name:        "target.txt",
						Size:        1024,
					},
					FileObject{
						ContentType: "application/x-directory",
						ID:          300,
						Name:        "three",
					},
				},
			}
			body, err := json.Marshal(resp)
			if err != nil {
				panic(err)
			}

			return newResponse(http.StatusOK, string(body)), nil
		}
		require.FailNow(t, "Should not get more than 3 requests")
		return nil, nil
	})

	id, err := c.GetObjectID("one/two/target.txt")
	require.Nil(t, err)
	require.Equal(t, 201, id)
	require.Equal(t, 3, reqCount)
}
