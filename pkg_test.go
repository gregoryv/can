package can

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

func Test_should(t *testing.T) {
	if got := should([]byte("in"), io.EOF); string(got) != "in" {
		t.Fail()
	}
}

func serve(v string, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, v)
	})
}
