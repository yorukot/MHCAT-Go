package fakemongo

import mhcatmongo "github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/adapters/mongo"

func LiveIndexes(collection string, indexes ...mhcatmongo.IndexInfo) map[string][]mhcatmongo.IndexInfo {
	for index := range indexes {
		if indexes[index].Collection == "" {
			indexes[index].Collection = collection
		}
	}
	return map[string][]mhcatmongo.IndexInfo{collection: indexes}
}
