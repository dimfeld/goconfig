package goconfig

import (
	"bytes"
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
