package common

import (
	"strconv"
	"strings"
)

func UnmarshalTraversalIDS(traversalIDs string) ([]uint, error) {
	splitIds := strings.Split(traversalIDs, ",")
	var ids = make([]uint, len(splitIds))
	for i, id := range splitIds {
		ii, err := strconv.Atoi(id)
		if err != nil {
			return nil, err
		}
		ids[i] = uint(ii)
	}
	return ids, nil
}
