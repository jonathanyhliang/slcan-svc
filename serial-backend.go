package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/tarm/serial"
)

var ChanToSerialBackend = make(chan slcanFrame)

func SerialBackend() {
	rlptr := 0
	rl := make([]byte, len("T1234567880123456789abcdef\r\x00"))

	log.SetPrefix("slcan: ")
	log.SetFlags(0)
	c := &serial.Config{Name: "COM4", Baud: 115200, ReadTimeout: time.Second}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Flush()
	if err != nil {
		log.Fatal(err)
	}

	// Initialise SLCAN port
	_, err = s.Write([]byte("C\rO\r\x00"))

	for {
		select {
		case f := <-ChanToSerialBackend:
			if sl, err := EncapsDataFrame(f); err == nil {
				_, err = s.Write([]byte(sl))
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
						if f, err := DecapsDataFrame(rl); err == nil {
							_ = backendDB.WriteToSlcanDB(f)
						}
					}
				}
			}
		}
	}
}

func SerialReadLine(s *serial.Port) []byte {
	p := 0
	max := len("T1234567880123456789abcdef\r\x00")
	ln := make([]byte, max)
	for {
		rb := make([]byte, 1)
		if n, err := s.Read(rb); n > 0 && err == nil {
			fmt.Print(rb)
			ln[p] = rb[0]
			p += 1
			if p >= max {
				p = 0
			} else {
				if ln[p-1] == byte('\r') || ln[p-1] == byte('\n') {
					ln[p] = byte('\x00')
					return ln
				}
			}
		}
	}
}

func EncapsDataFrame(f slcanFrame) ([]byte, error) {
	var s string

	// Determine slcan data frame prefix
	// Append slcan filter ID
	if f.ID <= 0x7ff {
		s += fmt.Sprintf("t%03x", f.ID)
	} else if f.ID <= 0x1fffffff {
		s += fmt.Sprintf("T%08x", f.ID)
	} else {
		return nil, errors.New("frame id out of range")
	}

	// Determine and append slcan frame dlc
	dlc := len(f.Data)
	if dlc > 8 {
		return nil, errors.New("frame data out of range")
	}

	// Append slcan frame dlc, data, and terminators
	s += fmt.Sprintf("%1x", dlc)
	s += fmt.Sprintf("%x", []byte(f.Data))
	s += "\r\x00"

	return []byte(s), nil
}

func DecapsDataFrame(b []byte) (slcanFrame, error) {
	var f slcanFrame
	var dlc, p int

	if b[0] == 't' {
		id, err := strconv.ParseInt(string(b[1:4]), 16, 32)
		if err != nil {
			return slcanFrame{}, errors.New("invalid frame id")
		}
		f.ID = uint32(id)
		dlc = int(b[4] - '0')
		p = 5
	} else if b[0] == 'T' {
		id, err := strconv.ParseInt(string(b[1:9]), 16, 32)
		if err != nil {
			return slcanFrame{}, errors.New("invalid frame id")
		}
		f.ID = uint32(id)
		dlc = int(b[9] - '0')
		p = 10
	} else {

	}

	if dlc > 8 {
		return slcanFrame{}, errors.New("invalid frame dlc")
	}

	l := len(b)
	if l != p+dlc*2 {
		return slcanFrame{}, errors.New("invalid frame data length")
	}

	if b[l] != byte('\r') {
		return slcanFrame{}, errors.New("invalid frame termination")
	}

	b = b[p:l]
	d := make([]byte, dlc)
	if _, err := hex.Decode(d, b); err != nil {
		return slcanFrame{}, errors.New("invalid frame data")
	}
	f.Data = string(d)

	return f, nil
}
