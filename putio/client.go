package putio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type client struct {
	oauthToken string
}

func (c *client) List(dirID int) (*ListFilesResponse, error) {
	log.Printf("I am listing dir with ID %d", dirID)
	u := fmt.Sprintf("https://api.put.io/v2/files/list?parent_id=%d&oauth_token=%s", dirID, c.oauthToken)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
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
	log.Printf("I am getting file with ID %d", fileID)
	redirect, err := c.getRedirectURLForFile(fileID)
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(redirect)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (c *client) getRedirectURLForFile(fileID int) (string, error) {
	u := fmt.Sprintf("https://api.put.io/v2/files/%d/download?oauth_token=%s", fileID, c.oauthToken)
	resp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("Unexpected status code from server: %d", resp.StatusCode)
	}
	return resp.Header.Get("Location"), nil
}
