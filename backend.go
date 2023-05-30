package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
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
	GetMessage(id string) error
	PostMessage(m Message) error
	Reboot() error
}

type SlcanBackend struct {
	ch   chan Message
	rst  chan bool
	hold bool
}

func NewSlcanBackend() Backend {
	return &SlcanBackend{
		ch:   make(chan Message),
		rst:  make(chan bool),
		hold: false,
	}
}

func (b *SlcanBackend) Handler(port string, baud int, url string) error {
	c := &serial.Config{Name: port, Baud: baud, ReadTimeout: time.Second}
	s, err := serial.OpenPort(c)
	if err != nil {
		return ErrBackendPortOpen
	}

	err = s.Flush()
	if err != nil {
		return ErrBackendPortFlush
	}

	rlptr := 0
	rl := make([]byte, len("T1234567880123456789abcdef\r\x00"))

	// Initialise SLCAN port
	_, err = s.Write([]byte("C\rO\r\x00"))
	if err != nil {
		return ErrBackendSlcanInit
	}

	for {
		select {
		case m := <-b.ch:
			if sl, err := encapsSlcanFrame(m); err == nil {
				_, err = s.Write([]byte(sl))
			} else {
				return err
			}
		case <-b.rst:
			// Frontend receives "Reboot" request, prompt SLCAN device to reset
			if _, err := s.Write([]byte("bbbbbb\r\x00")); err != nil {
				return ErrBackendReboot
			}
			// Wait for SLCAN device to reboot
			time.Sleep(3 * time.Second)
			// SLCAN boots into MCUboot, prompt MCUboot to enter serial recovery mode
			if _, err := s.Write([]byte("bbbbbb")); err != nil {
				return ErrBackendReboot
			}
			// To prevent frontend requests from accessing serial backend
			b.hold = true
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
			// Wait for MCUmgr service to establish serial connection
			time.Sleep(10 * time.Second)
			// Attemp to recover serial connection once MCUmgr service handover
			for {
				s, err = serial.OpenPort(c)
				if err == nil {
					// Give SLCAN device 20s to boot + init
					time.Sleep(20 * time.Second)
					err = s.Flush()
					if err != nil {
						return ErrBackendPortFlush
					}
					_, err = s.Write([]byte("C\rO\r\x00"))
					if err != nil {
						return ErrBackendSlcanInit
					}
					break
				}
				time.Sleep(1 * time.Second)
			}
			// Recover requests from frontend
			b.hold = false

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
						}
					}
				}
			} else {
			}
		}
	}
}

func (b *SlcanBackend) GetMessage(id string) error {
	if b.hold == true {
		return ErrBackendOnhold
	}
	return nil
}

func (b *SlcanBackend) PostMessage(m Message) error {
	if b.hold == true {
		return ErrBackendOnhold
	}
	b.ch <- m
	return nil
}

func (b *SlcanBackend) Reboot() error {
	b.rst <- true
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
