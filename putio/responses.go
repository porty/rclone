package putio

import (
	"strings"
	"time"
)

type jsonTime time.Time

func (j *jsonTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	t, err := time.Parse("2006-01-02T15:04:05", s)
	if err != nil {
		return err
	}
	*j = jsonTime(t)
	return nil
}

func (j jsonTime) String() string {
	return time.Time(j).String()
}

type FileObject struct {
	ContentType       string      `json:"content_type"`
	Crc32             string      `json:"crc32"`
	CreatedAt         jsonTime    `json:"created_at"`
	FirstAccessedAt   interface{} `json:"first_accessed_at"`
	Icon              string      `json:"icon"`
	ID                int         `json:"id"`
	IsMp4Available    bool        `json:"is_mp4_available"`
	IsShared          bool        `json:"is_shared"`
	Name              string      `json:"name"`
	OpensubtitlesHash interface{} `json:"opensubtitles_hash"`
	ParentID          int         `json:"parent_id"`
	Screenshot        interface{} `json:"screenshot"`
	Size              int64       `json:"size"`
}

type ListFilesResponse struct {
	Files        []FileObject `json:"files"`
	Parent       FileObject   `json:"parent"`
	Status       string       `json:"status"`
	ErrorType    string       `json:"error_type"`
	ErrorMessage string       `json:"error_message"`
}
