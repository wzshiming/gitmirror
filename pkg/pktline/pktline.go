// Package pktline implements the Git pkt-line format for reading and writing
// git protocol data.
// Reference: https://git-scm.com/docs/protocol-common#_pkt_line_format
package pktline

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

var (
	// ErrInvalidPktLine is returned when a pkt-line is invalid.
	ErrInvalidPktLine = errors.New("invalid pkt-line")
	// FlushPkt represents a flush packet (0000).
	FlushPkt = []byte("0000")
	// DelimPkt represents a delimiter packet (0001).
	DelimPkt = []byte("0001")
)

// Scanner reads pkt-line formatted data.
type Scanner struct {
	r   *bufio.Reader
	err error
}

// NewScanner creates a new pkt-line scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// Scan reads the next pkt-line and returns it.
// Returns nil for flush packets (0000).
// Returns io.EOF when no more data is available.
func (s *Scanner) Scan() ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}

	// Read 4-byte length prefix
	lenHex := make([]byte, 4)
	_, err := io.ReadFull(s.r, lenHex)
	if err != nil {
		s.err = err
		return nil, err
	}

	// Check for flush packet
	if bytes.Equal(lenHex, FlushPkt) {
		return nil, nil
	}

	// Check for delimiter packet
	if bytes.Equal(lenHex, DelimPkt) {
		return []byte{}, nil
	}

	// Decode length
	length, err := parseHexLength(lenHex)
	if err != nil {
		s.err = err
		return nil, err
	}

	if length < 4 {
		s.err = ErrInvalidPktLine
		return nil, ErrInvalidPktLine
	}

	// Read payload (length includes the 4-byte prefix)
	payload := make([]byte, length-4)
	_, err = io.ReadFull(s.r, payload)
	if err != nil {
		s.err = err
		return nil, err
	}

	return payload, nil
}

// Err returns the first error encountered during scanning.
func (s *Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

// Encode encodes a payload as a pkt-line.
func Encode(payload []byte) []byte {
	if len(payload) == 0 {
		return FlushPkt
	}
	length := len(payload) + 4
	return append([]byte(fmt.Sprintf("%04x", length)), payload...)
}

// EncodeFlush returns a flush packet.
func EncodeFlush() []byte {
	return FlushPkt
}

// EncodeDelim returns a delimiter packet.
func EncodeDelim() []byte {
	return DelimPkt
}

// Writer writes pkt-line formatted data.
type Writer struct {
	w io.Writer
}

// NewWriter creates a new pkt-line writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// WriteLine writes a single pkt-line.
func (w *Writer) WriteLine(data []byte) error {
	_, err := w.w.Write(Encode(data))
	return err
}

// WriteFlush writes a flush packet.
func (w *Writer) WriteFlush() error {
	_, err := w.w.Write(FlushPkt)
	return err
}

// parseHexLength parses a 4-character hex length string.
func parseHexLength(b []byte) (int, error) {
	dst := make([]byte, 2)
	_, err := hex.Decode(dst, b)
	if err != nil {
		return 0, ErrInvalidPktLine
	}
	return int(dst[0])<<8 | int(dst[1]), nil
}
