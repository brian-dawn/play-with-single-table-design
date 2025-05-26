package web

import (
	"bytes"
	"net/http"

	"github.com/Joker/hpp"
)

type prettyPrintResponseWrapper struct {
	buf        *bytes.Buffer
	httpWriter http.ResponseWriter
}

func (wrapper prettyPrintResponseWrapper) Header() http.Header {
	return wrapper.httpWriter.Header()
}
func (wrapper prettyPrintResponseWrapper) Write(bytes []byte) (int, error) {
	return wrapper.buf.Write(bytes)
}
func (wrapper prettyPrintResponseWrapper) WriteHeader(statusCode int) {
	wrapper.httpWriter.WriteHeader(statusCode)
}

func PrettyPrintHTML(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapped := prettyPrintResponseWrapper{
			buf:        new(bytes.Buffer),
			httpWriter: w,
		}
		defer func(toClose *bytes.Buffer) {
			toClose.Reset()
			toClose = nil
		}(wrapped.buf)
		next.ServeHTTP(wrapped, r)
		hpp.Format(wrapped.buf, w)
	})
}
