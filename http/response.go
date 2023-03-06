package http

import (
	"encoding/json"
	"net/http"
)

type Res struct {
	ErrMsg   string `json:"errmsg,omitempty"`
	ErrCode  string `json:"errcode,omitempty"`
	Response any    `json:"response,omitempty"`
}

func JSON(w http.ResponseWriter, data interface{}, codes ...int) error {
	code := http.StatusOK
	if len(codes) > 0 {
		code = codes[0]
	}
	return writeJSON(w, &Res{Response: data}, code)
}

func ErrJSON(w http.ResponseWriter, errCode, errMsg string, codes ...int) error {
	code := http.StatusInternalServerError
	if len(codes) == 0 {
		code = codes[0]
	}
	return writeJSON(w, &Res{ErrCode: errCode, ErrMsg: errMsg}, code)
}

func writeJSON(w http.ResponseWriter, res *Res, code int) error {
	w.Header().Set(ContentType, "application/json")
	w.WriteHeader(code)
	bs, err := json.Marshal(res)
	if err != nil {
		return err
	}
	_, err = w.Write(bs)
	return err
}
