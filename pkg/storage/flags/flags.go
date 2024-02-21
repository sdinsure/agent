package storageflags

import (
	"flag"
	"fmt"
)

func NewPrefixFlagSet(prefix string) *PrefixFlagSet {
	return NewPrefixFlagSetWithValues(prefix, ValueSet{
		AutoMigrate: false,
		DbEndpoint:  "",
		DbUser:      "",
		DbPassword:  "",
		DbName:      "",
	})
}

func NewPrefixFlagSetWithValues(prefix string, defaultValue ValueSet) *PrefixFlagSet {
	return &PrefixFlagSet{
		prefix:       prefix,
		defaultValue: defaultValue,
	}
}

type ValueSet struct {
	AutoMigrate bool
	DbEndpoint  string
	DbUser      string
	DbPassword  string
	DbName      string
}

type PrefixFlagSet struct {
	prefix string

	flagValue    ValueSet
	defaultValue ValueSet
}

func (p *PrefixFlagSet) WithDefaultValue(d ValueSet) {
	p.defaultValue = d
}

func (p *PrefixFlagSet) Init() {
	flag.BoolVar(&p.flagValue.AutoMigrate, p.FlagAutoMigrateName(), p.defaultValue.AutoMigrate, "run auto-migrate")
	flag.StringVar(&p.flagValue.DbEndpoint, p.FlagDbEndpointName(), p.defaultValue.DbEndpoint, "db address")
	flag.StringVar(&p.flagValue.DbUser, p.FlagDbUserName(), p.defaultValue.DbUser, "db user")
	flag.StringVar(&p.flagValue.DbPassword, p.FlagDbPasswordName(), p.defaultValue.DbPassword, "db password")
	flag.StringVar(&p.flagValue.DbName, p.FlagDbNameName(), p.defaultValue.DbName, "db name for database")
}

func (p *PrefixFlagSet) withPrefix(name string) string {
	return fmt.Sprintf("%s_%s", p.prefix, name)
}

func (p *PrefixFlagSet) FlagAutoMigrateName() string {
	return p.withPrefix("auto_migrate")
}

func (p *PrefixFlagSet) FlagDbEndpointName() string {
	return p.withPrefix("db_endpoint")
}

func (p *PrefixFlagSet) FlagDbUserName() string {
	return p.withPrefix("db_user")
}

func (p *PrefixFlagSet) FlagDbPasswordName() string {
	return p.withPrefix("db_password")
}

func (p *PrefixFlagSet) FlagDbNameName() string {
	return p.withPrefix("db_name")
}

func (p *PrefixFlagSet) FlagAutoMigrate() bool {
	return p.flagValue.AutoMigrate
}

func (p *PrefixFlagSet) FlagDbEndpoint() string {
	if p.flagValue.DbEndpoint == "" {
		panic(fmt.Sprintf("%s should be specified", p.FlagDbEndpointName()))
	}
	return p.flagValue.DbEndpoint
}

func (p *PrefixFlagSet) FlagDbUser() string {
	if p.flagValue.DbUser == "" {
		panic(fmt.Sprintf("%s should be specified", p.FlagDbUserName()))
	}
	return p.flagValue.DbUser
}

func (p *PrefixFlagSet) FlagDbPassword() string {
	if p.flagValue.DbPassword == "" {
		panic(fmt.Sprintf("%s should be specified", p.FlagDbPasswordName()))
	}
	return p.flagValue.DbPassword
}

func (p *PrefixFlagSet) FlagDbName() string {
	if p.flagValue.DbName == "" {
		panic(fmt.Sprintf("%s should be specified", p.FlagDbNameName()))
	}
	return p.flagValue.DbName
}
