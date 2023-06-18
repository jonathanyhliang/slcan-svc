package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tarm/serial"
)

var (
	ErrBackendPortOpen     = errors.New("Backend: port open error")
	ErrBackendPortClose    = errors.New("Backend: port close error")
	ErrBackendPortFlush    = errors.New("Backend: port flush error")
	ErrBackendSlcanInit    = errors.New("Backend: SLCAN initialise error")
	ErrBackendInvalidID    = errors.New("Backend: invalid ID")
	ErrBackendInvalidData  = errors.New("Backend: invalid data")
	ErrBackendInvalidFrame = errors.New("Backend: invalid frame")
	ErrBackendReboot       = errors.New("Backend: reboot failed")
	ErrBackendOnhold       = errors.New("Backend: on hold")
	ErrBackendMsgQueue     = errors.New("Backend: message queue ping failed")
)

type Backend interface {
	Handler(port string, baud int, url string) error
	GetMessage(id int) error
	PostMessage(m Message) error
	Reboot() error
	Unlock() error
}

type SlcanBackend struct {
	init chan bool
	ch   chan Message
	rst  chan bool
	hold sync.Mutex
}

func NewSlcanBackend() Backend {
	return &SlcanBackend{
		init: make(chan bool),
		ch:   make(chan Message),
		rst:  make(chan bool),
	}
}

func (b *SlcanBackend) Handler(port string, baud int, url string) error {
	var s *serial.Port
	var err error
	var sl []byte
	c := &serial.Config{Name: port, Baud: baud, ReadTimeout: time.Second}

	rlptr := 0
	rl := make([]byte, len("T1234567880123456789abcdef\r\x00"))
	rb := make([]byte, 1)

	s, err = serial.OpenPort(c)
	if err != nil {
		return ErrBackendPortOpen
	}

	err = s.Flush()
	if err != nil {
		return ErrBackendPortFlush
	}

	// Initialise SLCAN port
	_, err = s.Write([]byte("C\rO\r\x00"))
	if err != nil {
		return ErrBackendSlcanInit
	}

	for {
		select {
		case <-b.init:
			s, err = serial.OpenPort(c)
			if err != nil {
				return ErrBackendPortOpen
			}

			err = s.Flush()
			if err != nil {
				return ErrBackendPortFlush
			}

			// Initialise SLCAN port
			_, err = s.Write([]byte("C\rO\r\x00"))
			if err != nil {
				return ErrBackendSlcanInit
			}

		case m := <-b.ch:
			if sl, err = encapsSlcanFrame(m); err == nil {
				_, err = s.Write([]byte(sl))
			} else {
				return err
			}
		case <-b.rst:
			// Frontend receives "Reboot" request, prompt SLCAN device to reset
			if _, err = s.Write([]byte("bbbbbb\r\x00")); err != nil {
				return ErrBackendReboot
			}
			// Wait for SLCAN device to reboot
			time.Sleep(3 * time.Second)
			// SLCAN boots into MCUboot, prompt MCUboot to enter serial recovery mode
			if _, err := s.Write([]byte("bbbbbb")); err != nil {
				return ErrBackendReboot
			}
			// To prevent frontend requests from accessing serial backend
			b.hold.Lock()
			// Wait for any ongoing serial transactions to complete
			time.Sleep(3 * time.Second)
			// Close serial connection
			if err := s.Close(); err != nil {
				return ErrBackendPortOpen
			}
			// Ping MCUmgr service for firmware update
			if err := msgQueuePing(url); err != nil {
				return err
			}

			for !b.hold.TryLock() {
				time.Sleep(time.Second)
			}

			b.hold.Unlock()

		default:
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
						}
					}
				}
			} else {
			}
		}
	}
}

func (b *SlcanBackend) GetMessage(id int) error {
	if !b.hold.TryLock() {
		return ErrBackendOnhold
	}
	defer b.hold.Unlock()
	return nil
}

func (b *SlcanBackend) PostMessage(m Message) error {
	if !b.hold.TryLock() {
		return ErrBackendOnhold
	}
	defer b.hold.Unlock()
	b.ch <- m
	return nil
}

func (b *SlcanBackend) Reboot() error {
	if !b.hold.TryLock() {
		return ErrBackendOnhold
	}
	defer b.hold.Unlock()
	b.rst <- true
	return nil
}

func (b *SlcanBackend) Unlock() error {
	b.hold.Unlock()
	b.init <- true
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
		if err != nil || id >= 0x800 {
			return Message{}, ErrBackendInvalidID
		}
		m.ID = uint32(id)
		dlc = int(f[4] - '0')
		p = 5
	} else if f[0] == 'T' {
		id, err := strconv.ParseInt(string(f[1:9]), 16, 32)
		if err != nil || id >= 0x20000000 {
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

	if f[p+dlc*2] != byte('\r') {
		return Message{}, ErrBackendInvalidFrame
	}

	f = f[p : p+dlc*2]
	d := make([]byte, dlc)
	if _, err := hex.Decode(d, f); err != nil {
		return Message{}, ErrBackendInvalidFrame
	}
	m.Data = string(d)

	return m, nil
}

func msgQueuePing(url string) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return ErrBackendMsgQueue
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return ErrBackendMsgQueue
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"handover", // name
		false,      // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return ErrBackendMsgQueue
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(""),
		})
	if err != nil {
		return ErrBackendMsgQueue
	}
	return nil
}
