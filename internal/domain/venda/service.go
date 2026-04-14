package venda

import (
	"errors"
	"fmt"
	"super-br/internal/domain/estoque"
	"super-br/internal/domain/movimentacao"
	"super-br/internal/domain/notificacao"
	"super-br/internal/domain/produto"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Service contém a lógica de negócio do domínio de vendas.
type Service struct {
	repo         *Repository
	estoqueRepo  *estoque.Repository
	produtoRepo  *produto.Repository
	movRepo      *movimentacao.Repository
	notifService *notificacao.Service
}

// NewService cria o service injetando todos os repositórios necessários.
func NewService(
	repo *Repository,
	estoqueRepo *estoque.Repository,
	produtoRepo *produto.Repository,
	movRepo *movimentacao.Repository,
	notifService *notificacao.Service,
) *Service {
	return &Service{
		repo:         repo,
		estoqueRepo:  estoqueRepo,
		produtoRepo:  produtoRepo,
		movRepo:      movRepo,
		notifService: notifService,
	}
}

// itemInput representa um item a ser incluído na venda.
type itemInput struct {
	ProdutoID uint
	TipoPreco string
}

// pagamentoInput representa uma forma de pagamento registrada na venda.
type pagamentoInput struct {
	Tipo  string
	Valor float64
}

// CriarVenda inicia uma nova venda com status "pendente".
// Reserva os itens de estoque com FOR UPDATE para evitar venda dupla.
func (s *Service) CriarVenda(
	nomeCliente, documentoCliente, telefoneCliente, observacoes string,
	itens []itemInput,
	pagamentos []pagamentoInput,
	usuarioID uint,
) (*Venda, error) {
	if len(itens) == 0 {
		return nil, errors.New("a venda deve conter ao menos um item")
	}

	var vendaCriada *Venda

	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var valorTotal float64
		var itensCriados []ItemVenda

		for _, input := range itens {
			p, err := s.produtoRepo.BuscarPorID(input.ProdutoID)
			if err != nil {
				return fmt.Errorf("produto não encontrado: id %d", input.ProdutoID)
			}

			var valorUnitario float64
			switch input.TipoPreco {
			case "varejo":
				valorUnitario = p.ValorVarejo
			case "atacado":
				valorUnitario = p.ValorAtacado
			default:
				return errors.New("tipo de preço inválido — use 'atacado' ou 'varejo'")
			}

			var itemEstoque estoque.ItemEstoque
			result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("produto_id = ? AND estado = ?", input.ProdutoID, "disponivel").
				First(&itemEstoque)
			if result.Error != nil {
				return errors.New("sem estoque disponível para o produto: " + p.Nome)
			}

			itemEstoque.Estado = "reservado"
			if err := tx.Save(&itemEstoque).Error; err != nil {
				return err
			}

			var resumo estoque.Estoque
			if err := tx.Where("produto_id = ?", input.ProdutoID).First(&resumo).Error; err != nil {
				return errors.New("resumo de estoque não encontrado")
			}
			if resumo.QtdAtual <= 0 {
				return errors.New("estoque insuficiente para: " + p.Nome)
			}
			resumo.QtdAtual--
			resumo.ValorTotal -= p.ValorAtacado
			if err := tx.Save(&resumo).Error; err != nil {
				return err
			}

			valorTotal += valorUnitario
			itensCriados = append(itensCriados, ItemVenda{
				ItemEstoqueID: itemEstoque.ID,
				ValorUnitario: valorUnitario,
				Quantidade:    1,
			})
		}

		novaVenda := &Venda{
			Data:             time.Now(),
			NomeCliente:      nomeCliente,
			DocumentoCliente: documentoCliente,
			TelefoneCliente:  telefoneCliente,
			Observacoes:      observacoes,
			ValorTotal:       valorTotal,
			Status:           StatusPendente,
			UsuarioID:        usuarioID,
		}
		if err := tx.Create(novaVenda).Error; err != nil {
			return err
		}

		for i := range itensCriados {
			itensCriados[i].VendaID = novaVenda.ID
			if err := tx.Create(&itensCriados[i]).Error; err != nil {
				return err
			}
		}

		for _, pg := range pagamentos {
			if pg.Valor <= 0 {
				return errors.New("valor de pagamento deve ser maior que zero")
			}
			pagamento := Pagamento{
				VendaID: novaVenda.ID,
				Tipo:    pg.Tipo,
				Valor:   pg.Valor,
			}
			if err := tx.Create(&pagamento).Error; err != nil {
				return err
			}
		}

		vendaCriada = novaVenda
		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.repo.BuscarPorID(vendaCriada.ID)
}

// ConfirmarVenda finaliza uma venda pendente, marcando os itens como "vendido".
func (s *Service) ConfirmarVenda(vendaID, usuarioID uint) (*Venda, error) {
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var v Venda
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Itens").
			First(&v, vendaID).Error; err != nil {
			return errors.New("venda não encontrada")
		}

		if v.Status != StatusPendente {
			return errors.New("somente vendas com status 'pendente' podem ser confirmadas")
		}

		for _, item := range v.Itens {
			if err := tx.Model(&estoque.ItemEstoque{}).
				Where("id = ?", item.ItemEstoqueID).
				Update("estado", "vendido").Error; err != nil {
				return err
			}

			if err := s.movRepo.Registrar(tx, item.ItemEstoqueID, usuarioID, "saida"); err != nil {
				return err
			}
		}

		v.Status = StatusConcluida
		if err := tx.Save(&v).Error; err != nil {
			return err
		}

		return s.notifService.NotificarVendaRealizada(tx, v.ID, v.NomeCliente, v.ValorTotal)
	})

	if err != nil {
		return nil, err
	}

	return s.repo.BuscarPorID(vendaID)
}

// CancelarVenda reverte uma venda pendente, devolvendo os itens ao estoque.
func (s *Service) CancelarVenda(vendaID, usuarioID uint) (*Venda, error) {
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var v Venda
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Itens").
			First(&v, vendaID).Error; err != nil {
			return errors.New("venda não encontrada")
		}

		if v.Status != StatusPendente {
			return errors.New("somente vendas com status 'pendente' podem ser canceladas")
		}

		for _, item := range v.Itens {
			var itemEstoque estoque.ItemEstoque
			if err := tx.First(&itemEstoque, item.ItemEstoqueID).Error; err != nil {
				return errors.New("item de estoque não encontrado ao cancelar")
			}

			itemEstoque.Estado = "disponivel"
			if err := tx.Save(&itemEstoque).Error; err != nil {
				return err
			}

			p, err := s.produtoRepo.BuscarPorID(itemEstoque.ProdutoID)
			if err != nil {
				return errors.New("produto não encontrado ao recalcular estoque")
			}

			var resumo estoque.Estoque
			if err := tx.Where("produto_id = ?", itemEstoque.ProdutoID).First(&resumo).Error; err != nil {
				return errors.New("resumo de estoque não encontrado")
			}
			resumo.QtdAtual++
			resumo.ValorTotal += p.ValorAtacado
			if err := tx.Save(&resumo).Error; err != nil {
				return err
			}

			if err := s.movRepo.Registrar(tx, item.ItemEstoqueID, usuarioID, "entrada"); err != nil {
				return err
			}
		}

		v.Status = StatusCancelada
		return tx.Save(&v).Error
	})

	if err != nil {
		return nil, err
	}

	return s.repo.BuscarPorID(vendaID)
}

// BuscarPorID retorna uma venda pelo ID com todos os relacionamentos carregados.
func (s *Service) BuscarPorID(id uint) (*Venda, error) {
	v, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, errors.New("venda não encontrada")
	}
	return v, nil
}

// ListarPorPeriodo retorna as vendas realizadas em um intervalo de datas.
func (s *Service) ListarPorPeriodo(inicio, fim time.Time) ([]Venda, error) {
	if fim.Before(inicio) {
		return nil, errors.New("data de fim deve ser posterior à data de início")
	}
	return s.repo.ListarPorPeriodo(inicio, fim)
}

// ListarPorStatus retorna as vendas filtradas pelo status informado.
func (s *Service) ListarPorStatus(status string) ([]Venda, error) {
	switch status {
	case StatusPendente, StatusConcluida, StatusCancelada:
	default:
		return nil, errors.New("status inválido — use 'pendente', 'concluida' ou 'cancelada'")
	}
	return s.repo.ListarPorStatus(status)
}