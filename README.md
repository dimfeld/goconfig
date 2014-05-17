goconfig [![Build Status](https://travis-ci.org/dimfeld/goconfig.png?branch=master)](https://travis-ci.org/dimfeld/goconfig) [![GoDoc](http://godoc.org/github.com/dimfeld/goconfig?status.png)](http://godoc.org/github.com/dimfeld/goconfig)
========

Loads configuration from TOML files, allowing environment variables to override.

The package exposes a single function: `Load(configObj interface{}, reader io.Reader, envPrefix string)`.

configObj contains a pointer to a configuration structure. The reader should supply TOML data, or can be nil for Load to just read from the environment. The environment variables corresponding to a field has the form ENVPREFIX_FIELDNAME.

## Field Tags
A field in the configuration structure can be tagged with `goconfigenv:"VARNAME"`. When this is present, the system will query the environment variable `VARNAME` for the value. Note that the passed envPrefix is not appended in this case.  `goconfigenv:"-"` can be used to indicate that a particular member should never be loaded from the environment.

The TOML parser used by this package also supports tags. The presence of a `toml:"name"` tag causes the parser to look the the corresponding key in the TOML data to assign to the given member. As above, `toml:"-"` can be used to indicate that a particular key should not be loaded from TOML. Note that if `toml:"-"` is added to a member, this package will also be unable to get a value for it from the environment.

## Example

````go
func ExampleLoad() {
    type exampleConfigStruct struct {
        B bool
        I int
    }

    // Note that the fields must be exported for the TOML parser to see them.
    // The type itself does not have to be exported though.
    type exampleConfig struct {
        // Overridden TOML key
        B bool `toml:"boolean"`
        // Overridden TOML key and environment variable
        I int `toml:"SomeInt" goconfigenv:"SYSTEMINT"`
        // Overridden environment variable
        S       string `goconfigenv:"IMPORTANTSTRING"`
        IntList []int
        Str     exampleConfigStruct
        StrList []exampleConfigStruct
    }

    tomlExample := `
boolean = true
Someint = 1
s = "ABC"
intList = [1, 2, 3]
[str]
b = true
i = 2

[[strList]]
b = false
i = 3

[[strList]]
b = true
i = 4
`
    config := exampleConfig{}
    reader := bytes.NewReader([]byte(tomlExample))
    Load(&config, reader, "EXAMPLE")
    fmt.Println("Original data loaded")
    fmt.Println(config)

    // The environment still uses the member name.
    os.Setenv("EXAMPLE_B", "false")
    // Overridden environment variable name. Note that this doesn't use the prefix.
    os.Setenv("SYSTEMINT", "3")
    // The parser will automatically add quotes around a string variable if
    // not present in the environment.
    os.Setenv("IMPORTANTSTRING", "DEF")
    // Lists of primitives follow the TOML format
    os.Setenv("EXAMPLE_INTLIST", "[3,4,5]")
    // Single newline separates structure items
    os.Setenv("EXAMPLE_STR", "b = false\ni = 4")
    // Single newline separate structure items
    // Double newline separates list items
    os.Setenv("EXAMPLE_STRLIST", "b = true\ni=6\n\nb = false\ni = 7\n\nb=true\ni=8")
    Load(&config, reader, "EXAMPLE")

    fmt.Println("Data overridden by environment")
    fmt.Println(config)

    // Output:
    // Original data loaded
    // {true 1 ABC [1 2 3] {true 2} [{false 3} {true 4}]}
    // Data overridden by environment
    // {false 3 DEF [3 4 5] {false 4} [{true 6} {false 7} {true 8}]}
}
````

## Acknowledgements

This package uses the TOML parser from [github.com/BurntSushi/toml](https://github.com/BurntSushi/toml)
