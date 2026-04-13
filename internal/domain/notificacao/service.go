package notificacao

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Service contém a lógica de negócio do domínio de notificações.
type Service struct {
	repo *Repository
}

// NewService cria o service com o repositório injetado.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// notificarAdmins envia uma notificação para todos os admins e superadmins ativos.
// Deve ser chamado dentro de uma transação existente para garantir atomicidade.
func (s *Service) notificarAdmins(tx *gorm.DB, tipo, mensagem string) error {
	adminIDs, err := s.repo.BuscarAdmins()
	if err != nil {
		return err
	}

	for _, adminID := range adminIDs {
		n := &Notificacao{
			UsuarioID: adminID,
			Tipo:      tipo,
			Mensagem:  mensagem,
			Lida:      false,
		}
		if err := s.repo.Criar(tx, n); err != nil {
			return err
		}
	}

	return nil
}

// NotificarEntradaEstoque dispara notificação quando uma bateria entra no estoque.
func (s *Service) NotificarEntradaEstoque(tx *gorm.DB, nomeProduto, codLote string) error {
	mensagem := fmt.Sprintf("Entrada no estoque: %s — Lote %s", nomeProduto, codLote)
	return s.notificarAdmins(tx, TipoEntradaEstoque, mensagem)
}

// NotificarSaidaEstoque dispara notificação quando uma bateria sai do estoque.
func (s *Service) NotificarSaidaEstoque(tx *gorm.DB, nomeProduto string, itemID uint) error {
	mensagem := fmt.Sprintf("Saída do estoque: %s — Item ID %d", nomeProduto, itemID)
	return s.notificarAdmins(tx, TipoSaidaEstoque, mensagem)
}

// NotificarEstoqueBaixo dispara notificação quando qtd_atual fica abaixo do limiar.
func (s *Service) NotificarEstoqueBaixo(tx *gorm.DB, nomeProduto string, qtdAtual, minimo int) error {
	mensagem := fmt.Sprintf(
		"Estoque baixo: %s — %d unidade(s) disponível(is) (mínimo: %d)",
		nomeProduto, qtdAtual, minimo,
	)
	return s.notificarAdmins(tx, TipoEstoqueBaixo, mensagem)
}

// NotificarVendaRealizada dispara notificação quando uma venda é confirmada.
func (s *Service) NotificarVendaRealizada(tx *gorm.DB, vendaID uint, nomeCliente string, valorTotal float64) error {
	mensagem := fmt.Sprintf(
		"Venda #%d confirmada — Cliente: %s — Total: R$ %.2f",
		vendaID, nomeCliente, valorTotal,
	)
	return s.notificarAdmins(tx, TipoVendaRealizada, mensagem)
}

// Listar retorna todas as notificações do usuário autenticado.
// Se apenasNaoLidas for true, retorna somente as não lidas.
func (s *Service) Listar(usuarioID uint, apenasNaoLidas bool) ([]Notificacao, error) {
	if apenasNaoLidas {
		return s.repo.ListarNaoLidasPorUsuario(usuarioID)
	}
	return s.repo.ListarPorUsuario(usuarioID)
}

// MarcarComoLida marca uma notificação específica como lida.
// Valida que a notificação pertence ao usuário autenticado.
func (s *Service) MarcarComoLida(id, usuarioID uint) error {
	if err := s.repo.MarcarComoLida(id, usuarioID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("notificação não encontrada")
		}
		return errors.New("erro ao marcar notificação como lida")
	}
	return nil
}

// MarcarTodasComoLidas marca todas as notificações do usuário como lidas.
func (s *Service) MarcarTodasComoLidas(usuarioID uint) error {
	return s.repo.MarcarTodasComoLidas(usuarioID)
}