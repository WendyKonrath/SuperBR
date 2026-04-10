// Package config carrega as variáveis de ambiente do arquivo .env
// e as disponibiliza para o restante da aplicação de forma centralizada.
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config agrupa todas as configurações da aplicação.
// Os valores são lidos do arquivo .env na raiz do projeto.
type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
	JWTSecret  string
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

	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		ServerPort: os.Getenv("SERVER_PORT"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
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