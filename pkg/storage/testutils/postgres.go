package testutil

import (
	"github.com/sdinsure/agent/pkg/logger"
	storageflags "github.com/sdinsure/agent/pkg/storage/flags"
	storagepostgres "github.com/sdinsure/agent/pkg/storage/postgres"
)

var testutilflagset = storageflags.NewPrefixFlagSetWithValues("testutils", storageflags.ValueSet{
	AutoMigrate: true,
	DbEndpoint:  "127.0.0.1:5432",
	DbUser:      "postgres",
	DbPassword:  "password",
	DbName:      "unittest",
})

func init() {
	testutilflagset.Init()
}

func NewTestPostgresCli(log logger.Logger) (*storagepostgres.PostgresDb, error) {
	return storagepostgres.NewPostgresDbHelper(log, testutilflagset)
}
