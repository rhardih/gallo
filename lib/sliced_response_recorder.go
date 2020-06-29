package lib

import (
	"net/http/httptest"
)

// SlicedResponseRecorder is a wrapper for httptest.ResponseRecorder, which
// changes Body from a bytes.Buffer to a finalised []byte, with the contents of
// the buffer.
//
// This is a utility struct for use with msgpack.(Unm/M)arshal, which doesn't
// read the contents of buffers as is.
type SlicedResponseRecorder struct {
	Body []byte
	*httptest.ResponseRecorder
}

func NewSlicedResponseRecorder(r *httptest.ResponseRecorder) *SlicedResponseRecorder {
	return &SlicedResponseRecorder{r.Body.Bytes(), r}
}
