package packet

import (
	"encoding/binary"
	"errors"
)

var ErrShortPacket = errors.New("packet: veri beklenenden kısa")

type Writer struct {
	buf []byte
}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Byte(v byte) *Writer {
	w.buf = append(w.buf, v)
	return w
}

func (w *Writer) Uint16(v uint16) *Writer {
	w.buf = binary.LittleEndian.AppendUint16(w.buf, v)
	return w
}

func (w *Writer) Uint32(v uint32) *Writer {
	w.buf = binary.LittleEndian.AppendUint32(w.buf, v)
	return w
}

func (w *Writer) String(s string) *Writer {
	w.Uint16(uint16(len(s)))
	w.buf = append(w.buf, s...)
	return w
}

func (w *Writer) Bytes() []byte {
	return w.buf
}

type Reader struct {
	buf []byte
	err error
}

func NewReader(b []byte) *Reader {
	return &Reader{buf: b}
}

func (r *Reader) Err() error {
	return r.err
}

func (r *Reader) take(n int) []byte {
	if r.err != nil {
		return nil
	}
	if len(r.buf) < n {
		r.err = ErrShortPacket
		return nil
	}
	out := r.buf[:n]
	r.buf = r.buf[n:]
	return out
}

func (r *Reader) Byte() byte {
	b := r.take(1)
	if b == nil {
		return 0
	}
	return b[0]
}

func (r *Reader) Uint16() uint16 {
	b := r.take(2)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint16(b)
}

func (r *Reader) SkipString() {
	r.take(int(r.Uint16()))
}

func (r *Reader) String() string {
	n := r.Uint16()
	b := r.take(int(n))
	if b == nil {
		return ""
	}
	return string(b)
}
