package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"

	_ "github.com/lib/pq"
)

type SecurityEvent struct {
	ClientIP  string
	Method    string
	Path      string
	IsBlocked bool
	Reason    string
}

type DBService struct {
	db        *sql.DB
	eventChan chan SecurityEvent
}

func NewDBService(connStr string) (*DBService, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	service := &DBService{
		db:        db,
		eventChan: make(chan SecurityEvent, 1000), // Búfer para 1,000 eventos
	}

	// Iniciar Worker en segundo plano (Goroutine)
	go service.startWorker()

	return service, nil
}

// LogEvent envía el evento al canal sin bloquear al cliente HTTP
func (s *DBService) LogEvent(event SecurityEvent) {
	select {
	case s.eventChan <- event:
	default:
		log.Println("⚠️ [DB ALERTA] Búfer lleno. Descartando registro.")
	}
}

func (s *DBService) startWorker() {
	for event := range s.eventChan {
		vec := generateSimpleEmbedding(event.Path)
		vecStr := fmt.Sprintf("[%f,%f,%f]", vec[0], vec[1], vec[2])

		query := `
			INSERT INTO security_logs (client_ip, method, path, is_blocked, reason, embedding)
			VALUES ($1, $2, $3, $4, $5, $6::vector)
		`
		_, err := s.db.Exec(query, event.ClientIP, event.Method, event.Path, event.IsBlocked, event.Reason, vecStr)
		if err != nil {
			log.Printf("❌ [DB ERROR] Error guardando evento: %v\n", err)
		} else {
			log.Printf("💾 [DB PERSISTENCIA] Evento guardado (IP: %s | Bloqueado: %t | Embedding: %s)\n", event.ClientIP, event.IsBlocked, vecStr)
		}
	}
}

// Genera un vector numérico simplificado [longitud, caracteres_especiales, puntaje_sospecha]
func generateSimpleEmbedding(text string) [3]float64 {
	lengthFactor := math.Min(float64(len(text))/100.0, 1.0)

	specials := 0
	for _, ch := range text {
		if ch == '\'' || ch == '"' || ch == '<' || ch == '>' || ch == '-' || ch == ';' {
			specials++
		}
	}
	specialFactor := math.Min(float64(specials)/10.0, 1.0)
	suspicionScore := (lengthFactor + specialFactor) / 2.0

	return [3]float64{lengthFactor, specialFactor, suspicionScore}
}
