package storagepostgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	logger "github.com/sdinsure/agent/pkg/logger"
	storage "github.com/sdinsure/agent/pkg/storage"
	storageflags "github.com/sdinsure/agent/pkg/storage/flags"
)

type PostgresOption interface {
	apply(o *postgresOptions)
}

type postgresOptions struct {
	batchSize int
}

type BatchSize int

func (b BatchSize) apply(o *postgresOptions) {
	o.batchSize = int(b)
}

func newOptions(opts ...PostgresOption) *postgresOptions {
	o := &postgresOptions{
		batchSize: 1000,
	}

	for _, opt := range opts {
		opt.apply(o)
	}
	return o
}

func NewPostgresDbHelper(log logger.Logger, flagset *storageflags.PrefixFlagSet) (*PostgresDb, error) {
	dsn, err := storage.NewDSN(storage.Postgres, flagset.FlagDbEndpoint(), flagset.FlagDbUser(), flagset.FlagDbPassword(), flagset.FlagDbName(), nil)
	if err != nil {
		return nil, fmt.Errorf("storage: failed to construct dsn, err:%+v\n", err)
	}
	return NewPostgresDb(dsn, log, flagset)
}

func NewPostgresDb(dsn string, log logger.Logger, flagset *storageflags.PrefixFlagSet, opts ...PostgresOption) (*PostgresDb, error) {
	o := newOptions(opts...)

	newLogger := gormlogger.New(
		newGlogWriter(log),
		gormlogger.Config{
			SlowThreshold:             5 * time.Second, // Slow SQL threshold
			LogLevel:                  gormlogger.Info, // Log level
			IgnoreRecordNotFoundError: true,            // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,           // Disable color
		},
	)
	// add log for every query
	newLogger.LogMode(gormlogger.Info)
	gormDb, err := gorm.Open(openDriver(dsn), &gorm.Config{
		CreateBatchSize: o.batchSize,
		Logger:          newLogger,
	})
	if err != nil {
		return nil, err
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	rawDb, _ := gormDb.DB()

	rawDb.SetMaxIdleConns(20)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	rawDb.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	rawDb.SetConnMaxLifetime(time.Hour)

	if err := rawDb.Ping(); err != nil {
		return nil, err
	}

	return &PostgresDb{
		log:     log,
		gormDb:  gormDb,
		flagset: flagset,
	}, nil
}

func openDriver(dsn string) gorm.Dialector {
	if strings.Contains(dsn, storage.Postgres.ProtocolPrefix()) {
		return postgres.Open(dsn)
	}
	return nil
}

type PostgresDb struct {
	log     logger.Logger
	gormDb  *gorm.DB
	flagset *storageflags.PrefixFlagSet
}

func (p *PostgresDb) GormDB() *gorm.DB {
	return p.gormDb
}

func (p *PostgresDb) With(ctx context.Context, tableName string) *gorm.DB {
	db := p.gormDb.WithContext(ctx)
	if tableName != "" {
		p.log.Info("postgres: use table name %s\n", tableName)
		db = db.Table(tableName)
	}
	return db
}

type hasTableName interface {
	TableName() string
}

func (p *PostgresDb) AutoMigrate(tables []interface{}) error {
	if !p.flagset.FlagAutoMigrate() {
		p.log.Info("postgres: auto migrate is turn off. skipped. use --%s to enable it\n", p.flagset.FlagAutoMigrateName())
		return nil
	}
	p.log.Info("postgres: auto migrate is turn on, running\n")
	for _, table := range tables {
		_, hasTableNameInf := table.(hasTableName)
		if hasTableNameInf {
			p.log.Info("postgres: automigrate for table: %s\n", table.(hasTableName).TableName())
		}
		if err := p.gormDb.AutoMigrate(table); err != nil {
			p.log.Error("postgres: automigrate failed:%s\n", err)
			return err
		}
	}
	return nil
}

// AutoMigrateItemTemporaryTable will create a item table named $tablename
func (p *PostgresDb) AutoMigrateToName(tableName string, table interface{}) error {
	return p.gormDb.Table(tableName).AutoMigrate(table)
}
