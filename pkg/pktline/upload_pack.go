package pktline

import (
	"bytes"
	"io"
	"strconv"
	"strings"
)

// UploadPackRequest represents a parsed git-upload-pack request.
type UploadPackRequest struct {
	// Wants contains the object IDs that the client wants.
	Wants []string
	// Haves contains the object IDs that the client already has.
	Haves []string
	// Capabilities contains the capabilities requested by the client.
	Capabilities []string
	// Done indicates whether the client sent a "done" message.
	Done bool
	// Shallow contains shallow commit boundaries.
	Shallow []string
	// Deepen is the depth for shallow clones.
	Deepen int
}

// ParseUploadPackRequest parses a git-upload-pack request from raw pkt-line data.
func ParseUploadPackRequest(r io.Reader) (*UploadPackRequest, error) {
	scanner := NewScanner(r)
	req := &UploadPackRequest{}

	for {
		line, err := scanner.Scan()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Flush packet
		if line == nil {
			continue
		}

		// Empty delimiter
		if len(line) == 0 {
			continue
		}

		// Remove trailing newline
		lineStr := strings.TrimSuffix(string(line), "\n")

		// Parse the line
		switch {
		case strings.HasPrefix(lineStr, "want "):
			// Format: "want <oid> [capabilities]" or "want <oid>"
			parts := strings.SplitN(lineStr[5:], " ", 2)
			if len(parts) > 0 {
				req.Wants = append(req.Wants, parts[0])
			}
			// Parse capabilities from first want line
			if len(parts) > 1 && len(req.Wants) == 1 {
				req.Capabilities = strings.Split(parts[1], " ")
			}

		case strings.HasPrefix(lineStr, "have "):
			oid := strings.TrimPrefix(lineStr, "have ")
			req.Haves = append(req.Haves, oid)

		case strings.HasPrefix(lineStr, "shallow "):
			oid := strings.TrimPrefix(lineStr, "shallow ")
			req.Shallow = append(req.Shallow, oid)

		case strings.HasPrefix(lineStr, "deepen "):
			// Parse deepen value
			req.Deepen = parseDeepen(lineStr[7:])

		case lineStr == "done":
			req.Done = true
		}
	}

	return req, nil
}

// EncodeUploadPackRequest encodes an UploadPackRequest back to pkt-line format.
func EncodeUploadPackRequest(req *UploadPackRequest) []byte {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// Write wants (first one includes capabilities)
	for i, want := range req.Wants {
		if i == 0 && len(req.Capabilities) > 0 {
			line := "want " + want + " " + strings.Join(req.Capabilities, " ") + "\n"
			_ = w.WriteLine([]byte(line))
		} else {
			_ = w.WriteLine([]byte("want " + want + "\n"))
		}
	}

	// Write shallow
	for _, shallow := range req.Shallow {
		_ = w.WriteLine([]byte("shallow " + shallow + "\n"))
	}

	// Write deepen if set
	if req.Deepen > 0 {
		_ = w.WriteLine([]byte("deepen " + strconv.Itoa(req.Deepen) + "\n"))
	}

	// Flush after wants
	_ = w.WriteFlush()

	// Write haves
	for _, have := range req.Haves {
		_ = w.WriteLine([]byte("have " + have + "\n"))
	}

	// Write done if set
	if req.Done {
		_ = w.WriteLine([]byte("done\n"))
	}

	// Final flush
	_ = w.WriteFlush()

	return buf.Bytes()
}

// parseDeepen parses a deepen value from a string.
func parseDeepen(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}
