# Moku
Moku tries to be a simple but powerful tree-based HTTP router. Don't use it.

## Features
Moku provides:

- Automatic redirect of `/foo/` to `/foo` or vice versa depending on what is
  defined. If one is defined but not the other, the one not defined will
  redirect to the one defined. If both are defined, no redirection happens.

- Path parameters such as `/foo/:id`, mapping `:id` to whatever is in its place
  in the request URL.

- Context (net/context) passed by argument eliminating need for locking

- Zero allocation serving static routes

- Configurable concurrency -- if routes will not be modified while the router is
  serving requests, mutex locking can be turned off for increased performance.

## Usage example
```go
package main

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/jsageryd/moku"
)

func main() {
	mux := moku.New()
	mux.GetFunc("/foo/:bar", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %s\n", moku.PathParams(ctx)["bar"])
	})
	http.Handle("/", mux)
	http.ListenAndServe(":8080", nil)
}
```

```
$ curl :8080/foo/world
Hello, world
```

## Licence
Copyright (c) 2015 Johan Sageryd <j@1616.se>

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
