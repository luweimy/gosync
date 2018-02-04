package main

import (
	"fmt"
	"log"
	"net/http"
)

func ErrorText(w http.ResponseWriter, r *http.Request, v interface{}) {
	w.WriteHeader(http.StatusBadRequest)
	log.Println(v)
	w.Write([]byte(fmt.Sprint(v)))
}

func SuccessText(w http.ResponseWriter, r *http.Request, v interface{}) {
	w.WriteHeader(http.StatusOK)
	log.Println(v)
	w.Write([]byte(fmt.Sprint(v)))
}
