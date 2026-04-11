// Package movimentacao registra todo histórico de entradas e saídas do estoque.
// Cada operação (entrada, saída, empréstimo, devolução) gera uma Movimentacao,
// garantindo rastreabilidade completa conforme exigido pelo documento de escopo.
package movimentacao

import (
	"super-br/internal/domain/usuario"
	"time"
)

// Movimentacao representa um evento de movimentação de um item no estoque.
// Tipos válidos: "entrada", "saida".
//
// O campo ItemID referencia item_estoques mas não carrega o struct diretamente
// para evitar importação circular com o pacote estoque.
// Os dados do item são retornados via ItemInfo, preenchido pelo repository
// com uma query separada ao listar movimentações.
type Movimentacao struct {
	ID        uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	ItemID    uint            `gorm:"not null" json:"item_id"`
	Tipo      string          `gorm:"type:varchar(10);not null" json:"tipo"`
	Data      time.Time       `gorm:"not null" json:"data"`
	UsuarioID uint            `gorm:"not null" json:"usuario_id"`
	Usuario   usuario.Usuario `gorm:"foreignKey:UsuarioID" json:"usuario"`
	CreatedAt time.Time       `json:"created_at"`

	// ItemInfo é preenchido pelo repository ao listar movimentações.
	// Não é uma coluna no banco — é montado via JOIN para exibir
	// o código de lote e o produto associado ao item.
	ItemInfo *ItemInfo `gorm:"-" json:"item,omitempty"`
}

// ItemInfo agrupa as informações do item e produto para exibição nas movimentações.
// É preenchido manualmente pelo repository via JOIN, sem depender do pacote estoque.
type ItemInfo struct {
	ItemID      uint    `json:"item_id"`
	CodLote     string  `json:"cod_lote"`
	Estado      string  `json:"estado"`
	ProdutoID   uint    `json:"produto_id"`
	NomeProduto string  `json:"nome_produto"`
	Categoria   string  `json:"categoria"`
}