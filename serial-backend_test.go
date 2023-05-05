package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncapsDataFrame(t *testing.T) {
	var frame = slcanFrame{
		ID:   0x7ff,
		Data: "200rpm",
	}
	s, err := EncapsDataFrame(frame)
	assert.Equal(t, []byte("t7ff632303072706d\r\x00"), s)
	assert.Equal(t, nil, err)
}

func TestDecapsDataFrame(t *testing.T) {
	var s = "t123632303072706d\r"
	f, err := DecapsDataFrame([]byte(s))
	assert.Equal(t, uint32(0x123), f.ID)
	assert.Equal(t, "200rpm", f.Data)
	assert.Equal(t, nil, err)
}
