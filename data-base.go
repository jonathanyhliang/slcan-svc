package main

import (
	"errors"
	"sync"
)

type slcanFrame struct {
	ID   uint32 `json:"id"`
	Data string `json:"data"`
}

type slcanDB struct {
	mutex sync.Mutex
	db    map[uint32]slcanFrame
}

func (c *slcanDB) ReadFromSlcanDB(id uint32) (slcanFrame, error) {
	var f slcanFrame
	c.mutex.Lock()
	f = c.db[id]
	c.mutex.Unlock()

	if f.ID != id {
		return f, errors.New("requested frame not found")
	}

	return f, nil
}

func (c *slcanDB) WriteToSlcanDB(f slcanFrame) error {
	c.mutex.Lock()
	c.db[f.ID] = f
	c.mutex.Unlock()

	return nil
}

func (c *slcanDB) UpdateToSlcanDB(f slcanFrame) error {
	var e slcanFrame
	c.mutex.Lock()
	e = c.db[f.ID]
	if e.ID != f.ID {
		return errors.New("requested frame not found")
	}
	c.db[f.ID] = f
	c.mutex.Unlock()

	return nil
}

func (c *slcanDB) RemoveFromSlcanDB(id uint32) error {
	var f slcanFrame
	c.mutex.Lock()
	f = c.db[id]
	if f.ID != id {
		return errors.New("requested frame not found")
	}
	delete(c.db, f.ID)
	c.mutex.Unlock()

	return nil
}

func (c *slcanDB) WriteToSerialBackend(f slcanFrame) error {
	ChanToSerialBackend <- f
	return nil
}

var backendDB = &slcanDB{
	db: make(map[uint32]slcanFrame),
}
