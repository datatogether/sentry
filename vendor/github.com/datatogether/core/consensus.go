package core

// Consensus is an enumeration of Meta graph values arranged by key
type Consensus map[string]map[string]int

// Metadata takes a store and gives back the actual metadata based on a provided stringMap
// Any key present in the consensus that isn't found in data will write the hash value instead
// Returned map should be valid for JSON encoding
func (c Consensus) Metadata(data map[string]interface{}) (map[string][]interface{}, error) {
	con := map[string][]interface{}{}
	for key, votes := range c {
		for hash, _ := range votes {
			if data[hash] != nil {
				con[key] = append(con[key], data[hash])
			} else {
				con[key] = append(con[key], hash)
			}
		}
	}

	return con, nil
}

// SumConsensus tallies the consensus around a given subject hash from a provided Metadata slice
func SumConsensus(subject string, blocks []*Metadata) (c Consensus, values map[string]interface{}, err error) {
	c = Consensus{}
	values = map[string]interface{}{}
	var (
		hashMap  map[string]string
		valueMap map[string]interface{}
	)

	for _, bl := range blocks {
		if bl.Subject == subject {
			hashMap, valueMap, err = bl.HashMaps()
			if err != nil {
				return
			}

			for key, hash := range hashMap {
				values[hash] = valueMap[hash]

				if c[key] == nil {
					c[key] = map[string]int{
						hash: 1,
					}
				} else {
					c[key][hash]++
				}
			}
		}
	}

	return
}
