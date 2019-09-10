package main

import antidote "github.com/AntidoteDB/antidote-go-client"

type BObject struct {
	updates []*antidote.ApbUpdateOp
	reads   []*antidote.ApbBoundObject
}

var BObjects = map[string]func(bucket *antidote.Bucket, keys []antidote.Key, read bool, write bool) BObject{
	"counter": counterObject,
}

func counterObject(bucket *antidote.Bucket, keys []antidote.Key, read bool, write bool) BObject {
	object := BObject{}

	if write {
		updates := make([]*antidote.ApbUpdateOp, len(keys))
		for i, key := range keys {
			updateOp := (&antidote.CRDTUpdate{
				Key:  key,
				Type: antidote.CRDTType_COUNTER,
				Update: &antidote.ApbUpdateOperation{
					Counterop: &antidote.ApbCounterUpdate{Inc: nil},
				}}).ConvertToToplevel(bucket.Bucket)
			updates[i] = updateOp
		}
		object.updates = updates
	}

	if read {
		reads := make([]*antidote.ApbBoundObject, len(keys))
		for i, key := range keys {
			crdtType := antidote.CRDTType_COUNTER
			boundOp := &antidote.ApbBoundObject{Bucket: bucket.Bucket, Key: key, Type: &crdtType}
			reads[i] = boundOp
		}
		object.reads = reads
	}

	return object
}
