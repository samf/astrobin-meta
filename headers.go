package main

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

const fitsBlockSize = 2880
const fitsCardSize = 80

// readFITSHeader reads the primary HDU header of a FITS file and returns its
// keyword/value pairs as strings. Values are stored verbatim (string values
// have their surrounding quotes stripped); numeric conversion happens later.
//
// FITS headers are a sequence of 80-byte ASCII "card images" packed into
// 2880-byte blocks, terminated by an END card.
func readFITSHeader(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	header := map[string]string{}
	block := make([]byte, fitsBlockSize)
	for {
		if _, err := io.ReadFull(f, block); err != nil {
			return nil, fmt.Errorf("reading FITS header: %w", err)
		}
		for i := 0; i < fitsBlockSize; i += fitsCardSize {
			card := block[i : i+fitsCardSize]
			keyword := strings.TrimRight(string(card[0:8]), " ")
			if keyword == "END" {
				return header, nil
			}
			if keyword == "" || keyword == "COMMENT" || keyword == "HISTORY" {
				continue
			}
			// A value card has "= " in columns 9-10 (0-indexed 8-9).
			if card[8] != '=' || card[9] != ' ' {
				continue
			}
			header[keyword] = parseFITSValue(card[10:])
		}
	}
}

// parseFITSValue extracts the value portion of a FITS value card (everything
// after the "= " indicator), stripping string quotes or trailing comments.
func parseFITSValue(b []byte) string {
	s := strings.TrimLeft(string(b), " ")
	if len(s) > 0 && s[0] == '\'' {
		// String value: read up to the closing quote.
		rest := s[1:]
		if end := strings.IndexByte(rest, '\''); end >= 0 {
			return strings.TrimSpace(rest[:end])
		}
		return strings.TrimSpace(rest)
	}
	// Numeric or logical value: strip any trailing comment.
	if idx := strings.IndexByte(s, '/'); idx >= 0 {
		s = s[:idx]
	}
	return strings.TrimSpace(s)
}

// readXISFHeader reads the FITS-equivalent keywords from a monolithic XISF
// file. Such files start with:
//
//	8 bytes  signature 'XISF0100'
//	4 bytes  uint32 LE header length
//	4 bytes  reserved (zero)
//	N bytes  XML header (UTF-8)
//
// FITS-equivalent keywords appear in the XML as:
//
//	<FITSKeyword name="FILTER" value="Ha" comment="..."/>
func readXISFHeader(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	signature := make([]byte, 8)
	if _, err := io.ReadFull(f, signature); err != nil {
		return nil, fmt.Errorf("reading XISF signature: %w", err)
	}
	if string(signature) != "XISF0100" {
		return nil, fmt.Errorf("not a valid monolithic XISF file: %s", path)
	}

	var lengthAndReserved [8]byte
	if _, err := io.ReadFull(f, lengthAndReserved[:]); err != nil {
		return nil, fmt.Errorf("reading XISF header length: %w", err)
	}
	headerLength := binary.LittleEndian.Uint32(lengthAndReserved[0:4])
	// bytes 4:8 are reserved (zero) and ignored.

	xmlBytes := make([]byte, headerLength)
	if _, err := io.ReadFull(f, xmlBytes); err != nil {
		return nil, fmt.Errorf("reading XISF XML header: %w", err)
	}

	// Scan for all FITSKeyword elements regardless of nesting or namespace.
	header := map[string]string{}
	dec := xml.NewDecoder(bytes.NewReader(xmlBytes))
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parsing XISF XML header: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok || start.Name.Local != "FITSKeyword" {
			continue
		}
		var name, value string
		for _, attr := range start.Attr {
			switch attr.Name.Local {
			case "name":
				name = attr.Value
			case "value":
				value = attr.Value
			}
		}
		if name == "" {
			continue
		}
		value = strings.TrimSpace(value)
		// XISF sometimes wraps string values in single quotes.
		if len(value) >= 2 && strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
			value = strings.TrimSpace(value[1 : len(value)-1])
		}
		header[name] = value
	}
	return header, nil
}
