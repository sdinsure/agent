package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgresDsn(t *testing.T) {
	got, err := NewDSN(Postgres, "1.2.3.4:5432", "foo", "bar", "myfirstdb", nil)
	assert.NoError(t, err)
	assert.EqualValues(t, "postgres://foo:bar@1.2.3.4:5432/myfirstdb?connect_timeout=10&timezone=Asia%2FTaipei", got)
}
