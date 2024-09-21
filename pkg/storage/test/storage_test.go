package storage

import (
	"context"
	"testing"
	"time"

	"github.com/sdinsure/agent/pkg/errors"
	"github.com/sdinsure/agent/pkg/logger"
	storageerrors "github.com/sdinsure/agent/pkg/storage/errors"
	storagepostgres "github.com/sdinsure/agent/pkg/storage/postgres"
	storagetestutils "github.com/sdinsure/agent/pkg/storage/testutils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skip this test in short mode")
		return
	}
	// launch postgres with the following command
	// docker run --rm --name postgres \
	// -e TZ=gmt+8 \
	// -e POSTGRES_USER=postgres \
	// -e POSTGRES_PASSWORD=password \
	// -e POSTGRES_DB=unittest \
	// -p 5432:5432 -d library/postgres:14.1
	//
	// or run ./run_postgre.sh

	postgrescli, err := storagetestutils.NewTestPostgresCli(logger.NewLogger(true))
	assert.NoError(t, err)
	tcrud := &testCRUD{postgrescli}
	assert.Nil(t, tcrud.AutoMigrate())

	assert.Nil(t, tcrud.Write(context.Background(), "", &modelTest{Name: "name1"}))
	assert.Nil(t, tcrud.Write(context.Background(), "", &modelTest{Name: "name2"}))
	m, err := tcrud.Get(context.Background(), "", "name2")
	assert.Nil(t, err)
	assert.EqualValues(t, "name2", m.Name)

	m, err = tcrud.Get(context.Background(), "", "name3")
	assert.NotNil(t, err)
	assert.Nil(t, m)
}

type modelTest struct {
	Name      string         `gorm:"column:name;primaryKey"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (m modelTest) TableName() string {
	return "for-sdinsure-integration-test"
}

type testCRUD struct {
	*storagepostgres.PostgresDb
}

func (m *testCRUD) AutoMigrate() *errors.Error {
	tables := []interface{}{&modelTest{}}
	return storageerrors.WrapStorageError(m.PostgresDb.AutoMigrate(tables))
}

func (m *testCRUD) Write(ctx context.Context, tableName string, record *modelTest) *errors.Error {
	err := m.With(ctx, tableName).Clauses(clause.OnConflict{UpdateAll: true}).Create(record).Error
	return storageerrors.WrapStorageError(err)
}

func (m *testCRUD) Get(ctx context.Context, tableName string, name string) (*modelTest, *errors.Error) {
	record := &modelTest{}
	err := m.With(ctx, tableName).Where("name = ?", name).First(record).Error
	if err != nil {
		return nil, storageerrors.WrapStorageError(err)
	}
	return record, nil
}
