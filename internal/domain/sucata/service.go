package sucata

import (
	"errors"
	movs "super-br/internal/domain/movimentacao_sucata"

	"gorm.io/gorm"
)

// Service contém a lógica de negócio do domínio de sucata.
type Service struct {
	repo    *Repository
	movRepo *movs.Repository
}

// NewService cria o service com os repositórios injetados.
func NewService(repo *Repository, movRepo *movs.Repository) *Service {
	return &Service{repo: repo, movRepo: movRepo}
}

// CadastrarTipo registra um novo tipo de sucata com peso e valor unitário.
// Obrigatório antes de dar qualquer entrada de sucata desse tipo.
func (s *Service) CadastrarTipo(tipoBateria string, pesoUnitario, valorUnitario float64) (*EstoqueSucata, error) {
	if pesoUnitario <= 0 {
		return nil, errors.New("peso unitário deve ser maior que zero")
	}
	if valorUnitario <= 0 {
		return nil, errors.New("valor unitário deve ser maior que zero")
	}

	_, err := s.repo.BuscarPorTipo(tipoBateria)
	if err == nil {
		return nil, errors.New("tipo de sucata já cadastrado")
	}

	sucata := &EstoqueSucata{
		TipoBateria:   tipoBateria,
		PesoUnitario:  pesoUnitario,
		ValorUnitario: valorUnitario,
		Qtd:           0,
		PesoTotal:     0,
		ValorTotal:    0,
	}

	if err := s.repo.Criar(sucata); err != nil {
		return nil, errors.New("erro ao cadastrar tipo de sucata")
	}

	return sucata, nil
}

// AtualizarTipo altera o peso e valor unitário de um tipo de sucata existente.
// Recalcula automaticamente os totais com base na quantidade atual em estoque.
func (s *Service) AtualizarTipo(id uint, pesoUnitario, valorUnitario float64) (*EstoqueSucata, error) {
	if pesoUnitario <= 0 {
		return nil, errors.New("peso unitário deve ser maior que zero")
	}
	if valorUnitario <= 0 {
		return nil, errors.New("valor unitário deve ser maior que zero")
	}

	sucata, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, errors.New("tipo de sucata não encontrado")
	}

	sucata.PesoUnitario = pesoUnitario
	sucata.ValorUnitario = valorUnitario
	sucata.PesoTotal = float64(sucata.Qtd) * pesoUnitario
	sucata.ValorTotal = float64(sucata.Qtd) * valorUnitario

	if err := s.repo.Atualizar(sucata); err != nil {
		return nil, errors.New("erro ao atualizar tipo de sucata")
	}

	return sucata, nil
}

// EntradaSucata registra a chegada de unidades de sucata de um tipo específico.
// Atualiza o estoque e registra a movimentação em transação atômica.
func (s *Service) EntradaSucata(tipoBateria string, qtd int, usuarioID uint) (*EstoqueSucata, error) {
	if qtd <= 0 {
		return nil, errors.New("quantidade deve ser maior que zero")
	}

	sucata, err := s.repo.BuscarPorTipo(tipoBateria)
	if err != nil {
		return nil, errors.New("tipo de sucata não encontrado — cadastre o tipo antes de dar entrada")
	}

	err = s.repo.DB().Transaction(func(tx *gorm.DB) error {
		sucata.Qtd += qtd
		sucata.PesoTotal = float64(sucata.Qtd) * sucata.PesoUnitario
		sucata.ValorTotal = float64(sucata.Qtd) * sucata.ValorUnitario

		if err := tx.Save(sucata).Error; err != nil {
			return err
		}

		return s.movRepo.Registrar(tx, sucata.ID, usuarioID, "entrada_sucata", qtd)
	})

	if err != nil {
		return nil, errors.New("erro ao registrar entrada de sucata")
	}

	return sucata, nil
}

// SaidaSucata registra a saída de unidades de sucata de um tipo específico.
// Garante que o estoque nunca fica negativo e registra movimentação em transação atômica.
func (s *Service) SaidaSucata(tipoBateria string, qtd int, usuarioID uint) (*EstoqueSucata, error) {
	if qtd <= 0 {
		return nil, errors.New("quantidade deve ser maior que zero")
	}

	sucata, err := s.repo.BuscarPorTipo(tipoBateria)
	if err != nil {
		return nil, errors.New("tipo de sucata não encontrado")
	}

	if sucata.Qtd < qtd {
		return nil, errors.New("quantidade solicitada maior que o estoque disponível")
	}

	err = s.repo.DB().Transaction(func(tx *gorm.DB) error {
		sucata.Qtd -= qtd
		sucata.PesoTotal = float64(sucata.Qtd) * sucata.PesoUnitario
		sucata.ValorTotal = float64(sucata.Qtd) * sucata.ValorUnitario

		if err := tx.Save(sucata).Error; err != nil {
			return err
		}

		return s.movRepo.Registrar(tx, sucata.ID, usuarioID, "saida_sucata", qtd)
	})

	if err != nil {
		return nil, errors.New("erro ao registrar saída de sucata")
	}

	return sucata, nil
}

// Listar retorna todos os tipos de sucata cadastrados com seus respectivos estoques.
func (s *Service) Listar() ([]EstoqueSucata, error) {
	return s.repo.Listar()
}

// BuscarPorTipo retorna o estoque de sucata de um tipo específico.
func (s *Service) BuscarPorTipo(tipoBateria string) (*EstoqueSucata, error) {
	sucata, err := s.repo.BuscarPorTipo(tipoBateria)
	if err != nil {
		return nil, errors.New("tipo de sucata não encontrado")
	}
	return sucata, nil
}