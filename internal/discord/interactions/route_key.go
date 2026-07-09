package interactions

import "fmt"

type RouteKey struct {
	Kind    Type
	Feature string
	Action  string
	Version string
	Legacy  bool
}

func (k RouteKey) IsZero() bool {
	return k.Kind == "" && k.Feature == "" && k.Action == "" && k.Version == ""
}

func (k RouteKey) String() string {
	return fmt.Sprintf("%s:%s:%s:%s:%t", k.Kind, k.Version, k.Feature, k.Action, k.Legacy)
}
