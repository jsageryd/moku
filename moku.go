// Package moku provides a simple but powerful tree-based HTTP router.
package moku

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/context"
)

type mokuContextKey int

const (
	mokuPathParams mokuContextKey = iota
)

// Handler is http.Handler with added context
type Handler interface {
	ServeHTTPC(context.Context, http.ResponseWriter, *http.Request)
}

// HandlerFunc is http.HandlerFunc with added context
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// ServeHTTPC calls f(ctx, w, r) and responds not found if f is nil
func (f HandlerFunc) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if f != nil {
		f(ctx, w, r)
	} else {
		http.NotFound(w, r)
	}
}

// Mux is the router/muxer. Create an instance of Mux using New().
type Mux struct {
	sync.RWMutex
	rootNode *node

	/*
	   ConcurrentAdd (default true) can be set to false if routes will not be
	   added while the router is serving requests, for higher throughput. Setting
	   this to false will avoid taking a read lock on the routes tree on each
	   request in the assumption that its tree is not being concurrently altered.
	*/
	ConcurrentAdd bool

	/*
	   RedirectTrailingSlash (default true) controls whether or not redirection
	   occurs if a request is made to a route that matches it except for the
	   trailing slash. If true and /foo is defined, a request to /foo/ will be
	   redirected to /foo. If /foo/ is defined, a request made to /foo will be
	   redirected to /foo/. If both /foo and /foo/ are defined, no redirection
	   occurs.
	*/
	RedirectTrailingSlash bool
}

// PathParams extracts path params from given context
func PathParams(ctx context.Context) map[string]string {
	pathParams, ok := ctx.Value(mokuPathParams).(map[string]string)
	if ok {
		return pathParams
	}
	return nil
}

type node struct {
	nodes     map[string]*node
	pathParam struct {
		name string
		node *node
	}
	handler Handler
}

func newNode() *node {
	return &node{
		nodes: make(map[string]*node),
	}
}

// New creates a new Mux with default configuration.
func New() *Mux {
	return &Mux{
		rootNode: newNode(),

		ConcurrentAdd:         true,
		RedirectTrailingSlash: true,
	}
}

// Delete configures a DELETE route.
func (m *Mux) Delete(path string, handler Handler) error {
	return m.addRoute("DELETE", path, handler)
}

// DeleteFunc configures a DELETE route.
func (m *Mux) DeleteFunc(path string, handler HandlerFunc) error {
	return m.Delete(path, handler)
}

// Get configures a GET route.
func (m *Mux) Get(path string, handler Handler) error {
	return m.addRoute("GET", path, handler)
}

// GetFunc configures a GET route.
func (m *Mux) GetFunc(path string, handler HandlerFunc) error {
	return m.Get(path, handler)
}

// Head configures a HEAD route.
func (m *Mux) Head(path string, handler Handler) error {
	return m.addRoute("HEAD", path, handler)
}

// HeadFunc configures a HEAD route.
func (m *Mux) HeadFunc(path string, handler HandlerFunc) error {
	return m.Head(path, handler)
}

// Options configures an OPTIONS route.
func (m *Mux) Options(path string, handler Handler) error {
	return m.addRoute("OPTIONS", path, handler)
}

// OptionsFunc configures an OPTIONS route.
func (m *Mux) OptionsFunc(path string, handler HandlerFunc) error {
	return m.Options(path, handler)
}

// Patch configures a PATCH route.
func (m *Mux) Patch(path string, handler Handler) error {
	return m.addRoute("PATCH", path, handler)
}

// PatchFunc configures a PATCH route.
func (m *Mux) PatchFunc(path string, handler HandlerFunc) error {
	return m.Patch(path, handler)
}

// Post configures a POST route.
func (m *Mux) Post(path string, handler Handler) error {
	return m.addRoute("POST", path, handler)
}

// PostFunc configures a POST route.
func (m *Mux) PostFunc(path string, handler HandlerFunc) error {
	return m.Post(path, handler)
}

// Put configures a PUT route.
func (m *Mux) Put(path string, handler Handler) error {
	return m.addRoute("PUT", path, handler)
}

// PutFunc configures a PUT route.
func (m *Mux) PutFunc(path string, handler HandlerFunc) error {
	return m.Put(path, handler)
}

// Trace configures a TRACE route.
func (m *Mux) Trace(path string, handler Handler) error {
	return m.addRoute("TRACE", path, handler)
}

// TraceFunc configures a TRACE route.
func (m *Mux) TraceFunc(path string, handler HandlerFunc) error {
	return m.Trace(path, handler)
}

var errNoLeadingSlash = errors.New("Path does not being with leading slash")

func (m *Mux) addRoute(method string, path string, handler Handler) error {
	if m.ConcurrentAdd {
		m.Lock()
		defer m.Unlock()
	}
	if path[0] != '/' {
		return errNoLeadingSlash
	}

	currentNode, ok := m.rootNode.nodes[method]
	if !ok {
		currentNode = newNode()
		m.rootNode.nodes[method] = currentNode
	}
	err := splitString(path[1:], "/", func(part string) error {
		if len(part) > 0 && part[0] == ':' {
			if currentNode.pathParam.node == nil {
				currentNode.pathParam.name = part[1:]
				currentNode.pathParam.node = newNode()
			} else {
				if currentNode.pathParam.name != part[1:] {
					return fmt.Errorf(
						"Path param ':%s' of '%s' already defined as ':%s'",
						part,
						path,
						currentNode.pathParam.name,
					)
				}
			}
			currentNode = currentNode.pathParam.node
			return nil
		}
		t := currentNode.nodes
		child, ok := t[part]
		if !ok {
			child = newNode()
			t[part] = child
		}
		currentNode = child
		return nil
	})
	if err != nil {
		return err
	}
	currentNode.handler = handler
	return nil
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.ServeHTTPC(context.Background(), w, r)
}

// ServeHTTPC is ServeHTTP with added context
func (m *Mux) ServeHTTPC(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	pathParams := PathParams(ctx)
	if pathParams == nil {
		pathParams = make(map[string]string)
		ctx = context.WithValue(ctx, mokuPathParams, pathParams)
	}
	h, isRedirect := m.findHandler(r, pathParams)
	if h == nil {
		if isRedirect {
			var code int
			if r.Method == "GET" {
				code = http.StatusMovedPermanently
			} else {
				code = http.StatusTemporaryRedirect
			}
			http.Redirect(w, r, r.URL.String(), code)
		} else {
			http.NotFound(w, r)
		}
	} else {
		h.ServeHTTPC(ctx, w, r)
	}
}

var errDeadEnd = errors.New("Dead end")

func (m *Mux) findHandler(r *http.Request, pathParams map[string]string) (Handler, bool) {
	if m.ConcurrentAdd {
		m.RLock()
		defer m.RUnlock()
	}
	var node, lastNode *node
	var ok bool
	nextNodeCandidates := m.rootNode.nodes
	node, ok = nextNodeCandidates[r.Method]
	if ok {
		nextNodeCandidates = node.nodes
	} else {
		return nil, false
	}
	path := r.URL.Path[1:]
	err := splitString(path, "/", func(part string) error {
		lastNode = node
		node, ok = nextNodeCandidates[part]
		if ok {
			nextNodeCandidates = node.nodes
		} else if lastNode.pathParam.node != nil && part != "" {
			pathParams[lastNode.pathParam.name] = part
			node = lastNode.pathParam.node
			nextNodeCandidates = node.nodes
		} else {
			return errDeadEnd
		}
		return nil
	})
	if m.RedirectTrailingSlash && (node == nil || node.handler == nil) {
		return nil, setRedirectURL(r, node, lastNode)
	}
	if err == errDeadEnd {
		return nil, false
	}
	return node.handler, false
}

func setRedirectURL(r *http.Request, node, lastNode *node) bool {
	path := r.URL.Path
	if path[len(path)-1] == '/' {
		if lastNode != nil && lastNode.handler != nil {
			r.URL.Path = path[:len(path)-1]
			return true
		}
	} else {
		if node != nil {
			trailingNode, ok := node.nodes[""]
			if ok && trailingNode.handler != nil {
				r.URL.Path = path + "/"
				return true
			}
		}
	}
	return false
}

func splitString(s string, delimiter string, callback func(string) error) error {
	start := 0
	d := delimiter[0]
	for i := 0; i < len(s); i++ {
		if s[i] == d {
			err := callback(s[start:i])
			if err != nil {
				return err
			}
			start = i + 1
		}
	}
	return callback(s[start:])
}

// PrintRoutes prints the hierarchy of configured routes.
func (m *Mux) PrintRoutes() {
	type pathItem struct {
		name   string
		node   *node
		indent int
	}
	var item *pathItem
	var stack []*pathItem
	for name, node := range m.rootNode.nodes {
		stack = append(stack, &pathItem{name, node, 0})
	}
	if m.rootNode.pathParam.node != nil {
		name := ":" + m.rootNode.pathParam.name
		node := m.rootNode.pathParam.node
		stack = append(stack, &pathItem{name, node, 0})
	}
	for len(stack) > 0 {
		item, stack = stack[len(stack)-1], stack[:len(stack)-1]
		hasHandlerStr := "  "
		if item.node.handler != nil {
			hasHandlerStr = "* "
		}
		fmt.Printf("%s%s%s\n", hasHandlerStr, strings.Repeat("  ", item.indent), item.name)
		for name, node := range item.node.nodes {
			stack = append(stack, &pathItem{"/" + name, node, item.indent + 1})
		}
		if item.node.pathParam.node != nil {
			name := ":" + item.node.pathParam.name
			node := item.node.pathParam.node
			stack = append(stack, &pathItem{"/" + name, node, item.indent + 1})
		}
	}
}
