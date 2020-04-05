package pathtoregexp

import (
  "regexp"
  "strings"
  "strconv"
)

var PATH_REGEXP = regexp.MustCompile(strings.Join([]string{
  // Match escaped characters that would otherwise appear in future matches.
  // This allows the user to escape special characters that won't transform.
  "(\\\\.)",
  // Match Express-style parameters and un-named parameters with a prefix
  // and optional suffixes. Matches appear as:
  //
  // ":test(\\d+)?" => ["test", "\d+", undefined, "?"]
  // "(\\d+)"  => [undefined, undefined, "\d+", undefined]
  "(?:\\:(\\w+)(?:\\(((?:\\\\.|[^\\\\()])+)\\))?|\\(((?:\\\\.|[^\\\\()])+)\\))([+*?])?",
}, "|"))

// Token represent a token parsed.
type Token struct {
  Name string
  Prefix string
  Delimiter string
  Optional bool
  Repeat bool
  Partial bool
  Pattern string
  String bool // if this is a literal string then it is stored in name, this is due to the js version using multiple types in one array and Go doesn't have unions, could use interface{} with type checks but that will be uglier than this.
}

type ParseOptions struct {
  Delimiter string
  Delimiters string
}

var defaultParseOptions = &ParseOptions{
  Delimiter: "/",
  Delimiters: "./",
}

func NewParseOptions() *ParseOptions {
  return &ParseOptions{
    Delimiter: "/",
    Delimiters: "./",
  }
}

var escapeReg = regexp.MustCompile("([.+*?=^!:${}()[\\]|\\\\])")
var groupReg = regexp.MustCompile("([=!:$/()])")

func escapeString(str string) string {
  return escapeReg.ReplaceAllString(str, "\\$1")
}

func escapeGroup(group string) string {
  return groupReg.ReplaceAllString(group, "\\$1")
}

func Parse(str string, options *ParseOptions) []*Token {
  if options == nil {
    options = NewParseOptions()
  }
  tokens := []*Token{}
  key := 0
  index := 0
  path := ""
  pathEscaped := false
  matches := PATH_REGEXP.FindAllStringSubmatch(str, -1)
  idx := PATH_REGEXP.FindAllStringSubmatchIndex(str, -1)
  for x := 0; x < len(matches); x++ {
    res := matches[x]
    m := res[0]
    escaped := res[1]
    offset := idx[x][0]
    path += str[index:offset]
    index = offset + len(m)

    // Ignore already escaped sequences.
    if escaped != "" {
      path += string(escaped[1])
      pathEscaped = true
      continue
    }

    prev := ""
    next := ""
    if index > len(str) {
      next = string(str[index])
    }
    // next := str[index]
    name := res[2]
    capture := res[3]
    group := res[4]
    modifier := res[5]
    if !pathEscaped && len(path) > 0 {
      k := len(path) - 1
      if strings.IndexRune(options.Delimiters, rune(path[k])) > -1 {
        prev = string(path[k])
        path = path[0:k]
      }
    }

    // Push the current path onto the tokens.
    if path != "" {
      tokens = append(tokens, &Token{Name: path, String: true})
      path = ""
      pathEscaped = false
    }
    partial := prev != "" && next != "" && next != prev
    repeat := modifier == "+" || modifier == "*"
    optional := modifier == "?" || modifier == "*"
    var delimiter string
    if prev != "" { // Sometimes go can get ugly compared to JavaScript :(
      delimiter = prev
    } else {
      delimiter = options.Delimiter
    }
    var pattern string
    if capture != "" {
      pattern = capture
    } else {
      pattern = group
    }
    n := name
    if n == "" { n = strconv.Itoa(key + 1) }
    key++ // Increments are statements in Go, sad.

    if len(pattern) > 0 {
      pattern = escapeGroup(pattern)
    } else {
      pattern = "[^" + escapeString(delimiter) + "]+?"
    }
    tokens = append(tokens, &Token{
      Name: n,
      Prefix: prev,
      Delimiter: delimiter,
      Optional: optional,
      Repeat: repeat,
      Partial: partial,
      Pattern: pattern,
      String: false,
    })
  }
  if len(path) > 0 || index < len(str) {
    tokens = append(tokens, &Token{
      Name: path + str[index:],
      String: true,
    })
  }
  return tokens
}

type Options struct {
  *ParseOptions
  Strict bool
  Start bool
  End bool
  EndsWith []string
}

func NewOptions() *Options {
  return &Options{
    ParseOptions: NewParseOptions(),
    Strict: false,
    Start: true,
    End: true,
    EndsWith: []string{},
  }
}

func PathToRegexp(path string, options *Options) (*regexp.Regexp, []*Token, error) {
  if options == nil {
    options = NewOptions()
  }
  tokens := Parse(path, options.ParseOptions)
  keys := []*Token{}
  delimiter := escapeString(options.Delimiter)
  endsWithArr := []string{}
  if len(options.EndsWith) > 0 {
    endsWithArr = append(endsWithArr, options.EndsWith...)
  }
  for i, x := range endsWithArr {
    endsWithArr[i] = escapeString(x)
  }
  endsWithArr = append(endsWithArr, "$")
  endsWith := strings.Join(endsWithArr, "|")
  route := ""
  if options.Start {
    route = "^"
  }
  isEndDelimited := len(tokens) == 0
  for i, token := range tokens {
    if token.String {
      route += escapeString(token.Name)
      isEndDelimited = i == (len(tokens) - 1) && strings.IndexRune(options.Delimiters, rune(token.Name[len(token.Name) - 1])) > -1
    } else {
      capture := token.Pattern
      if token.Repeat {
        capture = "(?:" + token.Pattern + ")(?:" + escapeString(token.Delimiter) + "(?:" + token.Pattern + "))*"
      }
      keys = append(keys, token)
      if token.Optional {
        if token.Partial {
          route += escapeString(token.Prefix) + "(" + capture + ")?"
        } else {
          route += "(?:" + escapeString(token.Prefix) + "(" + capture + "))?"
        }
      } else {
        route += escapeString(token.Prefix) + "(" + capture + ")"
      }
    }
  }
  if options.End {
    if !options.Strict {
      route += "(?:" + delimiter + ")?"
    }
    if endsWith == "$" {
      route += "$"
    } else {
      route += "(?=" + endsWith + ")"
    }
  } else {
    if !options.Strict {
      route += "(?:" + delimiter + "(?=" + endsWith + "))?"
    }
    if !isEndDelimited {
      route += "(?=" + delimiter + "|" + endsWith + ")"
    }
  }
  reg, err := regexp.Compile(route)
  if err != nil {
    return nil, keys, err
  }
  return reg, keys, nil
}
