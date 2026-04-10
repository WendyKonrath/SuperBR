// Package notificacao gerencia alertas e avisos gerados automaticamente
// pelo sistema em resposta a eventos como movimentações de estoque,
// estoque baixo e geração de relatórios.
package notificacao

import (
	"super-br/internal/domain/usuario"
	"time"
)

// TiposNotificacao define as constantes dos tipos de alertas suportados.
// Usar constantes evita strings mágicas espalhadas pelo código.
const (
	TipoEstoqueBaixo   = "estoque_baixo"
	TipoEntradaEstoque = "entrada_estoque"
	TipoSaidaEstoque   = "saida_estoque"
	TipoVendaRealizada = "venda_realizada"
	TipoRelatorioGerado = "relatorio_gerado"
)

// Notificacao representa um alerta gerado pelo sistema para um usuário.
// O campo Lida indica se o usuário já visualizou o alerta.
type Notificacao struct {
	ID        uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	UsuarioID uint            `gorm:"not null" json:"usuario_id"`
	Usuario   usuario.Usuario `gorm:"foreignKey:UsuarioID" json:"usuario"`
	Tipo      string          `gorm:"type:varchar(50);not null" json:"tipo"`
	Mensagem  string          `gorm:"type:text;not null" json:"mensagem"`
	Lida      bool            `gorm:"default:false" json:"lida"`
	CreatedAt time.Time       `json:"created_at"`
}