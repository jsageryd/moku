package moku

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// Mux is the router/muxer. Create an instance of Mux using New().
type Mux struct {
	sync.RWMutex
	rootNode *node
	contexts map[*http.Request]*Context

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

// Context contains data related to a specific request.
type Context struct {
	PathParams map[string]string
}

func newContext() *Context {
	return &Context{
		PathParams: make(map[string]string),
	}
}

// Context returns the context object for the specified request.
func (m *Mux) Context(r *http.Request) *Context {
	context, ok := m.contexts[r]
	if !ok {
		context = newContext()
		m.contexts[r] = context
	}
	return context
}

type node struct {
	nodes     map[string]*node
	pathParam struct {
		name string
		node *node
	}
	handler http.HandlerFunc
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
		contexts: make(map[*http.Request]*Context),

		ConcurrentAdd:         true,
		RedirectTrailingSlash: true,
	}
}

// Delete configures a DELETE route.
func (m *Mux) Delete(path string, handler http.HandlerFunc) error {
	return m.addRoute("DELETE", path, handler)
}

// Get configures a GET route.
func (m *Mux) Get(path string, handler http.HandlerFunc) error {
	return m.addRoute("GET", path, handler)
}

// Head configures a HEAD route.
func (m *Mux) Head(path string, handler http.HandlerFunc) error {
	return m.addRoute("HEAD", path, handler)
}

// Options configures an OPTIONS route.
func (m *Mux) Options(path string, handler http.HandlerFunc) error {
	return m.addRoute("OPTIONS", path, handler)
}

// Patch configures a PATCH route.
func (m *Mux) Patch(path string, handler http.HandlerFunc) error {
	return m.addRoute("PATCH", path, handler)
}

// Post configures a POST route.
func (m *Mux) Post(path string, handler http.HandlerFunc) error {
	return m.addRoute("POST", path, handler)
}

// Put configures a PUT route.
func (m *Mux) Put(path string, handler http.HandlerFunc) error {
	return m.addRoute("PUT", path, handler)
}

// Trace configures a TRACE route.
func (m *Mux) Trace(path string, handler http.HandlerFunc) error {
	return m.addRoute("TRACE", path, handler)
}

var errNoLeadingSlash = errors.New("Path does not being with leading slash")

func (m *Mux) addRoute(method string, path string, handler http.HandlerFunc) error {
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
				currentNode.pathParam.name = part
				currentNode.pathParam.node = newNode()
			} else {
				if currentNode.pathParam.name != part {
					return fmt.Errorf(
						"Path param '%s' of '%s' already defined as '%s'",
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
	defer delete(m.contexts, r)
	h, isRedirect := m.findHandlerFunc(r)
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
		h(w, r)
	}
}

var errDeadEnd = errors.New("Dead end")

func (m *Mux) findHandlerFunc(r *http.Request) (http.HandlerFunc, bool) {
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
			m.Context(r).PathParams[lastNode.pathParam.name] = part
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
		name := m.rootNode.pathParam.name
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
			name := item.node.pathParam.name
			node := item.node.pathParam.node
			stack = append(stack, &pathItem{"/" + name, node, item.indent + 1})
		}
	}
}