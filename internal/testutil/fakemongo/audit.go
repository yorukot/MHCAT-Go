package fakemongo

import mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"

func Snapshot(name string, count int64, samples []mhcatmongo.SampleDocument) mhcatmongo.CollectionSnapshot {
	return mhcatmongo.CollectionSnapshot{
		Name:          name,
		DocumentCount: count,
		Samples:       samples,
	}
}
