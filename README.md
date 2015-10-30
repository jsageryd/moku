# Moku
Moku is a Go HTTP router focused on being simple and fast.

## Features
Moku is simple by design and the only features that makes it differ from Go's
built-in router are:

- Automatic redirect of `/foo/` to `/foo` or vice versa depending on what is
  defined. If one is defined but not the other, the one not defined will
  redirect to the one defined. If both are defined, no redirection happens.

- Path parameters such as `/foo/:id`, mapping `:id` to whatever is in its place
  in the request URL.

- Zero allocation serving static routes

- Configurable concurrency -- if routes will not be modified while the router is
  serving requests, mutex locking can be turned off for increased performance.

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
