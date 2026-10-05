package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

type AgentTools struct {
	db          *sql.DB
	rateLimiter *RedisRateLimiter
}

func NewAgentTools(db *sql.DB, rateLimiter *RedisRateLimiter) *AgentTools {
	return &AgentTools{
		db:          db,
		rateLimiter: rateLimiter,
	}
}

// Tool 1: RAG - Busca ataques similares en PostgreSQL usando distancia por coseno pgvector
func (t *AgentTools) SearchSimilarAttacks(queryVectorStr string) ([]string, error) {
	query := `
		SELECT path, reason, (embedding <=> $1::vector) AS distance
		FROM security_logs
		WHERE is_blocked = true
		ORDER BY distance ASC
		LIMIT 3;
	`
	rows, err := t.db.Query(query, queryVectorStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var path, reason string
		var distance float64
		if err := rows.Scan(&path, &reason, &distance); err == nil {
			results = append(results, fmt.Sprintf("Ataque similar en ruta '%s' (Motivo: %s | Distancia Vectorial: %.4f)", path, reason, distance))
		}
	}
	return results, nil
}

// Tool 2: Remedación Autónoma - Aplica un bloqueo estricto de IP en Redis por 1 hora
func (t *AgentTools) EmergencyBlockIP(ip string, reason string) string {
	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:%s", ip)

	// Forzar contador de Redis al máximo por 1 hora (3600s)
	t.rateLimiter.client.Set(ctx, key, 9999, 1*time.Hour)

	log.Printf("🤖 [AGENTE IA REMEDIACIÓN] IP %s bloqueada automáticamente en Redis. Motivo: %s", ip, reason)
	return fmt.Sprintf("ÉXITO: La IP %s ha sido aislada temporalmente durante 1 hora. Motivo registrado: %s", ip, reason)
}
