package main

import (
	"net/url"
	"regexp"
)

var (
	// Detección de SQL Injection (SQLi)
	sqliPattern = regexp.MustCompile(`(?i)(SELECT|INSERT|DELETE|DROP|UNION|ALTER|CREATE|TRUNCATE|--|\bOR\b\s+1=1)`)
	// Detección de Cross-Site Scripting (XSS)
	xssPattern = regexp.MustCompile(`(?i)(<script|javascript:|onload=|onerror=)`)
	// Detección de Path Traversal
	pathTraversalPattern = regexp.MustCompile(`(\.\./|\.\.\\)`)
)

type WAFResult struct {
	IsBlocked bool
	Reason    string
}

// InspectRequest inspecciona la petición HTTP buscando patrones de ataque
func InspectRequest(req *HTTPRequest) WAFResult {
	// Decodificar la URL (%20 -> espacios reales) antes de inspeccionar
	decodedPath, err := url.QueryUnescape(req.Path)
	if err != nil {
		decodedPath = req.Path
	}

	if sqliPattern.MatchString(decodedPath) {
		return WAFResult{IsBlocked: true, Reason: "Ataque detectado: SQL Injection"}
	}
	if xssPattern.MatchString(decodedPath) {
		return WAFResult{IsBlocked: true, Reason: "Ataque detectado: XSS (Cross-Site Scripting)"}
	}
	if pathTraversalPattern.MatchString(decodedPath) {
		return WAFResult{IsBlocked: true, Reason: "Ataque detectado: Path Traversal"}
	}

	return WAFResult{IsBlocked: false, Reason: "OK"}
}
