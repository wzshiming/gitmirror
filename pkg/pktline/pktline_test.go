package pktline

import (
	"bytes"
	"io"
	"testing"
)

func TestScanner_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    [][]byte
		wantErr bool
	}{
		{
			name:  "flush packet",
			input: "0000",
			want:  [][]byte{nil},
		},
		{
			name:  "single line with correct length",
			input: "0008test",
			want:  [][]byte{[]byte("test")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewScanner(bytes.NewReader([]byte(tt.input)))
			var got [][]byte
			for {
				line, err := scanner.Scan()
				if err == io.EOF {
					break
				}
				if err != nil {
					if tt.wantErr {
						return
					}
					t.Fatalf("Scan() error = %v", err)
				}
				got = append(got, line)
			}
			if len(got) != len(tt.want) {
				t.Errorf("Scan() got %d lines, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if !bytes.Equal(got[i], tt.want[i]) {
					t.Errorf("Scan() line[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		name    string
		payload []byte
		want    string
	}{
		{
			name:    "simple line",
			payload: []byte("hello\n"),
			want:    "000ahello\n",
		},
		{
			name:    "empty payload",
			payload: []byte{},
			want:    "0000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Encode(tt.payload)
			if string(got) != tt.want {
				t.Errorf("Encode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriter_WriteLine(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	if err := w.WriteLine([]byte("test")); err != nil {
		t.Fatalf("WriteLine() error = %v", err)
	}

	if buf.String() != "0008test" {
		t.Errorf("WriteLine() wrote %q, want %q", buf.String(), "0008test")
	}
}

func TestWriter_WriteFlush(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	if err := w.WriteFlush(); err != nil {
		t.Fatalf("WriteFlush() error = %v", err)
	}

	if buf.String() != "0000" {
		t.Errorf("WriteFlush() wrote %q, want %q", buf.String(), "0000")
	}
}
