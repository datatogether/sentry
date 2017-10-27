# config
--
    import "github.com/datatogether/config"

config is hassle-free config struct setting. setting things like
configStruct.StringField to the value of the enviornment variable STRING_FIELD.
It wraps the lovely godotenv package, which itself is a go port of the ruby
dotenv gem.

Basic use has three steps:

    1. declare a struct that defines all your config values (eg: type cfg struct{ ... })
    2. (optional) create a file or two that sets enviornement vars.
    3. call config.Load(&cfg, "envfilepath")

config will read that env file, setting environment variables for the running
process and infer any ENVIRIONMENT_VARIABLE value to it's
cfg.EnvironmentVariable counterpart/

config uses a CamelCase -> CAMEL_CASE convention to map values. So a struct
field cfg.FieldName will map to an environment variable FIELD_NAME. Types are
inferred, with errors returned for invalid env var settings.

using an env file is optional, calling Load(cfg) is perfectly valid.

config uses the reflect package to infer & set values reflect is a good choice
here because setting config happens seldom in the course of a running program,
and these days if you're parsing JSON, the reflect package is already included
in your binary

## Usage

#### func  EnvVarKey

```go
func EnvVarKey(key string) string
```
EnvVarKey takes a CamelCase string and converts it to a CAMEL_CASE string
(uppercase snake_case)

#### func  Load

```go
func Load(dst interface{}, filenames ...string) error
```
Load will read any passed-in env file(s) and load them into ENV for this
process, then assign values to the passed in dst struct pointer. Call this
function as close as possible to the start of your program (ideally in main) If
you call Load without any args it will default to loading .env in the current
path You can otherwise tell it which files to load (there can be more than one)
like

    godotenv.Load("fileone", "filetwo")

It's important to note that it WILL NOT OVERRIDE an env variable that already
exists - consider the .env file to set dev vars or sensible defaults

#### func  Overload

```go
func Overload(dst interface{}, filenames ...string) error
```
Overload will read any passed-in env file(s) and load them into ENV for this
process, then assign values to the passed in dst struct pointer. Call this
function as close as possible to the start of your program (ideally in main) If
you call Overload without any args it will default to loading .env in the
current path You can otherwise tell it which files to load (there can be more
than one) like

    godotenv.Overload("fileone", "filetwo")

It's important to note this WILL OVERRIDE an env variable that already exists -
consider the .env file to forcefilly set all vars.
