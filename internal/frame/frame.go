package frame

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"time"
)

const (
	header uint16 = 0xAA55
	footer uint16 = 0x55AA
)

var ErrBadHeader = errors.New("frame: geçersiz başlık")

type Conn struct {
	conn    net.Conn
	timeout time.Duration
}

func New(conn net.Conn, timeout time.Duration) *Conn {
	return &Conn{conn: conn, timeout: timeout}
}

func (c *Conn) Write(body []byte) error {
	if len(body) > math.MaxUint16 {
		return fmt.Errorf("frame: paket çok büyük (%d bayt)", len(body))
	}
	out := make([]byte, 0, len(body)+6)
	out = binary.BigEndian.AppendUint16(out, header)
	out = binary.LittleEndian.AppendUint16(out, uint16(len(body)))
	out = append(out, body...)
	out = binary.BigEndian.AppendUint16(out, footer)

	if err := c.conn.SetWriteDeadline(time.Now().Add(c.timeout)); err != nil {
		return err
	}
	_, err := c.conn.Write(out)
	return err
}

func (c *Conn) Read() ([]byte, error) {
	if err := c.conn.SetReadDeadline(time.Now().Add(c.timeout)); err != nil {
		return nil, err
	}

	var head [4]byte
	if _, err := io.ReadFull(c.conn, head[:]); err != nil {
		return nil, err
	}
	if binary.BigEndian.Uint16(head[:2]) != header {
		return nil, ErrBadHeader
	}

	size := int(binary.LittleEndian.Uint16(head[2:]))
	rest := make([]byte, size+2)
	if _, err := io.ReadFull(c.conn, rest); err != nil {
		return nil, err
	}
	return rest[:size], nil
}
