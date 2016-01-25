package moku

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func assertStatus(t *testing.T, mux *Mux, method string, path string, status int) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	mux.ServeHTTP(w, req)
	if w.Code != status {
		t.Errorf("Expected HTTP %d, got %d", status, w.Code)
	}
}

func assertPathParams(t *testing.T, mux *Mux, method string, path string, expectedPathParams map[string]string) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	context := mux.Context(req)
	mux.ServeHTTP(w, req)
	if len(context.PathParams) != len(expectedPathParams) {
		t.Errorf("Expected %d path params, got %d", len(expectedPathParams), len(context.PathParams))
	}
	for param, expectedValue := range expectedPathParams {
		gotValue := context.PathParams[param]
		if gotValue != expectedValue {
			t.Errorf("Expected path param \"%s\" = \"%s\", got \"%s\"", param, expectedValue, gotValue)
		}
	}
}

func assertHeader(t *testing.T, mux *Mux, method string, path string, headerKey string, headerValue string) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	mux.ServeHTTP(w, req)
	gotHeaderValue := w.Header().Get(headerKey)
	if gotHeaderValue != headerValue {
		t.Errorf("Expected header %s to be \"%s\", was \"%s\"", headerKey, headerValue, gotHeaderValue)
	}
}

func assertBodyEquals(t *testing.T, mux *Mux, method string, path string, expectedBody string) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	mux.ServeHTTP(w, req)
	if w.Body.String() != expectedBody {
		t.Errorf("Expected %s to return \"%s\", got \"%s\"", path, expectedBody, w.Body.String())
	}
}

func TestGetWithoutLeadingSlash(t *testing.T) {
	mux := New()
	path := "foo"
	err := mux.Get(path, func(w http.ResponseWriter, r *http.Request) {})
	if err != errNoLeadingSlash {
		t.Errorf("Expected errNoLeadingSlash, got %s", err)
	}
	assertStatus(t, mux, "GET", path, http.StatusNotFound)
}

func TestGetWithUnknownPath(t *testing.T) {
	mux := New()
	mux.Get("/foo", func(w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "GET", "/bar", http.StatusNotFound)
}

func TestDelete(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.Delete(path, func(w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "DELETE", path, http.StatusOK)
}

func TestGet(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.Get(path, func(w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "GET", path, http.StatusOK)
}

func TestHead(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.Head(path, func(w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "HEAD", path, http.StatusOK)
}

func TestOptions(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.Options(path, func(w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "OPTIONS", path, http.StatusOK)
}

func TestPatch(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.Patch(path, func(w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "PATCH", path, http.StatusOK)
}

func TestPost(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.Post(path, func(w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "POST", path, http.StatusOK)
}

func TestPut(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.Put(path, func(w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "PUT", path, http.StatusOK)
}

func TestTrace(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.Trace(path, func(w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "TRACE", path, http.StatusOK)
}

func newMuxWithGetPaths(paths []string) *Mux {
	mux := New()
	for _, path := range paths {
		mux.Get(path, func(p string) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, p)
			}
		}(path))
	}
	return mux
}

func newMuxWithPostPaths(paths []string) *Mux {
	mux := New()
	for _, path := range paths {
		mux.Post(path, func(p string) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, p)
			}
		}(path))
	}
	return mux
}

func TestMuxStaticSimple(t *testing.T) {
	paths := []string{
		"/",
		"/foo",
		"/foo/bar",
		"/bar/",
		"/bar/foo/",
	}
	mux := newMuxWithGetPaths(paths)
	mux.RedirectTrailingSlash = false

	for _, path := range paths {
		assertStatus(t, mux, "GET", path, http.StatusOK)
		expectedBody := path
		assertBodyEquals(t, mux, "GET", path, expectedBody)
	}
}

func TestMuxStaticNotFound(t *testing.T) {
	mux := New()
	mux.RedirectTrailingSlash = false
	nilHandlerPath := "/nilhandler"
	mux.Get(nilHandlerPath, nil)
	assertStatus(t, mux, "GET", nilHandlerPath, http.StatusNotFound)
	assertStatus(t, mux, "GET", "/undefined", http.StatusNotFound)
}

func TestMuxStaticRedirectTrailingSlashGet(t *testing.T) {
	mux := newMuxWithGetPaths([]string{
		"/",
		"/foo",
		"/foo/bar",
		"/bar/",
		"/bar/foo/",
		"/a",
		"/a/:id",
		"/b/",
		"/b/:id/",
	})
	mux.Get("/nilhandler", nil)
	mux.RedirectTrailingSlash = true

	expectations := []struct {
		requestedPath        string
		expectedRedirectPath string
		expectedCode         int
	}{
		{"/", "", http.StatusOK},
		{"/foo", "", http.StatusOK},
		{"/foo/", "/foo", http.StatusMovedPermanently},
		{"/foo/bar", "", http.StatusOK},
		{"/foo/bar/", "/foo/bar", http.StatusMovedPermanently},
		{"/bar", "/bar/", http.StatusMovedPermanently},
		{"/bar/", "", http.StatusOK},
		{"/bar/foo", "/bar/foo/", http.StatusMovedPermanently},
		{"/bar/foo/", "", http.StatusOK},
		{"/a", "", http.StatusOK},
		{"/a/", "/a", http.StatusMovedPermanently},
		{"/a/5", "", http.StatusOK},
		{"/a/5/", "/a/5", http.StatusMovedPermanently},
		{"/b", "/b/", http.StatusMovedPermanently},
		{"/b/", "", http.StatusOK},
		{"/b/5", "/b/5/", http.StatusMovedPermanently},
		{"/b/5/", "", http.StatusOK},
		{"/nilhandler", "", http.StatusNotFound},
		{"/undefined", "", http.StatusNotFound},
	}
	for _, e := range expectations {
		assertStatus(t, mux, "GET", e.requestedPath, e.expectedCode)
		assertHeader(t, mux, "GET", e.requestedPath, "Location", e.expectedRedirectPath)
	}
}

func TestMuxStaticRedirectTrailingSlashPost(t *testing.T) {
	mux := newMuxWithPostPaths([]string{
		"/",
		"/foo",
		"/foo/bar",
		"/bar/",
		"/bar/foo/",
	})
	mux.Post("/nilhandler", nil)
	mux.RedirectTrailingSlash = true

	expectations := []struct {
		requestedPath        string
		expectedRedirectPath string
		expectedCode         int
	}{
		{"/", "", http.StatusOK},
		{"/foo", "", http.StatusOK},
		{"/foo/", "/foo", http.StatusTemporaryRedirect},
		{"/foo/bar", "", http.StatusOK},
		{"/foo/bar/", "/foo/bar", http.StatusTemporaryRedirect},
		{"/bar", "/bar/", http.StatusTemporaryRedirect},
		{"/bar/", "", http.StatusOK},
		{"/bar/foo", "/bar/foo/", http.StatusTemporaryRedirect},
		{"/bar/foo/", "", http.StatusOK},
		{"/nilhandler", "", http.StatusNotFound},
		{"/undefined", "", http.StatusNotFound},
	}
	for _, e := range expectations {
		assertStatus(t, mux, "POST", e.requestedPath, e.expectedCode)
		assertHeader(t, mux, "POST", e.requestedPath, "Location", e.expectedRedirectPath)
	}
}

func TestMuxPathParams(t *testing.T) {
	paths := []string{
		"/foo/:id",
		"/foo/:id/bar",
		"/foo/:id/bar/:id2",
	}

	expectations := []struct {
		requestedPath      string
		expectedPathParams map[string]string
	}{
		{"/foo/1", map[string]string{"id": "1"}},
		{"/foo/1/bar", map[string]string{"id": "1"}},
		{"/foo/1/bar/2", map[string]string{"id": "1", "id2": "2"}},
	}

	mux := newMuxWithGetPaths(paths)

	for _, e := range expectations {
		assertStatus(t, mux, "GET", e.requestedPath, http.StatusOK)
		assertPathParams(t, mux, "GET", e.requestedPath, e.expectedPathParams)
	}
}

func TestDuplicatePathParam(t *testing.T) {
	mux := New()
	mux.Get("/:foo", nil)
	err := mux.Get("/:bar", nil)
	if err == nil {
		t.Errorf("Expected path param already defined error, got nil")
	}
}

func TestSplitString(t *testing.T) {
	stringSplits := map[string][]string{
		"":           {""},
		"/":          {"", ""},
		"//":         {"", "", ""},
		"/foo":       {"", "foo"},
		"foo":        {"foo"},
		"foo/":       {"foo", ""},
		"/foo/":      {"", "foo", ""},
		"/foo/bar":   {"", "foo", "bar"},
		"/foo/bar/":  {"", "foo", "bar", ""},
		"/foo/bar//": {"", "foo", "bar", "", ""},
	}
	delimiter := "/"
	for str, expectedSplit := range stringSplits {
		gotSplit := []string{}
		splitString(str, delimiter, func(part string) error {
			gotSplit = append(gotSplit, part)
			return nil
		})

		gotStdLibSplit := strings.Split(str, delimiter)
		if !splitSlicesEqual(gotSplit, gotStdLibSplit) {
			t.Errorf("Split %q => %q, strings.Split says it should be %q", str, gotSplit, gotStdLibSplit)
		}

		if !splitSlicesEqual(gotSplit, expectedSplit) {
			t.Errorf("Split %q => %q, expected %q", str, gotSplit, expectedSplit)
		}
	}
}

func TestSplitStringWithBreak(t *testing.T) {
	gotSplit := []string{}
	err := splitString("foo/bar/baz", "/", func(part string) error {
		gotSplit = append(gotSplit, part)
		return errors.New("foo")
	})
	if err == nil {
		t.Errorf("Split did not break")
	}
	if len(gotSplit) != 1 {
		t.Errorf("Expected break after first element, got %d elements: %q", len(gotSplit), gotSplit)
	}
}

func splitSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for n := range a {
		if a[n] != b[n] {
			return false
		}
	}
	return true
}

func BenchmarkMuxStaticSimple(b *testing.B) {
	mux := New()
	mux.Get("/foo", func(w http.ResponseWriter, r *http.Request) {})
	for n := 0; n < b.N; n++ {
		r, err := http.NewRequest("GET", "/foo", nil)
		if err != nil {
			panic(err)
		}
		b.StartTimer()
		mux.ServeHTTP(nil, r)
		b.StopTimer()
	}
}
