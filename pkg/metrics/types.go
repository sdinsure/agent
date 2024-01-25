package metrics

import "fmt"

type TypeNamespace string

func NewTypeNamespace(ns string) TypeNamespace {
	return TypeNamespace(ns)
}

func (t TypeNamespace) String() string {
	return string(t)
}

type TypeSubsystem string

func NewTypeSubsystem(ss string) TypeSubsystem {
	return TypeSubsystem(ss)
}

func (t TypeSubsystem) String() string {
	return string(t)
}

type TypeMetricName string

func NewTypeMetricName(n string) TypeMetricName {
	return TypeMetricName(n)
}

func (t TypeMetricName) String() string {
	return string(t)
}

func normalizedNames(ns TypeNamespace, ss TypeSubsystem, name TypeMetricName) string {
	return fmt.Sprintf("%s_%s_%s", ns.String(), ss.String(), name.String())
}
