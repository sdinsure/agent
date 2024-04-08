package reflection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Foo string `json:"foo"`
}

func TestStructTest(t *testing.T) {
	ts := &testStruct{Foo: "strfoo"}

	fieldValue, found := GetStringValue(ts, "Foo")
	assert.True(t, found)
	assert.EqualValues(t, fieldValue, "strfoo")

	fieldValue, found = GetStringValue(ts, "Foo2")
	assert.False(t, found)
	assert.EqualValues(t, fieldValue, "")
}
