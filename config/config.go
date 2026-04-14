// Package config carrega as variáveis de ambiente do arquivo .env
// e as disponibiliza para o restante da aplicação de forma centralizada.
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config agrupa todas as configurações da aplicação.
// Os valores são lidos do arquivo .env na raiz do projeto.
type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	ServerPort         string
	JWTSecret          string
	EstoqueMinimo      int    // Limiar de estoque baixo para disparo de notificação
	PastaComprovantes  string // Diretório onde os PDFs de comprovante são salvos
	PastaRelatorios    string // Diretório onde os PDFs de relatório são salvos
}

// Load lê o arquivo .env e retorna a configuração populada.
// A aplicação encerra imediatamente se o arquivo não for encontrado
// ou se campos obrigatórios estiverem ausentes.
func Load() *Config {
	// godotenv.Load não sobrescreve variáveis já definidas no ambiente do SO,
	// o que permite rodar em containers sem arquivo .env.
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: arquivo .env não encontrado — usando variáveis do ambiente do sistema")
	}

	// Lê o limiar de estoque mínimo — padrão 3 se não configurado.
	estoqueMinimo := 3
	if val := os.Getenv("ESTOQUE_MINIMO"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			estoqueMinimo = n
		}
	}

	// Diretório de comprovantes — padrão ./storage/comprovantes
	pastaComprovantes := os.Getenv("PASTA_COMPROVANTES")
	if pastaComprovantes == "" {
		pastaComprovantes = "./storage/comprovantes"
	}

	// Diretório de relatórios — padrão ./storage/relatorios
	pastaRelatorios := os.Getenv("PASTA_RELATORIOS")
	if pastaRelatorios == "" {
		pastaRelatorios = "./storage/relatorios"
	}

	cfg := &Config{
		DBHost:            os.Getenv("DB_HOST"),
		DBPort:            os.Getenv("DB_PORT"),
		DBUser:            os.Getenv("DB_USER"),
		DBPassword:        os.Getenv("DB_PASSWORD"),
		DBName:            os.Getenv("DB_NAME"),
		ServerPort:        os.Getenv("SERVER_PORT"),
		JWTSecret:         os.Getenv("JWT_SECRET"),
		EstoqueMinimo:     estoqueMinimo,
		PastaComprovantes: pastaComprovantes,
		PastaRelatorios:   pastaRelatorios,
	}

	// Valida campos obrigatórios — falhar cedo evita erros obscuros depois.
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET não configurado — defina no .env")
	}
	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBName == "" {
		log.Fatal("Configurações de banco de dados incompletas no .env")
	}
	if cfg.ServerPort == "" {
		cfg.ServerPort = "8080"
	}

	return cfg
}