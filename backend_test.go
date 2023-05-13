package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncapsDataFrame(t *testing.T) {
	var m Message

	// valid message
	m = Message{0x7ff, "200rpm"}
	s, err := encapsSlcanFrame(m)
	assert.Equal(t, []byte("t7ff632303072706d\r\x00"), s)
	assert.Equal(t, nil, err)

	// id out of range
	m = Message{0x20000000, ""}
	s, err = encapsSlcanFrame(m)
	assert.Empty(t, s)
	assert.NotEqual(t, nil, err)

	// data length out of range, dlc > 8
	m = Message{0x7ff, "123456789"}
	s, err = encapsSlcanFrame(m)
	assert.Empty(t, s)
	assert.NotEqual(t, nil, err)
}

func TestDecapsDataFrame(t *testing.T) {
	var s string

	// valid frame
	s = "t123632303072706d\r"
	m, err := decapsSlcanFrame([]byte(s))
	assert.Equal(t, uint32(0x123), m.ID)
	assert.Equal(t, "200rpm", m.Data)
	assert.Equal(t, nil, err)

	s = "T12345678632303072706d\r"
	m, err = decapsSlcanFrame([]byte(s))
	assert.Equal(t, uint32(0x12345678), m.ID)
	assert.Equal(t, "200rpm", m.Data)
	assert.Equal(t, nil, err)

	// id out of range
	s = "t800632303072706d\r"
	m, err = decapsSlcanFrame([]byte(s))
	assert.Empty(t, m)
	assert.NotEqual(t, nil, err)

	s = "T20000000632303072706d\r"
	m, err = decapsSlcanFrame([]byte(s))
	assert.Empty(t, m)
	assert.NotEqual(t, nil, err)
}
