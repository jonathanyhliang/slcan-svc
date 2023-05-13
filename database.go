package main

import (
	"errors"
	"sync"
)

var (
	ErrDatabaseAlreadyExists = errors.New("Database: already exists")
	ErrDatabaseNotFound      = errors.New("Database: request not found")
)

type Message struct {
	ID   uint32 `json:"id"`
	Data string `json:"data"`
}

type Database struct {
	mtx sync.Mutex
	db  map[uint32]Message
}

func (d *Database) GetData(id uint32) (Message, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	m, ok := d.db[id]
	if !ok {
		return Message{}, ErrDatabaseNotFound
	}
	return m, nil
}

func (d *Database) PostData(m Message) error {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	_, ok := d.db[m.ID]
	if ok {
		return ErrDatabaseAlreadyExists
	}
	d.db[m.ID] = m
	return nil
}

func (d *Database) PutData(id uint32, m Message) error {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	_, ok := d.db[id]
	if !ok {
		return ErrDatabaseNotFound
	}
	d.db[id] = m
	return nil
}

func (d *Database) DeleteData(id uint32) error {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	_, ok := d.db[id]
	if !ok {
		return ErrDatabaseNotFound
	}
	delete(d.db, id)
	return nil
}

func (d *Database) WriteData(m Message) error {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	d.db[m.ID] = m
	return nil
}

var db = &Database{
	db: map[uint32]Message{},
}
