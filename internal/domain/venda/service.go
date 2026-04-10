package venda

import (
	"errors"
	"fmt"
	"super-br/internal/domain/estoque"
	"super-br/internal/domain/movimentacao"
	"super-br/internal/domain/produto"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Service contém a lógica de negócio do domínio de vendas.
type Service struct {
	repo        *Repository
	estoqueRepo *estoque.Repository
	produtoRepo *produto.Repository
	movRepo     *movimentacao.Repository
}

// NewService cria o service injetando todos os repositórios necessários.
func NewService(
	repo *Repository,
	estoqueRepo *estoque.Repository,
	produtoRepo *produto.Repository,
	movRepo *movimentacao.Repository,
) *Service {
	return &Service{
		repo:        repo,
		estoqueRepo: estoqueRepo,
		produtoRepo: produtoRepo,
		movRepo:     movRepo,
	}
}

// itemInput representa um item a ser incluído na venda,
// com o ID do produto e o tipo de preço desejado (atacado ou varejo).
type itemInput struct {
	ProdutoID  uint
	TipoPreco  string // "atacado" ou "varejo"
}

// pagamentoInput representa uma forma de pagamento usada na venda.
type pagamentoInput struct {
	Tipo  string
	Valor float64
}

// CriarVenda inicia uma nova venda com status "pendente".
// Reserva os itens de estoque (estado "reservado") para evitar venda dupla,
// calcula o valor total e registra os pagamentos informados.
//
// Toda a operação acontece em uma única transação atômica:
// se qualquer passo falhar, nenhuma alteração é persistida.
func (s *Service) CriarVenda(
	nomeCliente, documentoCliente, telefoneCliente string,
	itens []itemInput,
	pagamentos []pagamentoInput,
	usuarioID uint,
) (*Venda, error) {
	// Valida que a venda tem ao menos um item (regra de negócio do documento).
	if len(itens) == 0 {
		return nil, errors.New("a venda deve conter ao menos um item")
	}
	if len(pagamentos) == 0 {
		return nil, errors.New("informe ao menos uma forma de pagamento")
	}

	var vendaCriada *Venda

	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var valorTotal float64
		var itensCriados []ItemVenda

		// ── Passo 1: reservar cada item do estoque ────────────────────────────
		// Usamos FOR UPDATE para garantir que nenhum outro usuário consiga
		// reservar o mesmo item simultaneamente (race condition).
		for _, input := range itens {
			// Busca o produto para obter o preço correto.
			p, err := s.produtoRepo.BuscarPorID(input.ProdutoID)
			if err != nil {
				return fmt.Errorf("produto não encontrado: id %d", input.ProdutoID)
			}

			// Determina o valor unitário conforme o tipo de preço solicitado.
			var valorUnitario float64
			switch input.TipoPreco {
			case "varejo":
				valorUnitario = p.ValorVarejo
			case "atacado":
				valorUnitario = p.ValorAtacado
			default:
				return errors.New("tipo de preço inválido — use 'atacado' ou 'varejo'")
			}

			// Busca e trava o primeiro item disponível do produto com FOR UPDATE.
			var itemEstoque estoque.ItemEstoque
			result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("produto_id = ? AND estado = ?", input.ProdutoID, "disponivel").
				First(&itemEstoque)
			if result.Error != nil {
				return errors.New("sem estoque disponível para o produto: " + p.Nome)
			}

			// Marca o item como reservado — impede nova venda do mesmo item.
			itemEstoque.Estado = "reservado"
			if err := tx.Save(&itemEstoque).Error; err != nil {
				return err
			}

			// Atualiza o resumo de estoque: reduz QtdAtual.
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

		// ── Passo 2: validar que os pagamentos cobrem o valor total ───────────
		var totalPago float64
		for _, pg := range pagamentos {
			if pg.Valor <= 0 {
				return errors.New("valor de pagamento deve ser maior que zero")
			}
			totalPago += pg.Valor
		}
		// Tolerância de 1 centavo para erros de arredondamento de float.
		if totalPago < valorTotal-0.01 {
			return errors.New("valor total dos pagamentos é menor que o valor da venda")
		}

		// ── Passo 3: criar a venda ─────────────────────────────────────────────
		novaVenda := &Venda{
			Data:             time.Now(),
			NomeCliente:      nomeCliente,
			DocumentoCliente: documentoCliente,
			TelefoneCliente:  telefoneCliente,
			ValorTotal:       valorTotal,
			Status:           StatusPendente,
			UsuarioID:        usuarioID,
		}
		if err := tx.Create(novaVenda).Error; err != nil {
			return err
		}

		// ── Passo 4: criar os itens de venda vinculados ───────────────────────
		for i := range itensCriados {
			itensCriados[i].VendaID = novaVenda.ID
			if err := tx.Create(&itensCriados[i]).Error; err != nil {
				return err
			}
		}

		// ── Passo 5: criar os pagamentos ──────────────────────────────────────
		for _, pg := range pagamentos {
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

	// Recarrega a venda com todos os relacionamentos para retornar completa.
	return s.repo.BuscarPorID(vendaCriada.ID)
}

// ConfirmarVenda finaliza uma venda pendente, marcando os itens como "vendido"
// e registrando a movimentação de saída para cada bateria vendida.
//
// Fluxo:
//  1. Verifica que a venda existe e está pendente
//  2. Marca cada item de estoque como "vendido"
//  3. Registra movimentação de saída para cada item
//  4. Atualiza status da venda para "concluida"
func (s *Service) ConfirmarVenda(vendaID, usuarioID uint) (*Venda, error) {
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		// Busca e trava a venda para evitar confirmação simultânea.
		var v Venda
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Itens").
			First(&v, vendaID).Error; err != nil {
			return errors.New("venda não encontrada")
		}

		if v.Status != StatusPendente {
			return errors.New("somente vendas com status 'pendente' podem ser confirmadas")
		}

		// Marca cada item como "vendido" e registra a movimentação.
		for _, item := range v.Itens {
			if err := tx.Model(&estoque.ItemEstoque{}).
				Where("id = ?", item.ItemEstoqueID).
				Update("estado", "vendido").Error; err != nil {
				return err
			}

			// Registra a saída no histórico de movimentações.
			if err := s.movRepo.Registrar(tx, item.ItemEstoqueID, usuarioID, "saida"); err != nil {
				return err
			}
		}

		// Atualiza o status da venda.
		v.Status = StatusConcluida
		return tx.Save(&v).Error
	})

	if err != nil {
		return nil, err
	}

	return s.repo.BuscarPorID(vendaID)
}

// CancelarVenda reverte uma venda pendente, devolvendo os itens ao estoque disponível.
//
// Fluxo:
//  1. Verifica que a venda existe e está pendente
//  2. Devolve cada item de estoque para estado "disponivel"
//  3. Recalcula o resumo de estoque (incrementa QtdAtual)
//  4. Registra movimentação de entrada (devolução) para cada item
//  5. Atualiza status da venda para "cancelada"
//
// Somente vendas pendentes podem ser canceladas — vendas concluídas
// exigem um processo de devolução separado (a implementar futuramente).
func (s *Service) CancelarVenda(vendaID, usuarioID uint) (*Venda, error) {
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		// Busca e trava a venda.
		var v Venda
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Itens").
			First(&v, vendaID).Error; err != nil {
			return errors.New("venda não encontrada")
		}

		if v.Status != StatusPendente {
			return errors.New("somente vendas com status 'pendente' podem ser canceladas")
		}

		// Devolve cada item ao estoque e recalcula o resumo.
		for _, item := range v.Itens {
			// Busca o item de estoque para saber o produto.
			var itemEstoque estoque.ItemEstoque
			if err := tx.First(&itemEstoque, item.ItemEstoqueID).Error; err != nil {
				return errors.New("item de estoque não encontrado ao cancelar")
			}

			// Volta o item para disponível.
			itemEstoque.Estado = "disponivel"
			if err := tx.Save(&itemEstoque).Error; err != nil {
				return err
			}

			// Recalcula o resumo: incrementa QtdAtual e ValorTotal.
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

			// Registra a devolução no histórico de movimentações.
			if err := s.movRepo.Registrar(tx, item.ItemEstoqueID, usuarioID, "entrada"); err != nil {
				return err
			}
		}

		// Atualiza o status da venda.
		v.Status = StatusCancelada
		return tx.Save(&v).Error
	})

	if err != nil {
		return nil, err
	}

	return s.repo.BuscarPorID(vendaID)
}

// BuscarPorID retorna uma venda pelo seu ID com todos os relacionamentos carregados.
func (s *Service) BuscarPorID(id uint) (*Venda, error) {
	v, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, errors.New("venda não encontrada")
	}
	return v, nil
}

// ListarPorPeriodo retorna as vendas realizadas em um intervalo de datas.
// Utilizado pelo domínio de relatório para gerar relatórios diários e mensais.
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
		// status válido
	default:
		return nil, errors.New("status inválido — use 'pendente', 'concluida' ou 'cancelada'")
	}
	return s.repo.ListarPorStatus(status)
}