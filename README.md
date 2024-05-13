<div align="center" style="margin-bottom:20px">
  <img src=".assets/banner.png" alt="restapi" />
  <h1>truly RESTful API endpoint functions</h1>
  <div class='chipstrip'>
    <a href="https://github.com/blugnu/restapi/actions/workflows/release.yml">
      <img alt="build-status" src="https://github.com/blugnu/restapi/actions/workflows/release.yml/badge.svg"/>
    </a>
    <a href="https://goreportcard.com/report/github.com/blugnu/restapi" >
        <img alt="go report" src="https://goreportcard.com/badge/github.com/blugnu/restapi"/>
    </a>
    <a>
      <img alt="go version >= 1.18" src="https://img.shields.io/github/go-mod/go-version/blugnu/restapi?style=flat-square"/>
    </a>
    <a href="https://github.com/blugnu/restapi/blob/master/LICENSE">
      <img alt="MIT License" src="https://img.shields.io/github/license/blugnu/restapi?color=%234275f5&style=flat-square"/>
    </a>
    <a href="https://coveralls.io/github/blugnu/restapi?branch=master">
      <img alt="coverage" src="https://img.shields.io/coveralls/github/blugnu/restapi?style=flat-square"/>
    </a>
    <a href="https://pkg.go.dev/github.com/blugnu/restapi">
      <img alt="docs" src="https://pkg.go.dev/badge/github.com/blugnu/restapi"/>
    </a>
  </div>
</div>

## Installation

```bash
$ go get github.com/blugnu/restapi
```

## Features

- [x] [Eliminate tedious http.ResponseWriter boilerplate](#the-solution)
- [x] [Simplifies endpoint function unit tests](#simplified-unit-tests)
- [x] [Automatic content marshalling based on request 'Accept' header; supports:](#result-response):
  - `application/json` (_default if `Accept` header is not set or is `*/*`_)
  - `application/xml`
  - `text/json`
  - `text/xml`
- [X] [Consistent error responses](#error-responses)
- [x] [Configurable error response content](#error-response-mechanism-and-customization)
- [x] [`LogError` extension point](#error-logging) (_for reporting implementation errors_)
- [x] [RFC7807 support](#rfc7807-support) (_experimental_)

## The Problem

Implementing REST API endpoints in Golang can involve a lot of boilerplate code to handle HTTP
responses correctly.  This can result in code that is harder to read and maintain and may even lead
to incorrect responses if the correct order of operations is not followed when writing response
headers.

```go
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
    type data struct {
        ID      int    `json:"id"`
        Name    string `json:"name"`
        Surame  string `json:"surname"`
    }

    // Parse request body
    var person data
    if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("invalid request body"))
        return
    }

    // Validate request body
    if person.Name == "" || person.Surname == "" {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("missing name or surname"))
        return
    }

    // Store data in database
    data.ID, err := h.db.Store(person)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(err.Error()))
        return
    }

    // Marshal data to JSON
    body, err := json.Marshal(data)
    if err != nil {
        w.Write([]byte(err.Error())) // WRONG: status code not set; will incorrectly respond with 200 OK
        return
    }

    // Write response
    w.WriteHeader(http.StatusCreated)
    w.Header().Set("Content-Type", "application/json") // WRONG: this must be applied before calling WriteHeader
    w.Write(body)
}
```

## The Solution

The `restapi` package simplifies the implementation of REST API endpoints in Golang.  The package
provides a `HandleRequest` function to simplify the handling of request bodies, and middleware that
takes care of the routine work of marshalling content and writing responses.

Combined, these allow your endpoint functions to focus on and express the concerns of your API domain.

With `restapi` the above example could be rewritten as follows:

```go
import "github.com/blugnu/restapi"

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) any {
    type data struct {
        ID       int    `json:"id"`
        Name     string `json:"name"`
        Surname  string `json:"surname"`
    }
    return restapi.HandleRequest(r, func(person *data) any {
        if data == nil {
            return restapi.BadRequest("missing request body")
        }

        // Validate request body
        if person.Name == "" || person.Surname == "" {
            return restapi.BadRequest("missing name or surname")
        }

        // Store data in database
        data.ID, err := h.db.Store(person)
        if err != nil {
            return err
        }

        return restapi.Created().WithValue(data)
    })
}
```

## Simplified Unit Tests

In addition to simplifying the implementation of endpoint functions themselves, unit tests for those
functions are also simplified.  Instead of testing indirectly by establishing a recorder and
laboriously testing the response, you can test the endpoint function directly:

### Example Unit Test (_using `net/http/httptest` package_)

```go
func TestPost(t *testing.T) {
    h := &Handler{ db: &MockDB{} }
    r := httptest.NewRequest(http.MethodPost, "/post", strings.NewReader(`{"name":"John","surname":"Doe"}`))
    w := httptest.NewRecorder()

    // Call the endpoint function directly
    h.Post(w, r)

    // Check the response
    if w.Code != http.StatusCreated {
        t.Errorf("expected: %d\ngot     : %d", http.StatusCreated, w.Code)
    }
    if w.Header().Get("Content-Type") != "application/json" {
        t.Errorf("expected: %s\ngot     : %s", "application/json", w.Header().Get("Content-Type"))
    }
    wanted := `{"id":1,"name":"John","surname":"Doe"}`
    if w.Body.String() != wanted {
        t.Errorf("expected: %s\ngot     : %s", wanted, w.Body.String())
    }
}
```

### Example Unit Test (_using `restapi` package_)

```go
import "github.com/blugnu/restapi"

func TestPost(t *testing.T) {
    // ARRANGE
    h := &Handler{ db: &MockDB{} }
    r := httptest.NewRequest(http.MethodPost, "/post", strings.NewReader(`{"name":"John","surname":"Doe"}`))

    // ACT
    result := h.Post(w, r)

    // ASSERT
    // (also illustrates tests using the `github.com/blugnu/test` package)
    test.That(t, result).Equals(&restapi.Result {
        Status: http.StatusCreated,
        ContentType: "application/json",
        Value: &Person{ ID: 1, Name: "John", Surname: "Doe" },
    })
}
```

## Response Generation: How It Works

The `restapi` package provides an http _end-ware_ `restapi.Handler()` function which
accepts a _modified_ `http.HandlerFunc` returning an `any` value:

```go
import "github.com/blugnu/restapi"

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) any {
    // Parse query parameters
    query := r.URL.Query()
    id := query.Get("id")
    if id == "" {
        return restapi.BadRequest("missing id parameter")
    }

    // Fetch data from database
    data, err := h.db.Get(id)
    if err != nil {
        return err
    }

    // Write response
    return data
}

func main() {
    http.Handle("/get", restapi.HandlerFunc(Get))
    http.ListenAndServe(":8080", nil)
}
```

> _Note: Although it functions in a similar manner, the `restapi.HandlerFunc()` function is referred
> to as '**end-ware**' rather than 'middleware' due to the additional `any` return value of a
> `restapi.EndpointFunction` signature differs
> from a http.Handler
> argument meaning that it must typically be placed at the **end** of any middleware chain_

In addition to the `HandlerFunc` endware, the `restapi` package also provides a `Handler()` endware
which accepts a `restapi.EndpointHandler` rather than a function:

```go
import "github.com/blugnu/restapi"

type GetHandler struct {
    db *Database
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) any {
    // Fetch data from database
    data, err := h.db.Get(id)
    if err != nil {
        return err
    }

    // Write response
    return data
}

func main() {
    db, err := ConnectDatabase()
    if err != nil {
        log.Fatal(err)
    }
    http.Handle("/get", restapi.Handler(GetHandler{db: db}))
    http.ListenAndServe(":8080", nil)
}
```

Whether using `HandlerFunc()` or `Handler()`, initial checks are performed on each received request
to identify and validate any `Accept` header before calling the supplied endpoint function. An
appropriate response is then constructed and written, according to the type of the value returned
by the endpoint function:

| Result Type | Response |
|-------------|----------|
| `error` | `Internal Server Error` (see: [Error Responses](#error-responses))|
| `*restapi.Error` | [Error Response](#error-responses)|
| `*restapi.Problem` | [RFC7807 Problem Details Response](#rfc7807-support)|
| `*restapi.Result` | [Result Response](#result-response)|
| `[]byte` | - Non-empty: `200 OK` response (`application/octect-stream`)<br>- Empty: `204 No Content` |
| `int` | response with the returned `int` as HTTP Status Code and no content |
| `<any other type>` | `200 OK` response with value marshalled as content |

## Result Response

For more control over the response, an endpoint function can return a `*restapi.Result` value,
obtained by calling one of the following functions:

| Function | Description |
|----------|-------------|
| `Created()` | a new `*Result` value with a `201 Created` status |
| `NoContent()` | a new `*Result` value with a `204 No Content` status |
| `OK()` | a new `*Result` value with a `200 OK` status |
| `Status()` | a new `*Result` value with a specified status code |

The `*Result` type provides methods to set additional details for the response:

| Method | Description |
|--------|-------------|
| `WithContent()` | Set the content (_and content type_) of the response |
| `WithHeader()`<br>`WithHeaders()`<br>`WithNonCanonicalHeader()` | Add canonical/non-canonical headers to the response |
| `WithValue()` | Set the value to be marshalled as the response content |

### Example Result Response (_implicit 200 OK_)

```go
import "github.com/blugnu/restapi"

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) any {
    // Parse query parameters
    query := r.URL.Query()
    id := query.Get("id")
    if id == "" {
        return restapi.BadRequest("missing id parameter")
    }

    // Fetch data from database
    data, err := h.db.Get(id)
    if err != nil {
        return err
    }

    // Will yield a 200 OK response with `data` marshalled
    // according to the request `Accept` header
    return data
}
```

### Example Result Response (_explicit 202 Accepted_)

```go
import "github.com/blugnu/restapi"

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) any {
    // Parse request body
    var data any
    if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
        return restapi.BadRequest("invalid request body")
    }

    // Store data in database asynchronously
    // (illustrates use of github.com/blugnu/ulog for context logging)
    go func () { if err = h.db.Store(data); err != nil {
        ulog.FromContext(rq.Context()).
            Error(ctx, err)
    } }()

    return restapi.Status(http.StatusAccepted)
}
```

## Error Responses

A `restapi` endpoint function can return an error response by returning an `error` or an `*restapi.Error`.

If an `error` is returned, a `500 Internal Server Error` response is generated.  For responses with
other status codes, an `*restapi.Error` value should be returned, obtained by calling one of the following
functions:

- `NewError()`
- `BadRequest()`
- `Forbidden()`
- `InternalServerError()`
- `NotFound()`
- `Unauthorized()`

All of these functions accept an optional set of `any` arguments and return an `*Error` value.
The arguments are applied according to type as follows:

- _**NewError() only:**_ the first of any `int` values is used as the HTTP `Error.Status` code
  (_any additional `int` values are ignored_)
- `string` values are concatenated with spaces as the `Error.Message`
- if one `error` value is provided, it is used as the `Error.Err`
- if multiple `error` values are provided, then `Error.Err` will be the result of
  `errors.Join()` on the provided errors

> _**NOTE:** `int` arguments are ignored by all functions except `NewError()`_

If `NewError()` is called without any `int` argument, `500` will be used.  The `*Error` type provides
methods that allow additional details to be provided for the error response:

| Method | Description |
|--------|-------------|
| `WithHeader()`<br>`WithHeaders()`<br>`WithNonCanonicalHeader()` | Add canonical/non-canonical headers to the response |
| `WithHelp()` | Adds a `help` message to the response |
| `WithProperty()` | Adds a `key`:`value` property to the response |

### Error Response Mechanism and Customization

When constructing an error response, the details of a `*restapi.Error` are passed to the
`restapi.ProjectError` function to be projected onto a response model.

> _**NOTE:** the `restapi.ProjectError` function may be replaced by your application to project an
> error onto a custom model in order to provide custom error responses.  Care should be taken to
> ensure that the resulting model projected by any replacement function is compatible with the
> `Content-Type` marshalling requirements of your API; typically this involves supporting both JSON
> and XML marshalling_

The default implementation of `ProjectError` returns a value supporting both JSON and XML marshalling,
equivalent to:

```go
type struct {
   Status     int              `json:"status" xml:"status"`
   Error      string           `json:"error" xml:"error"`
   Message    string           `json:"message,omitempty" xml:"message,omitempty"`
   Help       string           `json:"help,omitempty" xml:"help,omitempty"`
   Path       string           `json:"path" xml:"path"`
   Query      string           `json:"query,omitempty" xml:"query,omitempty"`
   Timestamp  time.Time        `json:"timestamp" xml:"timestamp"`
   Additional map[string]any   `json:"additional,omitempty" xml:"additional,omitempty"`
}
```

| Field | Description |
|-------|-------------|
| `Status` | The HTTP status code |
| `Error` | HTTP status text for the `Status` code |
| `Message` | a message providing details of the error (if provided) |
| `Help` | A help message (if provided) |
| `Path` | The request path |
| `Query` | The request query string (if any) |
| `Timestamp` | The time the error occurred (UTC) |
| `Additional` | Additional properties (if any) |

#### Errors and Messages

If both a `Message` and one or more `error`s is associated with a `*restapi.Error` response,
the `Message` in the response will be formatted to present the `Message` appended to the `error`,
separated by a `:` character.

##### example

```go
    err := errors.New("missing id")
    return restapi.BadRequest(err, "an id must be provided in the url query string")
```

will yield a response similar to:

```json
{
  "status": 400,
  "error": "Bad Request",
  "message": "missing id: an id must be provided in the url query string",
  "path": "/get",
  "timestamp": "2021-09-01T12:00:00Z"
}
```

#### Example JSON Error Response (default)

```json
{
  "status": 400,
  "error": "Bad Request",
  "message": "missing id parameter",
  "path": "/get",
  "timestamp": "2021-09-01T12:00:00Z"
}
```

#### Example XML Error Response (default)

```xml
<error>
   <status>400</status>
   <error>Bad Request</error>
   <message>missing id parameter</message>
   <path>/get</path>
   <timestamp>2021-09-01T12:00:00Z</timestamp>
</error>
```

### Errors During Error Response Construction

If an error occurs when attempting to an error response, a generic `plain/text` response is returned
with details of the original error and the error that occurred during processing.

## Error Logging

Errors might occur during the implementation of a REST API application caused by problems with
the implementation of the API itself (as opposed to _meaningful_ error responses intentionally
returned by the API).

i.e. if an application provides a custom error projection, errors may occur during the projection
or marshalling process that are not meaningful to the client, but are important to the application
developer.  Similarly, marshalling errors may occur if endpoint functions return complex `struct`
types, especially when implementing XML marshalling.

Such errors will be returned by the API as `500 Internal Server Error` responses but it may be
helpful to also include them in application logs or even to `panic` when they occur.

To support this, the `restapi` package provides a `restapi.LogError` extension point; this is a
function variable initially set to a no-op implementation; an application may replace this with
a function that will be called with details of any error that occurs during the processing of a
response.  The function is called with a `restapi.InternalError` value:

```go
type InternalError struct {
   Err         error
   Help        string
   Message     string
   Request     *http.Request
   ContentType string
}
```

> _**NOTE:** this does not provide tags to support JSON or XML marshalling; it is intended
> for use in application logs and should be marshalled according to the requirements of the
> application log system_

## RFC7807 Support

> _**NOTE:** EXPERIMENTAL_

The `restapi` package provides experimental support for
[RFC7807](https://www.rfc-editor.org/rfc/rfc7807) problem details responses.

An RFC7807 Problem Detail response is produced when an endpoint function returns a `*restapi.Problem`.
A `*restapi.Problem` value can be obtained by calling the `restapi.NewProblem()` function with
details of the problem to be reported.

The `*Problem` type provides methods to set additional details for the problem response.  Only fields
that are set will be included in the response.

> _**NOTE:** RFC7807 support may be subject to significant change in future versions of the
> `restapi` package; support may be removed if adoption of RFC7807 is not deemed sufficient to warrant
> continuing support_.
