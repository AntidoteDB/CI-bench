package main

import (
	antidote "github.com/AntidoteDB/antidote-go-client"
	"math/rand"
)

type BObject struct {
	updates []*antidote.ApbUpdateOp
	reads   []*antidote.ApbBoundObject
}

var BObjects = map[string]func(bucket *antidote.Bucket, keys []antidote.Key, read bool, write bool) BObject{
	"counter": counterObject,
	"set":     setObject,
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

func setObject(bucket *antidote.Bucket, keys []antidote.Key, read bool, write bool) BObject {
	object := BObject{}

	if write {
		optype := antidote.ApbSetUpdate_ADD
		updates := make([]*antidote.ApbUpdateOp, len(keys))
		for i, key := range keys {
			elems := make([][]byte, 10000)
			for _,elem := range elems {
				elem = make([]byte, 8)
				rand.Read(elem)
			}
			updateOp := (&antidote.CRDTUpdate{
				Key:  key,
				Type: antidote.CRDTType_ORSET,
				Update: &antidote.ApbUpdateOperation{
					Setop: &antidote.ApbSetUpdate{Adds: elems, Optype: &optype},
				}}).ConvertToToplevel(bucket.Bucket)
			updates[i] = updateOp
		}
		object.updates = updates
	}

	if read {
		reads := make([]*antidote.ApbBoundObject, len(keys))
		for i, key := range keys {
			crdtType := antidote.CRDTType_ORSET
			boundOp := &antidote.ApbBoundObject{Bucket: bucket.Bucket, Key: key, Type: &crdtType}
			reads[i] = boundOp
		}
		object.reads = reads
	}

	return object
}
