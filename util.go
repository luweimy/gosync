package main

import (
	"net/http"
	"time"
)

func GetMultipartFormValue(r *http.Request, key string) string {
	if r.MultipartForm != nil {
		values := r.MultipartForm.Value[key]
		if len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func NowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
