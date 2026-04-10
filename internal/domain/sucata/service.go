package sucata

import "errors"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) EntradaSucata(tipoBateria string, qtd int) (*EstoqueSucata, error) {
	if qtd <= 0 {
		return nil, errors.New("quantidade deve ser maior que zero")
	}

	sucata, err := s.repo.BuscarPorTipo(tipoBateria)

	if err != nil {
		// Não existe ainda — retorna erro pois o tipo precisa ser cadastrado antes
		return nil, errors.New("tipo de sucata não encontrado — cadastre o tipo antes de dar entrada")
	}

	// Atualiza os campos calculados
	sucata.Qtd += qtd
	sucata.PesoTotal = float64(sucata.Qtd) * sucata.PesoUnitario
	sucata.ValorTotal = float64(sucata.Qtd) * sucata.ValorUnitario

	if err := s.repo.Atualizar(sucata); err != nil {
		return nil, errors.New("erro ao registrar entrada de sucata")
	}

	return sucata, nil
}

func (s *Service) SaidaSucata(tipoBateria string, qtd int) (*EstoqueSucata, error) {
	if qtd <= 0 {
		return nil, errors.New("quantidade deve ser maior que zero")
	}

	sucata, err := s.repo.BuscarPorTipo(tipoBateria)
	if err != nil {
		return nil, errors.New("tipo de sucata não encontrado")
	}

	// Garante que o estoque nunca fica negativo
	if sucata.Qtd < qtd {
		return nil, errors.New("quantidade insuficiente em estoque")
	}

	sucata.Qtd -= qtd
	sucata.PesoTotal = float64(sucata.Qtd) * sucata.PesoUnitario
	sucata.ValorTotal = float64(sucata.Qtd) * sucata.ValorUnitario

	if err := s.repo.Atualizar(sucata); err != nil {
		return nil, errors.New("erro ao registrar saída de sucata")
	}

	return sucata, nil
}

func (s *Service) CadastrarTipo(tipoBateria string, pesoUnitario, valorUnitario float64) (*EstoqueSucata, error) {
	if pesoUnitario <= 0 || valorUnitario <= 0 {
		return nil, errors.New("peso e valor unitário devem ser maiores que zero")
	}

	// Verifica se já existe
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

func (s *Service) AtualizarTipo(id uint, pesoUnitario, valorUnitario float64) (*EstoqueSucata, error) {
	if pesoUnitario <= 0 || valorUnitario <= 0 {
		return nil, errors.New("peso e valor unitário devem ser maiores que zero")
	}

	sucata, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, errors.New("tipo de sucata não encontrado")
	}

	sucata.PesoUnitario = pesoUnitario
	sucata.ValorUnitario = valorUnitario

	// Recalcula os totais com os novos valores
	sucata.PesoTotal = float64(sucata.Qtd) * pesoUnitario
	sucata.ValorTotal = float64(sucata.Qtd) * valorUnitario

	if err := s.repo.Atualizar(sucata); err != nil {
		return nil, errors.New("erro ao atualizar tipo de sucata")
	}

	return sucata, nil
}

func (s *Service) Listar() ([]EstoqueSucata, error) {
	return s.repo.Listar()
}

func (s *Service) BuscarPorTipo(tipoBateria string) (*EstoqueSucata, error) {
	sucata, err := s.repo.BuscarPorTipo(tipoBateria)
	if err != nil {
		return nil, errors.New("tipo de sucata não encontrado")
	}
	return sucata, nil
}