package main

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

type HTTPRequest struct {
	Method  string
	Path    string
	Proto   string
	Headers map[string]string
}

// ParseHTTPRequest lee los bytes del socket TCP y extrae los datos HTTP
func ParseHTTPRequest(r io.Reader) (*HTTPRequest, error) {
	reader := bufio.NewReader(r)
	
	// Leer la primera línea (p. ej., "GET /index.html HTTP/1.1")
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	parts := strings.Fields(strings.TrimSpace(requestLine))
	if len(parts) < 3 {
		return nil, errors.New("formato HTTP invalido")
	}

	req := &HTTPRequest{
		Method:  parts[0],
		Path:    parts[1],
		Proto:   parts[2],
		Headers: make(map[string]string),
	}

	// Leer cabeceras HTTP
	for {
		line, err := reader.ReadString('\n')
		if err != nil || line == "\r\n" || line == "\n" {
			break // Fin de las cabeceras
		}
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			key := strings.TrimSpace(headerParts[0])
			val := strings.TrimSpace(headerParts[1])
			req.Headers[key] = val
		}
	}

	return req, nil
}
