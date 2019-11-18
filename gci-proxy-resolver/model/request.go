package model

import "net/http"

type Request struct {
	Body    []byte
	Headers http.Header
}