package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	connStr := "host=localhost port=5432 user=postgres password=sentinelpass dbname=cloudsentinel_db sslmode=disable"
	dbService, err := NewDBService(connStr)
	if err != nil {
		log.Fatalf("❌ Error conectando a PostgreSQL: %v", err)
	}
	fmt.Println("🐘 Conexión exitosa con PostgreSQL + pgvector")

	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		fmt.Printf("Error al abrir socket TCP: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	ipCache := NewLRUCache(10000)
	rateLimiter := NewRedisRateLimiter("localhost:6379", 5, 10*time.Second)
	agentTools := NewAgentTools(dbService.db, rateLimiter)

	fmt.Println("🚀 CloudSentinel WAF + Agentic AI Engine activo en TCP :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error al aceptar conexión: %v\n", err)
			continue
		}

		go handleConnection(conn, ipCache, rateLimiter, dbService, agentTools)
	}
}

func handleConnection(conn net.Conn, cache *LRUCache, rateLimiter *RedisRateLimiter, dbService *DBService, agentTools *AgentTools) {
	defer conn.Close()

	clientIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	cache.AddOrUpdate(clientIP)

	req, err := ParseHTTPRequest(conn)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	// 1. Control de Tasa (Rate Limiting)
	ctx := context.Background()
	allowed, count, err := rateLimiter.AllowRequest(ctx, clientIP)
	if err != nil {
		fmt.Printf("[REDIS ERROR] %v\n", err)
	}

	if !allowed {
		fmt.Printf("🛑 [RATE LIMIT BLOCK] IP: %s | Peticiones: %d\n", clientIP, count)
		response := "HTTP/1.1 429 Too Many Requests\r\nContent-Type: text/plain\r\n\r\n[BLOQUEADO] Has superado el limite de peticiones.\n"
		conn.Write([]byte(response))

		dbService.LogEvent(SecurityEvent{
			ClientIP:  clientIP,
			Method:    req.Method,
			Path:      req.Path,
			IsBlocked: true,
			Reason:    "Rate Limit Excedido",
		})
		return
	}

	// 2. Inspección del WAF
	wafResult := InspectRequest(req)
	if wafResult.IsBlocked {
		fmt.Printf("🚨 [WAF BLOCK] IP: %s | Motivo: %s | Ruta: %s\n", clientIP, wafResult.Reason, req.Path)
		response := fmt.Sprintf("HTTP/1.1 403 Forbidden\r\nContent-Type: text/plain\r\n\r\n[WAF BLOQUEO] %s\n", wafResult.Reason)
		conn.Write([]byte(response))

		dbService.LogEvent(SecurityEvent{
			ClientIP:  clientIP,
			Method:    req.Method,
			Path:      req.Path,
			IsBlocked: true,
			Reason:    wafResult.Reason,
		})

		// --- AGENTE DE IA EN ACCIÓN ---
		go func() {
			time.Sleep(500 * time.Millisecond) // Esperar persistencia en DB
			fmt.Println("\n🤖 [AGENTE IA] Iniciando análisis automático de amenaza...")
			
			// 1. Consulta RAG en pgvector
			similar, _ := agentTools.SearchSimilarAttacks("[0.5,0.5,0.5]")
			for _, match := range similar {
				fmt.Printf("   🔍 [RAG MATCH] %s\n", match)
			}

			// 2. Invocación de herramienta de remediación
			resultado := agentTools.EmergencyBlockIP(clientIP, wafResult.Reason)
			fmt.Printf("   ⚡ [ACCIÓN AGÉNTICA] %s\n\n", resultado)
		}()

		return
	}

	// 3. Tráfico Limpio
	fmt.Printf("✅ [TRAFICO LIMPIO] IP: %s | Peticion #%d | %s %s\n", clientIP, count, req.Method, req.Path)
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\n[ACCESO CONCEDIDO] Peticion limpia recibida (#%d).\n", count)
	conn.Write([]byte(response))

	dbService.LogEvent(SecurityEvent{
		ClientIP:  clientIP,
		Method:    req.Method,
		Path:      req.Path,
		IsBlocked: false,
		Reason:    "Tráfico Limpio",
	})
}
