// Package relatorio gerencia a geração e o armazenamento de relatórios
// em PDF sobre estoque e vendas, conforme o UC05 do documento de escopo.
package relatorio

import (
	"super-br/internal/domain/usuario"
	"time"
)

// TiposRelatorio define os tipos de relatório suportados pelo sistema.
const (
	TipoEstoque = "estoque"
	TipoVendas  = "vendas"
	TipoAmbos   = "ambos"
)

// Relatorio representa um relatório gerado pelo sistema.
// O campo CaminhoArquivo armazena o path do PDF gerado no servidor.
type Relatorio struct {
	ID             uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	UsuarioID      uint            `gorm:"not null" json:"usuario_id"`
	Usuario        usuario.Usuario `gorm:"foreignKey:UsuarioID" json:"usuario"`
	Tipo           string          `gorm:"type:varchar(20);not null" json:"tipo"`
	PeriodoInicio  time.Time       `gorm:"not null" json:"periodo_inicio"`
	PeriodoFim     time.Time       `gorm:"not null" json:"periodo_fim"`
	DataGeracao    time.Time       `gorm:"not null" json:"data_geracao"`
	CaminhoArquivo string          `gorm:"type:varchar(255)" json:"caminho_arquivo"`
	CreatedAt      time.Time       `json:"created_at"`
}