package accesslog

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func ExampleLogFunc() {
	l := log.New(os.Stdout, "", 0)
	fn := func(r *http.Request) func(int) {
		return func(code int) {
			l.Printf("%s %s %d %q", r.Method, r.URL, code, r.UserAgent())
		}
	}
	r := httptest.NewRequest("GET", "/index.html?foo=bar", nil)
	r.Header.Set("User-Agent", "user agent/0.1")
	fn(r)(http.StatusOK)
	// Output:
	// GET /index.html?foo=bar 200 "user agent/0.1"
}

func TestWithLog(t *testing.T) {
	b := new(bytes.Buffer)
	l := log.New(b, "", 0)
	r := httptest.NewRequest("GET", "/index.html?foo=bar", nil)
	r.Header.Set("User-Agent", "user agent/0.1")
	var h http.Handler
	h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	h = WithLog(h, l)
	h.ServeHTTP(httptest.NewRecorder(), r)
	want := "GET /index.html?foo=bar 204 \"user agent/0.1\"\n"
	if got := b.String(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}
func TestWithLog_implicitCode(t *testing.T) {
	b := new(bytes.Buffer)
	l := log.New(b, "", 0)
	r := httptest.NewRequest("GET", "/index.html?foo=bar", nil)
	r.Header.Set("User-Agent", "user agent/0.1")
	var h http.Handler
	h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok\n"))
	})
	h = WithLog(h, l)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	if want, got := http.StatusOK, rw.Code; want != got {
		t.Fatalf("wrong code: want %d, got %d", want, got)
	}
	if ct := rw.HeaderMap.Get("Content-Type"); ct != "text/plain" {
		t.Fatalf("wrong content-type: %q", ct)
	}

	want := "GET /index.html?foo=bar 200 \"user agent/0.1\"\n"
	if got := b.String(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}
