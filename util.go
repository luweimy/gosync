package main

import "net/http"

func GetMultipartFormValue(r *http.Request, key string) string {
	if r.MultipartForm != nil {
		values := r.MultipartForm.Value[key]
		if len(values) > 0 {
			return values[0]
		}
	}
	return ""
}
