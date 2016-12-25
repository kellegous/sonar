package web

import "net/http"

// ListenAndServe ...
func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, nil)
}
