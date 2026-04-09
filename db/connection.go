package db

import (
	"fmt"
	"log"

	"super-br/config"
	"super-br/internal/domain/estoque"
	"super-br/internal/domain/movimentacao"
	"super-br/internal/domain/notificacao"
	"super-br/internal/domain/produto"
	"super-br/internal/domain/relatorio"
	"super-br/internal/domain/sucata"
	"super-br/internal/domain/usuario"
	"super-br/internal/domain/venda"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Erro ao conectar no banco:", err)
	}

	// Cria todas as tabelas automaticamente
	err = db.AutoMigrate(
		&usuario.Usuario{},
		&produto.Produto{},
		&estoque.ItemEstoque{},
		&estoque.Estoque{},
		&sucata.EstoqueSucata{},
		&movimentacao.Movimentacao{},
		&venda.Venda{},
		&venda.ItemVenda{},
		&venda.Pagamento{},
		&notificacao.Notificacao{},
		&relatorio.Relatorio{},
	)
	if err != nil {
		log.Fatal("Erro ao criar tabelas:", err)
	}

	fmt.Println("Banco de dados conectado e tabelas criadas!")
	return db
}