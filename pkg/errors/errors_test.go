package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomizedError(t *testing.T) {
	stde := errors.New("std: err")
	e := New(CodeNotFound, errors.New("not found"))
	assert.EqualValues(t, "code(1), not found", e.Error())
	assert.EqualValues(t, CodeNotFound, e.Code())

	isCustomizedErr, err := As(e)
	assert.True(t, isCustomizedErr)
	assert.NotNil(t, err)
	assert.EqualValues(t, err, e)

	isCustomizedErr, err = As(stde)
	assert.False(t, isCustomizedErr)
	assert.Nil(t, err)
}
