// Package venda gerencia as vendas realizadas, seus itens e formas de pagamento.
package venda

import (
	"super-br/internal/domain/estoque"
	"super-br/internal/domain/usuario"
	"time"
)

const (
	StatusPendente  = "pendente"
	StatusConcluida = "concluida"
	StatusCancelada = "cancelada"
)

// Venda representa uma transação de venda realizada no sistema.
// O campo Observacoes é opcional e aparece no comprovante PDF.
type Venda struct {
	ID               uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	Data             time.Time       `gorm:"not null" json:"data"`
	NomeCliente      string          `gorm:"type:varchar(100);not null" json:"nome_cliente"`
	DocumentoCliente string          `gorm:"type:varchar(20)" json:"documento_cliente"`
	TelefoneCliente  string          `gorm:"type:varchar(20)" json:"telefone_cliente"`
	Observacoes      string          `gorm:"type:text" json:"observacoes"`
	ValorTotal       float64         `gorm:"type:decimal(10,2);not null" json:"valor_total"`
	Status           string          `gorm:"type:varchar(20);not null;default:'pendente'" json:"status"`
	UsuarioID        uint            `gorm:"not null" json:"usuario_id"`
	Usuario          usuario.Usuario `gorm:"foreignKey:UsuarioID" json:"usuario"`
	Itens            []ItemVenda     `gorm:"foreignKey:VendaID" json:"itens"`
	Pagamentos       []Pagamento     `gorm:"foreignKey:VendaID" json:"pagamentos"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// ItemVenda representa uma bateria individual vendida dentro de uma venda.
type ItemVenda struct {
	ID            uint                `gorm:"primaryKey;autoIncrement" json:"id"`
	VendaID       uint                `gorm:"not null" json:"venda_id"`
	Venda         Venda               `gorm:"foreignKey:VendaID" json:"-"`
	ItemEstoqueID uint                `gorm:"not null" json:"item_estoque_id"`
	ItemEstoque   estoque.ItemEstoque `gorm:"foreignKey:ItemEstoqueID" json:"item_estoque"`
	ValorUnitario float64             `gorm:"type:decimal(10,2);not null" json:"valor_unitario"`
	Quantidade    int                 `gorm:"not null;default:1" json:"quantidade"`
	CreatedAt     time.Time           `json:"created_at"`
}

// Pagamento representa uma forma de pagamento utilizada em uma venda.
// Tipos válidos: "pix", "credito", "debito", "dinheiro", "sucata".
type Pagamento struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	VendaID   uint      `gorm:"not null" json:"venda_id"`
	Venda     Venda     `gorm:"foreignKey:VendaID" json:"-"`
	Tipo      string    `gorm:"type:varchar(20);not null" json:"tipo"`
	Valor     float64   `gorm:"type:decimal(10,2);not null" json:"valor"`
	CreatedAt time.Time `json:"created_at"`
}