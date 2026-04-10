package estoque

import (
	"errors"
	"super-br/internal/domain/produto"
	"time"

	"gorm.io/gorm"
)

type Service struct {
	repo        *Repository
	produtoRepo *produto.Repository
}

func NewService(repo *Repository, produtoRepo *produto.Repository) *Service {
	return &Service{repo: repo, produtoRepo: produtoRepo}
}

// EntradaEstoque registra a entrada de um novo item no estoque
func (s *Service) EntradaEstoque(produtoID uint, codLote string, usuarioID uint) (*ItemEstoque, error) {
	// Verifica se o produto existe
	_, err := s.produtoRepo.BuscarPorID(produtoID)
	if err != nil {
		return nil, errors.New("produto não encontrado")
	}

	// Inicia uma transação — garante que tudo acontece junto ou nada acontece
	var novoItem *ItemEstoque
	err = s.repo.DB().Transaction(func(tx *gorm.DB) error {
		// Cria o item no estoque
		novoItem = &ItemEstoque{
			ProdutoID: produtoID,
			CodLote:   codLote,
			Estado:    "disponivel",
		}
		if err := tx.Create(novoItem).Error; err != nil {
			return err
		}

		// Atualiza o resumo do estoque
		var estoqueResumo Estoque
		result := tx.Where("produto_id = ?", produtoID).First(&estoqueResumo)

		if result.Error != nil {
			// Não existe ainda — cria o resumo
			p, _ := s.produtoRepo.BuscarPorID(produtoID)
			estoqueResumo = Estoque{
				ProdutoID:  produtoID,
				QtdAtual:   1,
				QtdPedido:  0,
				QtdTotal:   1,
				ValorTotal: p.ValorAtacado,
			}
			if err := tx.Create(&estoqueResumo).Error; err != nil {
				return err
			}
		} else {
			// Já existe — atualiza
			p, _ := s.produtoRepo.BuscarPorID(produtoID)
			estoqueResumo.QtdAtual++
			estoqueResumo.QtdTotal++
			estoqueResumo.ValorTotal += p.ValorAtacado
			if err := tx.Save(&estoqueResumo).Error; err != nil {
				return err
			}
		}

		// Registra a movimentação
		mov := map[string]interface{}{
			"item_id":    novoItem.ID,
			"tipo":       "entrada",
			"data":       time.Now(),
			"usuario_id": usuarioID,
			"created_at": time.Now(),
		}
		if err := tx.Table("movimentacaos").Create(&mov).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, errors.New("erro ao registrar entrada no estoque")
	}

	return novoItem, nil
}

// SaidaEstoque registra a saída de um item do estoque
func (s *Service) SaidaEstoque(itemID uint, usuarioID uint) (*ItemEstoque, error) {
	var itemAtualizado *ItemEstoque

	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		// Busca o item com lock para evitar condição de corrida
		var item ItemEstoque
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&item, itemID).Error; err != nil {
			return errors.New("item não encontrado")
		}

		// Verifica se o item está disponível
		if item.Estado != "disponivel" {
			return errors.New("item não está disponível para saída")
		}

		// Atualiza o estado do item
		item.Estado = "indisponivel"
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		itemAtualizado = &item

		// Atualiza o resumo do estoque
		var estoqueResumo Estoque
		if err := tx.Where("produto_id = ?", item.ProdutoID).First(&estoqueResumo).Error; err != nil {
			return err
		}

		// Garante que o estoque nunca fica negativo
		if estoqueResumo.QtdAtual <= 0 {
			return errors.New("estoque insuficiente")
		}

		p, _ := s.produtoRepo.BuscarPorID(item.ProdutoID)
		estoqueResumo.QtdAtual--
		estoqueResumo.QtdTotal--
		estoqueResumo.ValorTotal -= p.ValorAtacado
		if err := tx.Save(&estoqueResumo).Error; err != nil {
			return err
		}

		// Registra a movimentação
		mov := map[string]interface{}{
			"item_id":    item.ID,
			"tipo":       "saida",
			"data":       time.Now(),
			"usuario_id": usuarioID,
			"created_at": time.Now(),
		}
		if err := tx.Table("movimentacaos").Create(&mov).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return itemAtualizado, nil
}

func (s *Service) BuscarItemPorID(id uint) (*ItemEstoque, error) {
	item, err := s.repo.BuscarItemPorID(id)
	if err != nil {
		return nil, errors.New("item não encontrado")
	}
	return item, nil
}

func (s *Service) ListarItens() ([]ItemEstoque, error) {
	return s.repo.ListarItens()
}

func (s *Service) ListarItensPorProduto(produtoID uint) ([]ItemEstoque, error) {
	return s.repo.ListarItensPorProduto(produtoID)
}

func (s *Service) ListarItensPorEstado(estado string) ([]ItemEstoque, error) {
	return s.repo.ListarItensPorEstado(estado)
}

func (s *Service) ListarEstoque() ([]Estoque, error) {
	return s.repo.ListarEstoque()
}

func (s *Service) BuscarEstoquePorProduto(produtoID uint) (*Estoque, error) {
	estoque, err := s.repo.BuscarEstoquePorProduto(produtoID)
	if err != nil {
		return nil, errors.New("estoque não encontrado para esse produto")
	}
	return estoque, nil
}

func (s *Service) DevolverItem(itemID uint, usuarioID uint) (*ItemEstoque, error) {
	var itemAtualizado *ItemEstoque

	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var item ItemEstoque
		if err := tx.First(&item, itemID).Error; err != nil {
			return errors.New("item não encontrado")
		}

		// Só pode devolver item que está indisponível
		if item.Estado == "disponivel" {
			return errors.New("item já está disponível no estoque")
		}

		// Volta o item para disponível
		item.Estado = "disponivel"
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		itemAtualizado = &item

		// Atualiza o resumo do estoque
		var estoqueResumo Estoque
		if err := tx.Where("produto_id = ?", item.ProdutoID).First(&estoqueResumo).Error; err != nil {
			return err
		}

		p, _ := s.produtoRepo.BuscarPorID(item.ProdutoID)
		estoqueResumo.QtdAtual++
		estoqueResumo.QtdTotal++
		estoqueResumo.ValorTotal += p.ValorAtacado
		if err := tx.Save(&estoqueResumo).Error; err != nil {
			return err
		}

		// Registra a movimentação de devolução
		mov := map[string]interface{}{
			"item_id":    item.ID,
			"tipo":       "entrada",
			"data":       time.Now(),
			"usuario_id": usuarioID,
			"created_at": time.Now(),
		}
		if err := tx.Table("movimentacaos").Create(&mov).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return itemAtualizado, nil
}

func (s *Service) EmprestarItem(itemID uint, usuarioID uint) (*ItemEstoque, error) {
	var itemAtualizado *ItemEstoque

	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var item ItemEstoque
		if err := tx.First(&item, itemID).Error; err != nil {
			return errors.New("item não encontrado")
		}

		// Só pode emprestar item disponível
		if item.Estado != "disponivel" {
			return errors.New("item não está disponível para empréstimo")
		}

		item.Estado = "emprestado"
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		itemAtualizado = &item

		// Atualiza o resumo — item emprestado sai do qtd_atual
		var estoqueResumo Estoque
		if err := tx.Where("produto_id = ?", item.ProdutoID).First(&estoqueResumo).Error; err != nil {
			return err
		}

		if estoqueResumo.QtdAtual <= 0 {
			return errors.New("estoque insuficiente")
		}

		p, _ := s.produtoRepo.BuscarPorID(item.ProdutoID)
		estoqueResumo.QtdAtual--
		estoqueResumo.ValorTotal -= p.ValorAtacado
		if err := tx.Save(&estoqueResumo).Error; err != nil {
			return err
		}

		// Registra a movimentação
		mov := map[string]interface{}{
			"item_id":    item.ID,
			"tipo":       "saida",
			"data":       time.Now(),
			"usuario_id": usuarioID,
			"created_at": time.Now(),
		}
		if err := tx.Table("movimentacaos").Create(&mov).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return itemAtualizado, nil
}

func (s *Service) DevolverEmprestimo(itemID uint, usuarioID uint) (*ItemEstoque, error) {
	var itemAtualizado *ItemEstoque

	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		var item ItemEstoque
		if err := tx.First(&item, itemID).Error; err != nil {
			return errors.New("item não encontrado")
		}

		// Só pode devolver item que está emprestado
		if item.Estado != "emprestado" {
			return errors.New("item não está emprestado")
		}

		item.Estado = "disponivel"
		if err := tx.Save(&item).Error; err != nil {
			return err
		}
		itemAtualizado = &item

		// Atualiza o resumo — item volta ao qtd_atual
		var estoqueResumo Estoque
		if err := tx.Where("produto_id = ?", item.ProdutoID).First(&estoqueResumo).Error; err != nil {
			return err
		}

		p, _ := s.produtoRepo.BuscarPorID(item.ProdutoID)
		estoqueResumo.QtdAtual++
		estoqueResumo.ValorTotal += p.ValorAtacado
		if err := tx.Save(&estoqueResumo).Error; err != nil {
			return err
		}

		// Registra a movimentação
		mov := map[string]interface{}{
			"item_id":    item.ID,
			"tipo":       "entrada",
			"data":       time.Now(),
			"usuario_id": usuarioID,
			"created_at": time.Now(),
		}
		if err := tx.Table("movimentacaos").Create(&mov).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return itemAtualizado, nil
}