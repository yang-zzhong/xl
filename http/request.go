package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

const (
	ContentType = "Content-Type"
)

var (
	ErrUnknownContentType = errors.New("unknown content type")
)

func HeaderEqual(req *http.Request, header, val string) bool {
	return strings.EqualFold(req.Header.Get(header), val)
}

func GetParam(req *http.Request, p interface{}) error {
	switch {
	case HeaderEqual(req, ContentType, "application/json"):
		return GetJSON(req, p)
	}
	return ErrUnknownContentType
}

func GetJSON(req *http.Request, p interface{}) error {
	defer req.Body.Close()
	return json.NewDecoder(req.Body).Decode(p)
}
