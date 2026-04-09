package estoque

import (
	"super-br/internal/domain/produto"
	"time"
)

type ItemEstoque struct {
	ID        uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	ProdutoID uint            `gorm:"not null" json:"produto_id"`
	Produto   produto.Produto `gorm:"foreignKey:ProdutoID" json:"produto"`
	CodLote   string          `gorm:"type:varchar(50);not null" json:"cod_lote"`
	Estado    string          `gorm:"type:varchar(20);not null;default:'disponivel'" json:"estado"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type Estoque struct {
	ID         uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	ProdutoID  uint            `gorm:"not null;unique" json:"produto_id"`
	Produto    produto.Produto `gorm:"foreignKey:ProdutoID" json:"produto"`
	QtdAtual   int             `gorm:"not null;default:0" json:"qtd_atual"`
	QtdPedido  int             `gorm:"not null;default:0" json:"qtd_pedido"`
	QtdTotal   int             `gorm:"not null;default:0" json:"qtd_total"`
	ValorTotal float64         `gorm:"type:decimal(10,2);not null;default:0" json:"valor_total"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}