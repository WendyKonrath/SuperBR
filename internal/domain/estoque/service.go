package estoque

import (
	"errors"
	"super-br/internal/domain/movimentacao"
	"super-br/internal/domain/notificacao"
	"super-br/internal/domain/produto"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Service contém a lógica de negócio do domínio de estoque.
type Service struct {
	repo        *Repository
	produtoRepo *produto.Repository
	movRepo     *movimentacao.Repository
	notifService *notificacao.Service
	estoqueMinimo int
}

// NewService cria o service injetando os repositórios necessários.
func NewService(
	repo *Repository,
	produtoRepo *produto.Repository,
	movRepo *movimentacao.Repository,
	notifService *notificacao.Service,
	estoqueMinimo int,
) *Service {
	return &Service{
		repo:          repo,
		produtoRepo:   produtoRepo,
		movRepo:       movRepo,
		notifService:  notifService,
		estoqueMinimo: estoqueMinimo,
	}
}

// EntradaEstoque registra a chegada de uma nova bateria no estoque.
// Cria o ItemEstoque, atualiza o resumo Estoque, registra a Movimentacao
// e dispara notificação de entrada, tudo em uma única transação atômica.
func (s *Service) EntradaEstoque(produtoID uint, codLote string, usuarioID uint) (*ItemEstoque, error) {
	_, err := s.produtoRepo.BuscarPorID(produtoID)
	if err != nil {
		return nil, errors.New("produto não encontrado")
	}

	var novoItem *ItemEstoque

	err = s.repo.DB().Transaction(func(tx *gorm.DB) error {
		// 1. Cria o item individual no estoque.
		novoItem = &ItemEstoque{
			ProdutoID: produtoID,
			CodLote:   codLote,
			Estado:    "disponivel",
		}
		if err := tx.Create(novoItem).Error; err != nil {
			return err
		}

		// 2. Atualiza (ou cria) o resumo de estoque do produto.
		p, _ := s.produtoRepo.BuscarPorID(produtoID)

		var resumo Estoque
		result := tx.Where("produto_id = ?", produtoID).First(&resumo)
		if result.Error != nil {
			resumo = Estoque{
				ProdutoID:  produtoID,
				QtdAtual:   1,
				QtdPedido:  0,
				QtdTotal:   1,
				ValorTotal: p.ValorAtacado,
			}
			if err := tx.Create(&resumo).Error; err != nil {
				return err
			}
		} else {
			resumo.QtdAtual++
			resumo.QtdTotal++
			resumo.ValorTotal += p.ValorAtacado
			if err := tx.Save(&resumo).Error; err != nil {
				return err
			}
		}

		// 3. Registra a movimentação.
		if err := s.movRepo.Registrar(tx, novoItem.ID, usuarioID, "entrada"); err != nil {
			return err
		}

		// 4. Notifica entrada no estoque.
		return s.notifService.NotificarEntradaEstoque(tx, p.Nome, codLote)
	})

	if err != nil {
		return nil, errors.New("erro ao registrar entrada no estoque")
	}

	return novoItem, nil
}

// SaidaEstoque registra a saída de um item específico do estoque pelo seu ID.
func (s *Service) SaidaEstoque(itemID uint, usuarioID uint) (*ItemEstoque, error) {
	var itemAtualizado *ItemEstoque

	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var item ItemEstoque
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&item, itemID).Error; err != nil {
			return errors.New("item não encontrado")
		}

		if item.Estado != "disponivel" {
			return errors.New("item não está disponível para saída")
		}

		item.Estado = "indisponivel"
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		itemAtualizado = &item

		var resumo Estoque
		if err := tx.Where("produto_id = ?", item.ProdutoID).First(&resumo).Error; err != nil {
			return err
		}

		if resumo.QtdAtual <= 0 {
			return errors.New("estoque insuficiente")
		}

		p, _ := s.produtoRepo.BuscarPorID(item.ProdutoID)
		resumo.QtdAtual--
		resumo.QtdTotal--
		resumo.ValorTotal -= p.ValorAtacado
		if err := tx.Save(&resumo).Error; err != nil {
			return err
		}

		// Registra a movimentação.
		if err := s.movRepo.Registrar(tx, item.ID, usuarioID, "saida"); err != nil {
			return err
		}

		// Notifica saída do estoque.
		if err := s.notifService.NotificarSaidaEstoque(tx, p.Nome, item.ID); err != nil {
			return err
		}

		// Notifica estoque baixo se necessário.
		if resumo.QtdAtual <= s.estoqueMinimo {
			return s.notifService.NotificarEstoqueBaixo(tx, p.Nome, resumo.QtdAtual, s.estoqueMinimo)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return itemAtualizado, nil
}

// DevolverItem retorna ao estoque um item que estava indisponível (saída manual).
func (s *Service) DevolverItem(itemID uint, usuarioID uint) (*ItemEstoque, error) {
	var itemAtualizado *ItemEstoque

	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var item ItemEstoque
		if err := tx.First(&item, itemID).Error; err != nil {
			return errors.New("item não encontrado")
		}

		if item.Estado == "disponivel" {
			return errors.New("item já está disponível no estoque")
		}

		item.Estado = "disponivel"
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		itemAtualizado = &item

		var resumo Estoque
		if err := tx.Where("produto_id = ?", item.ProdutoID).First(&resumo).Error; err != nil {
			return err
		}

		p, _ := s.produtoRepo.BuscarPorID(item.ProdutoID)
		resumo.QtdAtual++
		resumo.QtdTotal++
		resumo.ValorTotal += p.ValorAtacado
		if err := tx.Save(&resumo).Error; err != nil {
			return err
		}

		return s.movRepo.Registrar(tx, item.ID, usuarioID, "entrada")
	})

	if err != nil {
		return nil, err
	}

	return itemAtualizado, nil
}

// EmprestarItem marca uma bateria como emprestada, saindo do estoque disponível.
func (s *Service) EmprestarItem(itemID uint, usuarioID uint) (*ItemEstoque, error) {
	var itemAtualizado *ItemEstoque

	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var item ItemEstoque
		if err := tx.First(&item, itemID).Error; err != nil {
			return errors.New("item não encontrado")
		}

		if item.Estado != "disponivel" {
			return errors.New("item não está disponível para empréstimo")
		}

		item.Estado = "emprestado"
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		itemAtualizado = &item

		var resumo Estoque
		if err := tx.Where("produto_id = ?", item.ProdutoID).First(&resumo).Error; err != nil {
			return err
		}

		if resumo.QtdAtual <= 0 {
			return errors.New("estoque insuficiente")
		}

		p, _ := s.produtoRepo.BuscarPorID(item.ProdutoID)
		resumo.QtdAtual--
		resumo.ValorTotal -= p.ValorAtacado
		if err := tx.Save(&resumo).Error; err != nil {
			return err
		}

		// Registra movimentação.
		if err := s.movRepo.Registrar(tx, item.ID, usuarioID, "saida"); err != nil {
			return err
		}

		// Notifica estoque baixo se necessário.
		if resumo.QtdAtual <= s.estoqueMinimo {
			return s.notifService.NotificarEstoqueBaixo(tx, p.Nome, resumo.QtdAtual, s.estoqueMinimo)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return itemAtualizado, nil
}

// DevolverEmprestimo retorna ao estoque um item que estava emprestado.
func (s *Service) DevolverEmprestimo(itemID uint, usuarioID uint) (*ItemEstoque, error) {
	var itemAtualizado *ItemEstoque

	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var item ItemEstoque
		if err := tx.First(&item, itemID).Error; err != nil {
			return errors.New("item não encontrado")
		}

		if item.Estado != "emprestado" {
			return errors.New("item não está registrado como emprestado")
		}

		item.Estado = "disponivel"
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		itemAtualizado = &item

		var resumo Estoque
		if err := tx.Where("produto_id = ?", item.ProdutoID).First(&resumo).Error; err != nil {
			return err
		}

		p, _ := s.produtoRepo.BuscarPorID(item.ProdutoID)
		resumo.QtdAtual++
		resumo.ValorTotal += p.ValorAtacado
		if err := tx.Save(&resumo).Error; err != nil {
			return err
		}

		return s.movRepo.Registrar(tx, item.ID, usuarioID, "entrada")
	})

	if err != nil {
		return nil, err
	}

	return itemAtualizado, nil
}

// RegistrarMovimentacao é um helper exposto para uso pelo domínio de Venda.
func (s *Service) RegistrarMovimentacao(tx *gorm.DB, itemID, usuarioID uint, tipo string) error {
	return s.movRepo.Registrar(tx, itemID, usuarioID, tipo)
}

// BuscarItemPorID retorna um item de estoque pelo ID.
func (s *Service) BuscarItemPorID(id uint) (*ItemEstoque, error) {
	item, err := s.repo.BuscarItemPorID(id)
	if err != nil {
		return nil, errors.New("item não encontrado")
	}
	return item, nil
}

// ListarItens retorna todos os itens de estoque.
func (s *Service) ListarItens() ([]ItemEstoque, error) {
	return s.repo.ListarItens()
}

// ListarItensPorProduto retorna itens de um produto específico.
func (s *Service) ListarItensPorProduto(produtoID uint) ([]ItemEstoque, error) {
	return s.repo.ListarItensPorProduto(produtoID)
}

// ListarItensPorEstado filtra itens pelo estado atual.
func (s *Service) ListarItensPorEstado(estado string) ([]ItemEstoque, error) {
	return s.repo.ListarItensPorEstado(estado)
}

// ListarEstoque retorna o resumo consolidado de estoque por produto.
func (s *Service) ListarEstoque() ([]Estoque, error) {
	return s.repo.ListarEstoque()
}

// BuscarEstoquePorProduto retorna o resumo de estoque de um produto específico.
func (s *Service) BuscarEstoquePorProduto(produtoID uint) (*Estoque, error) {
	e, err := s.repo.BuscarEstoquePorProduto(produtoID)
	if err != nil {
		return nil, errors.New("estoque não encontrado para esse produto")
	}
	return e, nil
}

// BuscarItemDisponivel localiza uma bateria disponível para venda.
func (s *Service) BuscarItemDisponivel(produtoID uint, tx *gorm.DB) (*ItemEstoque, error) {
	return s.repo.BuscarItemDisponivel(produtoID, tx)
}

// AtualizarResumoEstoque recalcula e persiste o resumo de estoque dentro de uma transação.
func (s *Service) AtualizarResumoEstoque(tx *gorm.DB, produtoID uint, deltaQtd int, deltaValor float64) error {
	var resumo Estoque
	if err := tx.Where("produto_id = ?", produtoID).First(&resumo).Error; err != nil {
		return errors.New("resumo de estoque não encontrado")
	}

	resumo.QtdAtual += deltaQtd
	resumo.QtdTotal += deltaQtd
	resumo.ValorTotal += deltaValor

	if resumo.QtdAtual < 0 {
		return errors.New("operação resultaria em estoque negativo")
	}

	return tx.Save(&resumo).Error
}

// AtualizarEstadoItem altera o estado de um item dentro de uma transação existente.
func (s *Service) AtualizarEstadoItem(tx *gorm.DB, itemID uint, novoEstado string) error {
	return tx.Model(&ItemEstoque{}).Where("id = ?", itemID).Update("estado", novoEstado).Error
}

// calcularValorTotal retorna o valor de atacado do produto para uso nos cálculos de resumo.
func (s *Service) calcularValorTotal(produtoID uint, qtd int) (float64, error) {
	p, err := s.produtoRepo.BuscarPorID(produtoID)
	if err != nil {
		return 0, errors.New("produto não encontrado ao calcular valor")
	}
	return p.ValorAtacado * float64(qtd), nil
}

// nowUTC retorna o tempo atual em UTC para padronização de timestamps.
func nowUTC() time.Time {
	return time.Now().UTC()
}