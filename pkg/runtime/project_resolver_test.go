package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathResolver(t *testing.T) {
	for _, testcase := range []struct {
		reqPath string
		found   string
		matched bool
	}{
		{
			reqPath: "/v1/projects/1",
			found:   "1",
			matched: true,
		},
		{
			reqPath: "/v1/projects/1/",
			found:   "1",
			matched: true,
		},
	} {

		found, matched := findProjectIdFromPath(testcase.reqPath)
		assert.EqualValues(t, testcase.matched, matched)
		assert.EqualValues(t, testcase.found, found)

	}

}
