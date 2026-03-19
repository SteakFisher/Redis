package store

func (r Redis) StreamAdd(streamKey string, entryID string, keyArr []string, valArr []string) string {
	val := r.m[streamKey]

	var newStream []*map[string]string

	if val == nil {
		val = &RedisValue{
			Type:   Stream,
			Stream: &newStream,
		}
	}

	entry := map[string]string{
		"id": entryID,
	}

	for i, _ := range keyArr {
		entry[keyArr[i]] = valArr[i]
	}

	newStream = append(newStream, &entry)

	val.Stream = &newStream

	r.m[streamKey] = val

	return entryID
}
