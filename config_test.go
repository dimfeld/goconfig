package goconfig

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"testing"
)

type testConfigStruct struct {
	B bool
	I int
}

type testData struct {
	B       bool
	I       int
	S       string
	IntList []int
	Str     testConfigStruct
	StrList []testConfigStruct
}

type tagTestData struct {
	BothTags string `toml:"customTomlTag" goconfigenv:"CUSTOMENVTAG"`
	TomlTag  string `toml:"tomlOnlyAttr"`
	EnvTag   string `goconfigenv:"GOCONFIGONLYATTR"`
	NoEnvTag string `goconfigenv:"-"`
	NoTag    string `toml:"-"`
}

var tomlFile []byte = []byte(`
b = true
i = 1
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
`)

var tomlTagFile []byte = []byte(`
customTomlTag = "a"
tomlOnlyAttr = "a"
envTag = "a"
noEnvTag = "a"
`)

func TestNoEnv(t *testing.T) {
	data := testData{}
	reader := bytes.NewReader(tomlFile)

	os.Clearenv()
	err := Load(&data, reader, "GOCONFIGTEST")
	if err != nil {
		t.Fatal("Failed parsing data with error", err)
	}

	expectedData := testData{
		B:       true,
		I:       1,
		S:       "ABC",
		IntList: []int{1, 2, 3},
		Str:     testConfigStruct{true, 2},
		StrList: []testConfigStruct{
			testConfigStruct{false, 3},
			testConfigStruct{true, 4},
		},
	}
	if !reflect.DeepEqual(data, expectedData) {
		t.Errorf("\nExpected %v\nsaw      %v", expectedData, data)
	}

}

func TestEnv(t *testing.T) {
	data := testData{}
	reader := bytes.NewReader(tomlFile)

	os.Clearenv()
	os.Setenv("GOCONFIGTEST_B", "false")
	os.Setenv("GOCONFIGTEST_I", "3")
	os.Setenv("GOCONFIGTEST_S", "DEF")
	os.Setenv("GOCONFIGTEST_INTLIST", "[3,4,5]")
	os.Setenv("GOCONFIGTEST_STR", "b = false\ni = 4")
	os.Setenv("GOCONFIGTEST_STRLIST", "b = true\ni=6\n\nb = false\ni = 7\n\nb=true\ni=8")

	err := Load(&data, reader, "GOCONFIGTEST")
	if err != nil {
		t.Fatal("Failed parsing data with error", err)
	}

	expectedData := testData{
		B:       false,
		I:       3,
		S:       "DEF",
		IntList: []int{3, 4, 5},
		Str:     testConfigStruct{false, 4},
		StrList: []testConfigStruct{
			testConfigStruct{true, 6},
			testConfigStruct{false, 7},
			testConfigStruct{true, 8},
		},
	}
	if !reflect.DeepEqual(data, expectedData) {
		t.Errorf("\nExpected %v\nsaw      %v", expectedData, data)
	}
}

func TestOneEnv(t *testing.T) {
	data := testData{}
	reader := bytes.NewReader(tomlFile)

	os.Clearenv()
	os.Setenv("GOCONFIGTEST_I", "2")
	err := Load(&data, reader, "GOCONFIGTEST")
	if err != nil {
		t.Fatal("Failed parsing data with error", err)
	}

	expectedData := testData{
		B:       true,
		I:       2,
		S:       "ABC",
		IntList: []int{1, 2, 3},
		Str:     testConfigStruct{true, 2},
		StrList: []testConfigStruct{
			testConfigStruct{false, 3},
			testConfigStruct{true, 4},
		},
	}
	if !reflect.DeepEqual(data, expectedData) {
		t.Errorf("\nExpected %v\nsaw      %v", expectedData, data)
	}
}

func TestConfigTags(t *testing.T) {
	data := tagTestData{NoTag: "z"}
	reader := bytes.NewReader(tomlTagFile)

	os.Clearenv()
	os.Setenv("CUSTOMENVTAG", "b")
	os.Setenv("GOCONFIGTEST_TOMLTAG", "c")
	os.Setenv("GOCONFIGONLYATTR", "d")
	os.Setenv("GOCONFIGTEST_NOENVTAG", "e") // Should be ignored
	os.Setenv("GOCONFIGTEST_NOTAG", "f")    // Should be ignored
	err := Load(&data, reader, "GOCONFIGTEST")
	if err != nil {
		t.Fatal("Failed parsing data with error", err)
	}

	expectedData := tagTestData{
		BothTags: "b",
		TomlTag:  "c",
		EnvTag:   "d",
		NoEnvTag: "a",
		NoTag:    "z",
	}
	if !reflect.DeepEqual(data, expectedData) {
		t.Errorf("\nExpected %v\nsaw      %v", expectedData, data)
	}
}

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
