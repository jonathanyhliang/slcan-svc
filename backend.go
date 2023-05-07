package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/tarm/serial"
)

var (
	ErrBackendInvalidID    = errors.New("Backend: invalid ID")
	ErrBackendInvalidData  = errors.New("Backend: invalid data")
	ErrBackendInvalidFrame = errors.New("Backend: invalid frame")
)

type Backend interface {
	Handler(port string, baud int, timeout time.Duration) error
	GetMessage(id string) error
	PostMessage(m Message) error
}

type SlcanBackend struct {
	ch chan Message
}

func NewSlcanBackend() Backend {
	return &SlcanBackend{
		ch: make(chan Message),
	}
}

func (b *SlcanBackend) Handler(port string, baud int, timeout time.Duration) error {
	c := &serial.Config{Name: port, Baud: baud, ReadTimeout: timeout}
	s, err := serial.OpenPort(c)
	if err != nil {
	}

	err = s.Flush()
	if err != nil {

	}

	rlptr := 0
	rl := make([]byte, len("T1234567880123456789abcdef\r\x00"))

	// Initialise SLCAN port
	_, err = s.Write([]byte("C\rO\r\x00"))
	if err != nil {
		return err
	}

	for {
		select {
		case m := <-b.ch:
			if sl, err := encapsSlcanFrame(m); err == nil {
				_, err = s.Write([]byte(sl))
			} else {
				return err
			}
		default:
			rb := make([]byte, 1)
			if n, err := s.Read(rb); n > 0 && err == nil {
				rl[rlptr] = rb[0]
				rlptr += 1
				if rlptr >= len("T1234567880123456789abcdef\r\x00") {
					rlptr = 0
				} else {
					if rl[rlptr-1] == byte('\r') || rl[rlptr-1] == byte('\n') {
						rl[rlptr] = byte('\x00')
						rlptr = 0
						if m, err := decapsSlcanFrame(rl); err == nil {
							_ = db.WriteData(m)
						} else {
							return err
						}
					}
				}
			} else {
				return err
			}
		}
	}
}

func (b *SlcanBackend) GetMessage(id string) error {
	return nil
}

func (b *SlcanBackend) PostMessage(m Message) error {
	b.ch <- m
	return nil
}

func encapsSlcanFrame(m Message) ([]byte, error) {
	var s string

	// Determine slcan data frame prefix
	// Append slcan filter ID
	if m.ID <= 0x7ff {
		s += fmt.Sprintf("t%03x", m.ID)
	} else if m.ID <= 0x1fffffff {
		s += fmt.Sprintf("T%08x", m.ID)
	} else {
		return nil, ErrBackendInvalidID
	}

	// Determine and append slcan frame dlc
	dlc := len(m.Data)
	if dlc > 8 {
		return nil, ErrBackendInvalidData
	}

	// Append slcan frame dlc, data, and terminators
	s += fmt.Sprintf("%1x", dlc)
	s += fmt.Sprintf("%x", []byte(m.Data))
	s += "\r\x00"

	return []byte(s), nil
}

func decapsSlcanFrame(f []byte) (Message, error) {
	var m Message
	var dlc, p int

	if f[0] == 't' {
		id, err := strconv.ParseInt(string(f[1:4]), 16, 32)
		if err != nil {
			return Message{}, ErrBackendInvalidID
		}
		m.ID = uint32(id)
		dlc = int(f[4] - '0')
		p = 5
	} else if f[0] == 'T' {
		id, err := strconv.ParseInt(string(f[1:9]), 16, 32)
		if err != nil {
			return Message{}, ErrBackendInvalidID
		}
		m.ID = uint32(id)
		dlc = int(f[9] - '0')
		p = 10
	} else {

	}

	if dlc > 8 {
		return Message{}, ErrBackendInvalidData
	}

	l := len(f)
	if l != p+dlc*2 {
		return Message{}, ErrBackendInvalidData
	}

	if f[l] != byte('\r') {
		return Message{}, ErrBackendInvalidFrame
	}

	f = f[p:l]
	d := make([]byte, dlc)
	if _, err := hex.Decode(d, f); err != nil {
		return Message{}, ErrBackendInvalidData
	}
	m.Data = string(d)

	return m, nil
}

// func SerialReadLine(s *serial.Port) []byte {
// 	p := 0
// 	max := len("T1234567880123456789abcdef\r\x00")
// 	ln := make([]byte, max)
// 	for {
// 		rb := make([]byte, 1)
// 		if n, err := s.Read(rb); n > 0 && err == nil {
// 			fmt.Print(rb)
// 			ln[p] = rb[0]
// 			p += 1
// 			if p >= max {
// 				p = 0
// 			} else {
// 				if ln[p-1] == byte('\r') || ln[p-1] == byte('\n') {
// 					ln[p] = byte('\x00')
// 					return ln
// 				}
// 			}
// 		}
// 	}
// }
