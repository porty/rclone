package putio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"sync"
)

type idCache struct {
	m     sync.RWMutex
	cache map[string]int
}

func newIDCache() *idCache {
	return &idCache{
		cache: map[string]int{},
	}
}

func (c *idCache) get(path string) (int, bool) {
	if path == "" {
		return 0, true
	}

	c.m.RLock()
	defer c.m.RUnlock()

	id, ok := c.cache[path]

	return id, ok
}

func (c *idCache) set(path string, id int) {
	c.m.Lock()
	defer c.m.Unlock()

	c.cache[path] = id
}

type client struct {
	oauthToken string
	client     Doer

	idCache *idCache
}

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

var ErrNotFound = errors.New("not found")

func newClient(oauthToken string) *client {
	return &client{
		oauthToken: oauthToken,
		client:     http.DefaultClient,
		idCache:    newIDCache(),
	}
}

func (c *client) List(dirID int) (*ListFilesResponse, error) {
	u := fmt.Sprintf("https://api.put.io/v2/files/list?parent_id=%d&oauth_token=%s", dirID, c.oauthToken)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var listFilesResponse ListFilesResponse
	err = json.Unmarshal(b, &listFilesResponse)
	if err != nil {
		return nil, err
	}

	if listFilesResponse.Status != "OK" {
		if listFilesResponse.ErrorType != "" && listFilesResponse.ErrorMessage != "" {
			return nil, fmt.Errorf("Receieved error response from server: %s: %s", listFilesResponse.ErrorType, listFilesResponse.ErrorMessage)
		}
		return nil, errors.New("Receieved error response from server")
	}

	return &listFilesResponse, nil
}

func (c *client) Get(fileID int) (io.ReadCloser, error) {
	redirect, err := c.getRedirectURLForFile(fileID)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest(http.MethodGet, redirect, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *client) GetObjectID(objPath string) (int, error) {
	absPath := path.Clean(strings.Trim(objPath, "/"))
	if absPath == "." {
		absPath = ""
	}

	// "" is root
	// "file.jpg" is file in root
	// "dir1/dir2/file.jpg" is file in subdirs

	if id, ok := c.idCache.get(absPath); ok {
		return id, nil
	}

	parts := strings.Split(absPath, "/")

	for i := 0; i < len(parts); i++ {
		parentDir := strings.Join(parts[:i], "/")
		parentID, ok := c.idCache.get(parentDir)
		if !ok {
			return 0, ErrNotFound
		}
		resp, err := c.List(parentID)
		if err != nil {
			return 0, err
		}
		if resp.Status != "OK" {
			return 0, errors.New("Error response from server: " + resp.Status)
		}
		for _, file := range resp.Files {
			c.idCache.set(strings.Trim(parentDir+"/"+file.Name, "/"), file.ID)
		}
		if id, ok := c.idCache.get(absPath); ok {
			return id, nil
		}
	}

	return 0, ErrNotFound
}

func (c *client) getRedirectURLForFile(fileID int) (string, error) {
	u := fmt.Sprintf("https://api.put.io/v2/files/%d/download?oauth_token=%s", fileID, c.oauthToken)
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("Expected a redirect from server but received a %d instead", resp.StatusCode)
	}
	return resp.Header.Get("Location"), nil
}
