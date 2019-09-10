package main

import (
	antidote "github.com/AntidoteDB/antidote-go-client"
)

type Benchmark struct {
	function func(client *antidote.Client, objects *BObject) error
	init     func(client *antidote.Client, objects *BObject) error
	read     bool
	write    bool
}

var Benchmarks = map[string]Benchmark{
	"staticWrite": {
		function: staticWrite,
		init:     nil,
		read:     false,
		write:    true,
	},
	"staticRead": {
		function: staticRead,
		init:     initRead,
		read:     true,
		write:    false,
	},
	"interactiveReadWrite": {
		function: interactiveReadWrite,
		init:     initRead,
		read:     true,
		write:    true,
	},
}

func staticWrite(client *antidote.Client, objects *BObject) error {
	tx := client.CreateStaticTransaction()
	return tx.Update(objects.updates...)
}

func staticRead(client *antidote.Client, objects *BObject) error {
	tx := client.CreateStaticTransaction()
	_, err := tx.Read(objects.reads...)
	return err
}

func interactiveReadWrite(client *antidote.Client, objects *BObject) error {
	tx, err := client.StartTransaction()
	if err != nil {
		return err
	}
	_, err = tx.Read(objects.reads...)
	if err != nil {
		return err
	}
	err = tx.Update(objects.updates...)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func initRead(client *antidote.Client, objects *BObject) error {
	tx := client.CreateStaticTransaction()
	for _, update := range objects.updates {
		err := tx.Update(update)
		if err != nil {
			return err
		}
	}
	return nil
}
