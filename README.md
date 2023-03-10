# README

AnyMail

## Getting started

Before running the application you will need a working PostgreSQL installation and a valid DSN (data source name) for connecting to the database.

Please open the `cmd/api/main.go` file and edit the `db-dsn` command-line flag to include your valid DSN as the default value.

```
flag.StringVar(&cfg.db.dsn, "db-dsn", "YOUR DSN GOES HERE", "postgreSQL DSN")
```

Note that this DSN must be in the format `user:pass@localhost:port/db` and **not** be prefixed with `postgres://`.

Make sure that you're in the root of the project directory, fetch the dependencies with `go mod tidy`, then run the application using `go run ./cmd/api`:

```
$ go mod tidy
$ go run ./cmd/api
```

If you make a request to the `GET /status` endpoint using `curl` you should get a response like this:

```
$ curl -i localhost:4444/status
HTTP/1.1 200 OK
Content-Type: application/json
Date: Mon, 09 May 2022 20:46:37 GMT
Content-Length: 23

{
    "Status": "OK",
}
```

## Project structure

Everything in the codebase is designed to be editable. Feel free to change and adapt it to meet your needs.

|                        |                                                            |
| ---------------------- | ---------------------------------------------------------- |
| **`assets`**           | Contains the non-code assets for the application.          |
| `↳ assets/emails/`     | Contains email templates.                                  |
| `↳ assets/migrations/` | Contains SQL migrations.                                   |
| `↳ assets/efs.go`      | Declares an embedded filesystem containing all the assets. |

|                           |                                                                                                                                                                                        |
| ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`cmd/api`**             | Your application-specific code (handlers, routing, middleware, helpers) for dealing with HTTP requests and responses.                                                                  |
| `↳ cmd/api/errors.go`     | Contains helpers for managing and responding to error conditions.                                                                                                                      |
| `↳ cmd/api/handlers.go`   | Contains your application HTTP handlers.                                                                                                                                               |
| `↳ cmd/api/main.go`       | The entry point for the application. Responsible for parsing configuration settings initializing dependencies and running the server. Start here when you're looking through the code. |
| `↳ cmd/api/middleware.go` | Contains your application middleware.                                                                                                                                                  |
| `↳ cmd/api/routes.go`     | Contains your application route mappings.                                                                                                                                              |
| `↳ cmd/api/server.go`     | Contains a helper functions for starting and gracefully shutting down the server.                                                                                                      |

|                         |                                                                      |
| ----------------------- | -------------------------------------------------------------------- |
| **`internal`**          | Contains various helper packages used by the application.            |
| `↳ internal/database/`  | Contains your database-related code (setup, connection and queries). |
| `↳ internal/funcs/`     | Contains custom template functions.                                  |
| `↳ internal/request/`   | Contains helper functions for decoding JSON requests.                |
| `↳ internal/response/`  | Contains helper functions for sending JSON responses.                |
| `↳ internal/smtp/`      | Contains a SMTP sender implementation.                               |
| `↳ internal/validator/` | Contains validation helpers.                                         |
| `↳ internal/version/`   | Contains the application version number definition.                  |

## Configuration settings

Configuration settings are managed via command-line flags in `main.go`.

You can try this out by using the `--http-port` flag to configure the network port that the server is listening:

```
$ go run ./cmd/api --http-port=9999
```

Feel free to adapt the `main()` function to parse additional command-line flags and store their values in the `config` struct. For example, to add a configuration setting to enable a 'debug mode' in your application you could do this:

```
type config struct {
    httpPort  int
    debug     bool
}

...

func main() {
    var cfg config

    flag.IntVar(&cfg.httpPort, "http-port", 4444, "port to listen on for HTTP requests")
    flag.BoolVar(&cfg.debug, "debug", false, "enable debug mode")

    flag.Parse()

    ...
}
```

If you don't want to use command-line flags for configuration, feel free to adapt the code so that the `config` struct is populated from environment variables or a settings file instead.

## Creating new handlers

Handlers are defined as `http.HandlerFunc` methods on the `application` struct. They take the pattern:

```
func (app *application) yourHandler(w http.ResponseWriter, r *http.Request) {
    // Your handler logic...
}
```

Handlers are defined in the `cmd/api/handlers.go` file. For small applications, it's fine for all handlers to live in this file. For larger applications (10+ handlers) you may wish to break them out into separate files.

## Handler dependencies

Any dependencies that your handlers have should be initialized in the `main()` function `cmd/api/main.go` and added to the `application` struct. All of your handlers, helpers and middleware that are defined as methods on `application` will then have access to them.

You can see an example of this in the `cmd/api/main.go` file where we initialize a new `logger` instance and add it to the `application` struct.

## Creating new routes

[HttpRouter](https://github.com/julienschmidt/httprouter) is used for routing. Routes are defined in the `routes()` method in the `cmd/api/routes.go` file. For example:

```
func (app *application) routes() http.Handler {
    mux := httprouter.New()

    mux.HandlerFunc("GET", "/your/path", app.yourHandler)

    return mux
}
```

For more information about HttpRouter and example usage, please see the [official documentation](https://github.com/julienschmidt/httprouter).

## Adding middleware

Middleware is defined as methods on the `application` struct in the `cmd/api/middleware.go` file. Feel free to add your own. They take the pattern:

```
func (app *application) yourMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Your middleware logic...
        next.ServeHTTP(w, r)
    })
}
```

You can then use this middleware by wrapping the router before returning it from the `routes()` method, like so:

```
func (app *application) routes() http.Handler {
    mux := httprouter.New()

    mux.HandlerFunc("GET", "/your/path", app.yourHandler)

    // Wrap the router with middleware.
    return app.yourMiddlware(app.yourOtherMiddleware(mux))
}
```

It's possible to use middleware on specific routes only:

```
func (app *application) routes() http.Handler {
    mux := httprouter.New()

    mux.HandlerFunc("GET", "/your/path", app.yourHandler)

    // Wrap this handler with route-specific middleware. Note that when
    // wrapping handler functions with route-specific middleware that you
    // need to convert them to a http.Handler by using the http.HandlerFunc()
    // adapter. Like so:
    mux.Handler("GET", "/your/other/path", app.yourOtherMiddleware(http.HandlerFunc(app.yourOtherHandler)))

    return app.yourMiddleware(mux)
}
```

## Sending JSON responses

JSON responses and a specific HTTP status code can be sent using the `response.JSON()` function. The `data` parameter can be any JSON-marshalable type.

```
func (app *application) yourHandler(w http.ResponseWriter, r *http.Request) {
    data := map[string]string{"hello":  "world"}

    err := response.JSON(w, http.StatusOK, data)
    if err != nil {
        app.serverError(w, r, err)
    }
}
```

Specific HTTP headers can optionally be sent with the response too:

```
func (app *application) yourHandler(w http.ResponseWriter, r *http.Request) {
    data := map[string]string{"hello":  "world"}

    headers := make(http.Header)
    headers.Set("X-Server", "Go")

    err := response.JSONWithHeaders(w, http.StatusOK, data, headers)
    if err != nil {
        app.serverError(w, r, err)
    }
}
```

## Parsing JSON requests

HTTP requests containing a JSON body can be decoded using the `request.DecodeJSON()` function. For example, to decode JSON into an `input` struct:

```
func (app *application) yourHandler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Name string `json:"Name"`
        Age  int    `json:"Age"`
    }

    err := request.DecodeJSON(w, r, &input)
    if err != nil {
        app.badRequest(w, r, err)
        return
    }

    ...
}
```

Note: The target decode destination passed to `request.DecodeJSON()` (which in the example above is `&input`) must be a non-nil pointer.

The `request.DecodeJSON()` function returns friendly, well-formed, error messages that are suitable to be sent directly to the client using the `app.badRequest()` helper.

## Validating JSON requests

The `internal/validator` package includes a simple (but powerful) `validator.Validator` type that you can use to carry out validation checks.

Extending the example above:

```
func (app *application) yourHandler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Name      string              `json:"Name"`
        Age       int                 `json:"Age"`
        Validator validator.Validator `json:"-"`
    }

    err := request.DecodeJSON(w, r, &input)
    if err != nil {
        app.badRequest(w, r, err)
        return
    }

    input.Validator.CheckField(input.Name != "", "Name", "Name is required")
    input.Validator.CheckField(input.Age != 0, "Age", "Age is required")
    input.Validator.CheckField(input.Age >= 21, "Age", "Age must be 21 or over")

    if input.Validator.HasErrors() {
        app.failedValidation(w, r, input.Validator)
        return
    }

    ...
}
```

The `app.failedValidation()` helper will send a `422` status code along with any validation error messages. For the example above, the JSON response will look like this:

```
{
    "FieldErrors": {
        "Age": "Age must be 21 or over",
        "Name": "Name is required"
    }
}
```

In the example above we use the `CheckField()` method to carry out validation checks for specific fields. You can also use the `Check()` method to carry out a validation check that is _not related to a specific field_. For example:

```
input.Validator.Check(input.Password == input.ConfirmPassword, "Passwords do not match")
```

The `validator.AddError()` and `validator.AddFieldError()` methods also let you add validation errors directly:

```
input.Validator.AddFieldError("Email", "This email address is already taken")
input.Validator.AddError("Passwords do not match")
```

The `internal/validator/helpers.go` file also contains some helper functions to simplify validations that are not simple comparison operations.

|                                            |                                                                          |
| ------------------------------------------ | ------------------------------------------------------------------------ |
| `NotBlank(value string)`                   | Check that the value contains at least one non-whitespace character.     |
| `MinRunes(value string, n int)`            | Check that the value contains at least n runes.                          |
| `MaxRunes(value string, n int)`            | Check that the value contains no more than n runes.                      |
| `Between(value, min, max T)`               | Check that the value is between the min and max values inclusive.        |
| `Matches(value string, rx *regexp.Regexp)` | Check that the value matches a specific regular expression.              |
| `In(value T, safelist ...T)`               | Check that a value is in a 'safelist' of specific values.                |
| `AllIn(values []T, safelist ...T)`         | Check that all values in a slice are in a 'safelist' of specific values. |
| `NotIn(value T, blocklist ...T)`           | Check that the value is not in a 'blocklist' of specific values.         |
| `NoDuplicates(values []T)`                 | Check that a slice does not contain any duplicate (repeated) values.     |
| `IsEmail(value string)`                    | Check that the value has the formatting of a valid email address.        |
| `IsURL(value string)`                      | Check that the value has the formatting of a valid URL.                  |

For example, to use the `Between` check your code would look similar to this:

```
input.Validator.CheckField(validator.Between(input.Age, 18, 30), "Age", "Age must between 18 and 30")
```

Feel free to add your own helper functions to the `internal/validator/helpers.go` file as necessary for your application.

## Working with the database

This codebase is set up to use PostgreSQL with the [lib/pq](https://github.com/lib/pq) driver. You can control which database you connect to using the `--db-dsn` command-line flag when starting the application to pass in a DSN, or by adapting the default value in `main()`.

The codebase is also configured to use [jmoiron/sqlx](https://github.com/jmoiron/sqlx), so you have access to the whole range of sqlx extensions as well as the standard library `Exec()`, `Query()` and `QueryRow()` methods .

The database is available to your handlers, middleware and helpers via the `application` struct. If you want, you can access the database and carry out queries directly. For example:

```
func (app *application) yourHandler(w http.ResponseWriter, r *http.Request) {
    ...

    _, err := app.db.Exec("INSERT INTO people (name, age) VALUES ($1, $2)", "Alice", 28)
    if err != nil {
        app.serverError(w, r, err)
        return
    }

    ...
}
```

Generally though, it's recommended to isolate your database logic in the `internal/database` package and extend the `DB` type to include your own methods. For example, you could create a `internal/database/people.go` file containing code like:

```
type Person struct {
    ID    int    `db:"id"`
    Name  string `db:"name"`
    Age   int    `db:"age"`
}

func (db *DB) NewPerson(name string, age int) error {
    _, err := db.Exec("INSERT INTO people (name, age) VALUES ($1, $2)", name, age)
    return err
}

func (db *DB) GetPerson(id int) (Person, error) {
    var person Person
    err := db.Get(&person, "SELECT * FROM people WHERE id = $1", id)
    return person, err
}
```

And then call this from your handlers:

```
func (app *application) yourHandler(w http.ResponseWriter, r *http.Request) {
    ...

    _, err := app.db.NewPerson("Alice", 28)
    if err != nil {
        app.serverError(w, r, err)
        return
    }

    ...
}
```

## Managing SQL migrations

The `Makefile` in the project root contains commands to easily create and work with database migrations:

|                                                |                                                                                       |
| ---------------------------------------------- | ------------------------------------------------------------------------------------- |
| `$ make migrations/new name=add_example_table` | Create a new database migration in the `assets/migrations` folder.                    |
| `$ make migrations/up`                         | Apply all up migrations.                                                              |
| `$ make migrations/down`                       | Apply all down migrations.                                                            |
| `$ make migrations/goto version=N`             | Migrate up or down to a specific migration (where N is the migration version number). |
| `$ make migrations/force version=N`            | Force the database to be specific version without running any migrations.             |
| `$ make migrations/version`                    | Display the currently in-use migration version.                                       |

Hint: You can run `$ make help` at any time for a reminder of these commands.

These `Makefile` tasks are simply wrappers around calls to the `github.com/golang-migrate/migrate/v4/cmd/migrate` tool. For more information, please see the [official documentation](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate).

By default all 'up' migrations are automatically run on application startup using embeded files from the `assets/migrations` directory. You can disable this by using the command-line flag `--db-automigrate=false` when running the application.

## Sending emails

The application is configured to support sending of emails via SMTP.

Email templates should be defined as files in the `assets/emails` folder. Each file should contain named templates for the email subject, plaintext body and — optionally — HTML body.

```
{{define "subject"}}Example subject{{end}}

{{define "plainBody"}}
This is an example body
{{end}}

{{define "htmlBody"}}
<!doctype html>
<html>
    <head>
        <meta name="viewport" content="width=device-width" />
        <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    </head>
    <body>
        <p>This is an example body</p>
    </body>
</html>
{{end}}
```

A further example can be found in the `assets/emails/example.tmpl` file. Note that your email templates automatically have access to the custom template functions defined in the `internal/funcs` package.

Emails can be sent from your handlers using `app.mailer.Send()`. For example, to send an email to `alice@example.com` containing the contents of the `assets/emails/example.tmpl` file:

```
func (app *application) yourHandler(w http.ResponseWriter, r *http.Request) {
    ...

    data := map[string]any{"Name": "Alice"}

    err := app.mailer.Send("alice@example.com", data, "example.tmpl")
    if err != nil {
        app.serverError(w, r, err)
        return
    }

   ...
}
```

Note: The second parameter to `Send()` should be a map or struct containing any dynamic data that you want to render in the email template.

The SMTP host, port, username, password and sender details can be configured using the `--smtp-host`, `--smtp-port`, `--smtp-username`, `--smtp-password` and `--smtp-from` command-line flags when starting the application, or by adapting the default values in `cmd/api/main.go`.

You may wish to use [Mailtrap](https://mailtrap.io/) or a similar tool for development purposes.

## Custom template functions

Custom template functions are defined in `internal/funcs/funcs.go` and are automatically made available to your

email templates when you use `app.mailer.Send()`
.

The following custom template functions are already included by default:

|                              |                                                                                                                                                                    |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `now`                        | Returns the current time.                                                                                                                                          |
| `timeSince arg1`             | Returns the time elapsed since arg1.                                                                                                                               |
| `timeUntil arg2`             | Returns the time until arg1.                                                                                                                                       |
| `formatTime arg1 arg2`       | Returns the time arg2 as formatted using the pattern arg1.                                                                                                         |
| `approxDuration arg1`        | Returns the approximate duration of arg1 in a 'human-friendly' format ("3 seconds", "2 months", "5 years") etc.                                                    |
| `uppercase arg1`             | Returns arg1 converted to uppercase.                                                                                                                               |
| `lowercase arg1`             | Returns arg1 converted to lowercase.                                                                                                                               |
| `pluralize arg1 arg2 arg3`   | If arg1 equals 1 then return arg2, otherwise return arg3.                                                                                                          |
| `slugify arg1`               | Returns the lowercase of arg1 with all non-ASCII characters and punctuation removed (expect underscores and hyphens). Whitespaces are also replaced with a hyphen. |
| `safeHTML arg1`              | Output the verbatim value of arg1 without escaping the content. This should only be used when arg1 is from a trusted source.                                       |
| `join arg1 arg2`             | Returns the values in arg1 joined using the separator arg2.                                                                                                        |
| `containsString arg1 arg2 `  | Returns true if arg1 contains the string value arg2.                                                                                                               |
| `incr arg1`                  | Increments arg1 by 1.                                                                                                                                              |
| `decr arg1`                  | Decrements arg1 by 1.                                                                                                                                              |
| `formatInt arg1`             | Returns arg1 formatted with commas as the thousands separator.                                                                                                     |
| `formatFloat arg1 arg2`      | Returns arg1 rounded to arg2 decimal places and formatted with commas as the thousands separator.                                                                  |
| `yesno arg1`                 | Returns "Yes" if arg1 is true, or "No" if arg1 is false.                                                                                                           |
| `urlSetParam arg1 arg2 arg3` | Returns the URL arg1 with the key arg2 and value arg3 added to the query string parameters.                                                                        |
| `urlDelParam arg1 arg2`      | Returns the URL arg1 with the key arg2 (and corresponding value) removed from the query string parameters.                                                         |

To add another custom template function, define the function in `internal/funcs/funcs.go` and add it to the `TemplateFuncs` map. For example:

```
var TemplateFuncs = template.FuncMap{
    ...
    "yourFunction": yourFunction,
}

func yourFunction(s string) (string, error) {
    // Do something...
}
```

## Running background tasks

A `backgroundTask()` helper is included in the `cmd/api/helpers.go` file. You can call this in your handlers, helpers and middleware to run any logic in a separate background goroutine. This useful for things like sending emails, or completing slow-running jobs.

You can call it like so:

```
func (app *application) yourHandler(w http.ResponseWriter, r *http.Request) {
    ...

    app.backgroundTask(func() {
        // The logic you want to execute in a background task goes here...
    })

    ...
}
```

Using the `backgroundTask()` helper will automatically recover any panics in the background task logic, and when performing a graceful shutdown the application will wait for any background tasks to finish running before it exits.

## Application version

The application version number is defined in a `Get()` function in the `internal/version/version.go` file. Feel free to change this as necessary.

```
package version

func Get() string {
    return "0.0.1"
}
```

## Changing the module path

The module path is currently set to `github.com/mofodox/anymail`. If you want to change this please find and replace all instances of `github.com/mofodox/anymail` in the codebase with your own module path.
