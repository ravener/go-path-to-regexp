# Go path-to-regexp
Turn a url path such as `/users/:name` in to a regular expression.

This is a golang port of [github.com/pillarjs/path-to-regexp](https://github.com/pillarjs/path-to-regexp) (parts of the README is also copied from it)

## Install
```sh
$ go get github.com/ravener/go-path-to-regexp
```

## Usage
Import it
```go
import (
  "github.com/ravener/go-path-to-regexp"
)
```
The package is exposed via `pathtoregexp`

```go
pathtoregexp.PathToRegexp("/users/:name", options)
```
Returns `(*regexp.Regexp, []*Token, error)` the matched output of the regexp will contain the named parameters.

Keys is list of named parameters found, the JavaScript API passed them into a given array, here it is a return value instead.

**Passing Options:**
```go
options := pathtoregexp.NewOptions() // NewOptions returns defaults.
// Modify options as needed
options.Strict = true
pathtoregexp.PathToRegexp("/users/:name", options)
// To just use default options you can also pass nil
pathtoregexp.PathToRegexp("/users/:name", nil)
```
- **Available Options**
  - **Sensitive** When `true` the regexp will be case sensitive. (default: `false`)
  - **Strict** When `true` the regexp allows an optional trailing delimiter to match. (default: `false`)
  - **End** When `true` the regexp will match to the end of the string. (default: `true`)
  - **Start** When `true` the regexp will match from the beginning of the string. (default: `true`)
  - Advanced options (use for non-pathname strings, e.g. host names):
    - **Delimiter** The default delimiter for segments. (default: `'/'`)
    - **EndsWith** Optional character, or list of characters, to treat as "end" characters.
    - **Delimiters** List of characters to consider delimiters when parsing. (default: `'./'`)
    
```go
re, keys, _ := pathtoregexp.PathToRegexp("/foo/:bar")
// re = ^/foo/([^/]+?)(?:/)?$
// keys = [&{Name:bar Prefix:/ Delimiter:/ Optional:false Repeat:false Partial:false Pattern:[^/]+? String:false}]
```

**Please note:** The `Regexp` returned by `path-to-regexp` is intended for ordered data (e.g. pathnames, hostnames). It does not handle
arbitrary data (e.g. query strings, URL fragments, JSON, etc).

#### Token Information

* `Name` The name of the token
* `Prefix` The prefix character for the segment (`/` or `.`)
* `Delimiter` The delimiter for the segment (same as prefix or `/`)
* `Optional` Indicates the token is optional (`bool`)
* `Repeat` Indicates the token is repeated (`bool`)
* `Partial` Indicates this token is a partial path segment (`bool`)
* `Pattern` The RegExp used to match this token (`string`)
* `String` Indicates that this isn't really a token and just a string literal stored in `Name` (`bool`)

The ugly `String` property is because the JavaScript version sometimes returns literal strings as tokens but Go's arrays aren't dynamic so we need a way to differentiate them while still using the same type.

See also the README for the [JavaScript path-to-regexp](https://github.com/pillarjs/path-to-regexp) not all of the concepts is documented here currently.

## The future.
This is the path matching library JavaScript frameworks like [express](https://npmjs.com/package/express) and [koa-router](https://npmjs.com/package/koa-router) use. I'm a fan of those frameworks and I want to see more Go frameworks similar to them. So here's a starting point to port some of the JavaScript ecosystem ;)

## TODO
- `Compile` "Reverse" Path-to-Regexp API

## License
[MIT](LICENSE)
