package oxweb

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

type getDeepTest struct {
	key   string
	value interface{}
	ok    bool
}

var getDeepTests = []getDeepTest{
	getDeepTest{"a", 1., true},
	getDeepTest{"b", "foo", true},
	getDeepTest{"not_there", nil, false},
	getDeepTest{"c", map[string]interface{}{"d": 2.}, true},
	getDeepTest{"c.d", 2., true},
	getDeepTest{"array.2", nil, false},
	getDeepTest{"c.d.not_there", nil, false},
	//	getDeepTest{"array", [2]float64{2., 3.}, true},
	getDeepTest{"array.1", 3., true},
	getDeepTest{"array.foo", nil, false},
}

var jsonString = `{
	"a": 1,
	"b": "foo",
	"c": {
		"d": 2
	},
	"array": [2,3]
}`

func TestGetDeep(t *testing.T) {
	var fixture JSONData

	reader := strings.NewReader(jsonString)
	jsonBytes, _ := ioutil.ReadAll(reader)
	err := json.Unmarshal(jsonBytes, &fixture)

	if err != nil {
		t.Errorf("Couldn't read jsonString")
	}

	for _, test := range getDeepTests {
		value, ok := GetDeep(test.key, fixture)
		if ok != test.ok {
			t.Errorf("For key '%s', expected ok = %t, but was %t", test.key, test.ok, ok)
		}

		if !reflect.DeepEqual(value, test.value) {
			t.Errorf("For key '%s', expected value = %v, but was %v", test.key, test.value, value)
		}
	}
}
