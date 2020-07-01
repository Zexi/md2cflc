package confluence

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"strings"
)

type Ancestor struct {
	Id int `json:"id"`
}

type Space struct {
	Key string `json:"key"`
}

type ContentCreate struct {
	Space     `json:"space"`
	Ancestors []Ancestor `json:"ancestors"`
	Content
}

func (w *Wiki) UpdateContentCreate(content ContentCreate) (*ContentCreate, error) {
	jsonbody, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	contentEndPoint, err := w.contentEndpoint(content.Id)
	req, err := http.NewRequest("PUT", contentEndPoint.String(), strings.NewReader(string(jsonbody)))
	req.Header.Add("Content-Type", "application/json")

	res, err := w.sendRequest(req, false)
	if err != nil {
		return nil, err
	}

	var newContent ContentCreate
	err = json.Unmarshal(res, &newContent)
	if err != nil {
		return nil, err
	}

	return &newContent, nil
}

func (w *Wiki) CreateContent(content *ContentCreate, debug bool) (*ContentCreate, error) {
	jsonbody, err := json.Marshal(content)

	if err != nil {
		return nil, err
	}
	contentEndPoint, err := w.contentEndpoint("")
	req, err := http.NewRequest("POST", contentEndPoint.String(), strings.NewReader(string(jsonbody)))
	req.Header.Add("Content-Type", "application/json")
	if debug {
		Debug(httputil.DumpRequestOut(req, true))

	}
	_, err = w.sendRequest(req, debug)
	if err != nil {
		return nil, err
	}

	return content, nil
}
