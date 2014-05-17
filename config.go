package goconfig

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"io"
	"os"
	"reflect"
	"strings"
)

// Load gets values from the TOML file specified in filename. It then
// searches the environment for variables that match members in the config.
// The environment variables take the form of ENVPREFIX_UPPERCASEMEMBER.
//
// Struct tags may be used to override the handling of a member.
// The TOML loading may be customized using the struct tags defined at
// github.com/BurntSushi/toml.
//
// Environment variable names may be customized using the format
// `goconfigenv:"ENVNAME"`. Variables customized in this way are not
// prefixed by envPrefix. If ENVNAME is "-", the variable will never be
// loaded from the environment. The values read from the environment are
// also fed into the TOML parser.
func Load(configObj interface{}, reader io.Reader, envPrefix string) error {
	t := reflect.TypeOf(configObj)
	if t.Kind() != reflect.Ptr {
		return errors.New("configObj is not a pointer")
	}

	v := reflect.Indirect(reflect.ValueOf(configObj))
	t = v.Type()
	if t.Kind() != reflect.Struct {
		return errors.New("configObj doesn't point to a struct")
	}

	// First, do the TOML input.
	if reader != nil {
		_, err := toml.DecodeReader(reader, configObj)
		if err != nil {
			return err
		}
	}

	numFields := t.NumField()
	for i := 0; i < numFields; i++ {
		f := t.Field(i)
		tomlKey := f.Tag.Get("toml")
		if tomlKey == "" {
			tomlKey = f.Name
		} else if tomlKey == "-" {
			continue
		}

		envKey := f.Tag.Get("goconfigenv")
		if envKey == "" {
			envKey = envPrefix + "_" + strings.ToUpper(f.Name)
		} else if envKey == "-" {
			continue
		}

		envValue := os.Getenv(envKey)
		if envValue == "" {
			continue
		}

		var envToml string
		switch f.Type.Kind() {
		case reflect.Struct:
			envToml = fmt.Sprintf("[%s]\n%s\n", tomlKey, envValue)

		case reflect.Array, reflect.Slice:
			fv := v.Field(i)
			// Set it to nil.
			fv.Set(reflect.Zero(f.Type))

			if f.Type.Elem().Kind() == reflect.Struct {
				values := strings.Split(envValue, "\n\n")
				for _, value := range values {
					envToml = envToml + fmt.Sprintf("[[%s]]\n%s\n", tomlKey, value)
				}

			} else {
				envToml = fmt.Sprintf("%s = %s\n", tomlKey, envValue)
			}

		case reflect.String:
			if envValue[0] != '"' && envValue[len(envValue)-1] != '"' {
				envToml = fmt.Sprintf("%s = \"%s\"\n", tomlKey, envValue)
			} else {
				envToml = fmt.Sprintf("%s = %s\n", tomlKey, envValue)
			}

		default:
			envToml = fmt.Sprintf("%s = %s\n", tomlKey, envValue)
		}

		_, err := toml.Decode(envToml, configObj)
		if err != nil {
			return fmt.Errorf("Failed parsing environment config %s: %s", envKey, err.Error())
		}
	}

	return nil
}
