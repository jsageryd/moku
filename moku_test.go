package moku

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/context"
)

func assertStatus(t *testing.T, mux *Mux, method string, path string, status int) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	mux.ServeHTTP(w, req)
	if w.Code != status {
		t.Errorf("Expected HTTP %d, got %d", status, w.Code)
	}
}

func assertPathParams(t *testing.T, mux *Mux, method string, definedPath string, requestPath string, expectedPathParams map[string]string) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, requestPath, nil)
	mux.GetFunc(definedPath, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		pathParams := PathParams(ctx)
		if len(pathParams) != len(expectedPathParams) {
			t.Errorf("Expected %d path params, got %d for %s", len(expectedPathParams), len(pathParams), requestPath)
		}
		for param, expectedValue := range expectedPathParams {
			gotValue := pathParams[param]
			if gotValue != expectedValue {
				t.Errorf("Expected path param \"%s\" = \"%s\", got \"%s\" for %s", param, expectedValue, gotValue, requestPath)
			}
		}
	})
	mux.ServeHTTP(w, req)
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

func TestGetFuncWithoutLeadingSlash(t *testing.T) {
	mux := New()
	path := "foo"
	err := mux.GetFunc(path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})
	if err != errNoLeadingSlash {
		t.Errorf("Expected errNoLeadingSlash, got %s", err)
	}
	assertStatus(t, mux, "GET", path, http.StatusNotFound)
}

func TestGetFuncWithUnknownPath(t *testing.T) {
	mux := New()
	mux.GetFunc("/foo", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "GET", "/bar", http.StatusNotFound)
}

func TestDeleteFunc(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.DeleteFunc(path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "DELETE", path, http.StatusOK)
}

func TestGetFunc(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.GetFunc(path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "GET", path, http.StatusOK)
}

func TestHeadFunc(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.HeadFunc(path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "HEAD", path, http.StatusOK)
}

func TestOptionsFunc(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.OptionsFunc(path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "OPTIONS", path, http.StatusOK)
}

func TestPatchFunc(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.PatchFunc(path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "PATCH", path, http.StatusOK)
}

func TestPostFunc(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.PostFunc(path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "POST", path, http.StatusOK)
}

func TestPutFunc(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.PutFunc(path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "PUT", path, http.StatusOK)
}

func TestTraceFunc(t *testing.T) {
	mux := New()
	path := "/foo"
	mux.TraceFunc(path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})
	assertStatus(t, mux, "TRACE", path, http.StatusOK)
}

func TestDelete(t *testing.T) {
	mux := New()
	path := "/foo"
	var handler HandlerFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
	mux.Delete(path, handler)
	assertStatus(t, mux, "DELETE", path, http.StatusOK)
}

func TestGet(t *testing.T) {
	mux := New()
	path := "/foo"
	var handler HandlerFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
	mux.Get(path, handler)
	assertStatus(t, mux, "GET", path, http.StatusOK)
}

func TestHead(t *testing.T) {
	mux := New()
	path := "/foo"
	var handler HandlerFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
	mux.Head(path, handler)
	assertStatus(t, mux, "HEAD", path, http.StatusOK)
}

func TestOptions(t *testing.T) {
	mux := New()
	path := "/foo"
	var handler HandlerFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
	mux.Options(path, handler)
	assertStatus(t, mux, "OPTIONS", path, http.StatusOK)
}

func TestPatch(t *testing.T) {
	mux := New()
	path := "/foo"
	var handler HandlerFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
	mux.Patch(path, handler)
	assertStatus(t, mux, "PATCH", path, http.StatusOK)
}

func TestPost(t *testing.T) {
	mux := New()
	path := "/foo"
	var handler HandlerFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
	mux.Post(path, handler)
	assertStatus(t, mux, "POST", path, http.StatusOK)
}

func TestPut(t *testing.T) {
	mux := New()
	path := "/foo"
	var handler HandlerFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
	mux.Put(path, handler)
	assertStatus(t, mux, "PUT", path, http.StatusOK)
}

func TestTrace(t *testing.T) {
	mux := New()
	path := "/foo"
	var handler HandlerFunc = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
	mux.Trace(path, handler)
	assertStatus(t, mux, "TRACE", path, http.StatusOK)
}

func newMuxWithGetPaths(paths []string) *Mux {
	mux := New()
	for _, path := range paths {
		mux.GetFunc(path, func(p string) HandlerFunc {
			return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, p)
			}
		}(path))
	}
	return mux
}

func newMuxWithPostPaths(paths []string) *Mux {
	mux := New()
	for _, path := range paths {
		mux.PostFunc(path, func(p string) HandlerFunc {
			return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
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
	mux.GetFunc(nilHandlerPath, nil)
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
	mux.GetFunc("/nilhandler", nil)
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
	mux.PostFunc("/nilhandler", nil)
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
	expectations := []struct {
		definedPath        string
		requestedPath      string
		expectedPathParams map[string]string
	}{
		{"/foo/:id", "/foo/1", map[string]string{"id": "1"}},
		{"/foo/:id/bar", "/foo/1/bar", map[string]string{"id": "1"}},
		{"/foo/:id/bar/:id2", "/foo/1/bar/2", map[string]string{"id": "1", "id2": "2"}},
	}

	for _, e := range expectations {
		assertPathParams(t, New(), "GET", e.definedPath, e.requestedPath, e.expectedPathParams)
	}
}

func TestDuplicatePathParam(t *testing.T) {
	mux := New()
	mux.GetFunc("/:foo", nil)
	err := mux.GetFunc("/:bar", nil)
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
	mux.GetFunc("/foo", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})
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
