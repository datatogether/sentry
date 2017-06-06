// config is hassle-free config struct setting. setting things like configStruct.StringField to
// the value of the enviornment variable STRING_FIELD.
// It wraps the lovely godotenv package, which itself is a go port of the ruby dotenv gem.
//
// Basic use has three steps:
//  1. declare a struct that defines all your config values (eg: type cfg struct{ ... })
//  2. (optional) create a file or two that sets enviornement vars.
//  3. call config.Load(&cfg, "envfilepath")
//
// config will read that env file, setting environment variables for the running process
// and infer any ENVIRIONMENT_VARIABLE value to it's cfg.EnvironmentVariable counterpart/
//
// config uses a CamelCase -> CAMEL_CASE convention to map values. So a struct field
// cfg.FieldName will map to an environment variable FIELD_NAME.
// Types are inferred, with errors returned for invalid env var settings.
//
// using an env file is optional, calling Load(cfg) is perfectly valid.
//
// config uses the reflect package to infer & set values
// reflect is a good choice here because setting config happens seldom in the course
// of a running program, and these days if you're parsing JSON, the reflect package is already
// included in your binary
package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

// Load will read any passed-in env file(s) and load them into ENV for this process,
// then assign values to the passed in dst struct pointer.
// Call this function as close as possible to the start of your program (ideally in main)
// If you call Load without any args it will default to loading .env in the current path
// You can otherwise tell it which files to load (there can be more than one) like
//		godotenv.Load("fileone", "filetwo")
// It's important to note that it WILL NOT OVERRIDE an env variable that already exists - consider the .env file to set dev vars or sensible defaults
func Load(dst interface{}, filenames ...string) error {
	cfg := reflect.ValueOf(dst).Elem()
	if !cfg.CanSet() {
		return fmt.Errorf("dst must be a struct pointer")
	}

	if len(filenames) > 0 {
		if err := godotenv.Load(filenames...); err != nil {
			return err
		}
	}

	return setValues(dst)
}

// Overload will read any passed-in env file(s) and load them into ENV for this process,
// then assign values to the passed in dst struct pointer.
// Call this function as close as possible to the start of your program (ideally in main)
// If you call Overload without any args it will default to loading .env in the current path
// You can otherwise tell it which files to load (there can be more than one) like
//		godotenv.Overload("fileone", "filetwo")
// It's important to note this WILL OVERRIDE an env variable that already exists - consider the .env file to forcefilly set all vars.
func Overload(dst interface{}, filenames ...string) error {
	cfg := reflect.ValueOf(dst).Elem()
	if !cfg.CanSet() {
		return fmt.Errorf("dst must be a struct pointer")
	}

	if len(filenames) > 0 {
		if err := godotenv.Overload(filenames...); err != nil {
			return err
		}
	}

	return setValues(dst)
}

// setValues maps values from environment variables to dst.
func setValues(dst interface{}) error {
	cfg := reflect.ValueOf(dst).Elem()
	typeOfT := cfg.Type()
	for i := 0; i < cfg.NumField(); i++ {
		envVarKey := EnvVarKey(typeOfT.Field(i).Name)
		f := cfg.Field(i)
		strVal := os.Getenv(envVarKey)

		switch f.Kind() {
		case reflect.String:
			f.SetString(strVal)
		case reflect.Int:
			v, err := strconv.ParseInt(strVal, 10, 64)
			if err != nil {
				return fmt.Errorf("error converting %s value to an integer: %s", envVarKey, err.Error())
			}
			f.SetInt(v)
		case reflect.Bool:
			f.SetBool(strVal == "true" || strVal == "TRUE" || strVal == "t")
		case reflect.Slice:
			switch f.Interface().(type) {
			case []string:
				f.Set(reflect.ValueOf(strings.Split(strVal, ",")))
				continue
			case []byte:
				f.Set(reflect.ValueOf([]byte(strVal)))
				continue
			}
			return fmt.Errorf("config package currently doesn't support setting values of type: %s. cannot set: %s", f.Kind(), envVarKey)
		default:
			return fmt.Errorf("config package currently doesn't support setting values of type: %s. cannot set: %s", f.Kind(), envVarKey)
		}
	}
	return nil
}

// EnvVarKey takes a CamelCase string and converts it to a CAMEL_CASE string (uppercase snake_case)
func EnvVarKey(key string) string {
	return strings.ToUpper(camelToSnake(key))
}

// camelToSnake converts CamelCase to snake_case
// slightly tweaked version of https://github.com/serenize/snaker
// TODO - initialism support for config vars like JSONConfigValue, etc.
func camelToSnake(s string) string {
	var words []string
	var lastPos int
	rs := []rune(s)

	for i := 0; i < len(rs); i++ {
		if i > 0 && unicode.IsUpper(rs[i]) && !unicode.IsUpper(rs[i-1]) {
			// if initialism := startsWithInitialism(s[lastPos:]); initialism != "" {
			// 	words = append(words, initialism)

			// 	i += len(initialism) - 1
			// 	lastPos = i
			// 	continue
			// }

			words = append(words, s[lastPos:i])
			lastPos = i
		}
	}

	// append the last word
	if s[lastPos:] != "" {
		words = append(words, s[lastPos:])
	}

	return strings.Join(words, "_")
}
