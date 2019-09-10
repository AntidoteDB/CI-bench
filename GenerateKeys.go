package main

import (
	"gonum.org/v1/gonum/stat/distuv"
	"encoding/binary"
	antidote "github.com/AntidoteDB/antidote-go-client"
)

func GenerateKeys(distribution string, numKeys int) *[]antidote.Key {
	keys := make([]antidote.Key, numKeys)

	switch distribution {
	case "uniform":
		for i := 0; i < numKeys; i++ {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(i))
			keys[i] = b
		}
	case "single":
		for i := 0; i < numKeys; i++ {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, 1)
			keys[i] = b
		}
	case "paretoInt":
		pareto := distuv.Pareto{Xm: 1, Alpha: 1.5}
		for i := 0; i < numKeys; i++ {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(pareto.Rand()))
			keys[i] = b
		}
	}
	return &keys
}
