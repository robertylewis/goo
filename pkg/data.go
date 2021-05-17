package pkg

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

// Data interface - built from YAML data files.
type Data interface {
	getChild(key string) (Data, error)
	getValue() (string, error)
}

// DataNode - means there is more data below.
type DataNode map[string]Data
func (n DataNode) getChild(key string) (Data, error) {
	return n[key], nil
}
func (n DataNode) getValue() (string, error) {
	return "", GenericError{"ERROR: Called getValue on a Node."}
}
func (n DataNode) setTitle(title string) {
	n["title"] = DataLeaf(title)
}

// DataLeaf - means there is no more data below.
type DataLeaf string
func (l DataLeaf) getChild(key string) (Data, error) {
	return nil, GenericError{"ERROR: Called getChild on a Leaf."}
}
func (l DataLeaf) getValue() (string, error) {
	return string(l), nil
}

// Cast a generic unmarshal to the Data format above.
func castData(m map[interface{}]interface{}) (Data, error) {
	var err error
	data := make(map[string]Data)
	for k, v := range m {
		switch v := v.(type) {
		case map[interface{}]interface{}:
			data[k.(string)], err = castData(v)
			Check(err)
		case string:
			data[k.(string)] = DataLeaf(v)
		default:
			fmt.Printf("data = %T\n", v)
			return nil, GenericError{"Malformed Data."}
		}
	}
	return DataNode(data), nil
}

// Gets data from a file in the format above.
func GetDataFromFile(filename string) Data {
	// Reads data out into a string.
	bytes, io_err := ioutil.ReadFile(filename)
	Check(io_err)

	// Unmarshal into m.
	m := make(map[interface{}]interface{})
	y_err := yaml.Unmarshal(bytes, &m)
	Check(y_err)

	// Build Data object.
	data, d_err := castData(m)
	Check(d_err)
	return data
}

// Converts a map of data sources to the format above.
func GetData(m map[string]string) Data {
	data := make(DataNode)
	for k, v := range m {
		data[k] = GetDataFromFile(v)
	}
	return data
}

// Gets a specific datapoint (home.title) from data.
func ExtractData(payload string, data Data) (string, error) {
	fields := strings.Split(payload, ".")
	curr := data
	var c_err error
	for _, field := range fields {
		curr, c_err = curr.getChild(field)
		Check(c_err)
	}
	return curr.getValue()
}
