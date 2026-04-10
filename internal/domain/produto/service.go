package produto

import "errors"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Criar(nome, categoria string, valorAtacado, valorVarejo float64) (*Produto, error) {
	if valorAtacado < 0 || valorVarejo < 0 {
		return nil, errors.New("valores não podem ser negativos")
	}

	// Verifica se já existe produto com mesmo nome e categoria
	_, err := s.repo.BuscarPorNomeECategoria(nome, categoria)
	if err == nil {
		return nil, errors.New("já existe um produto com esse nome e categoria")
	}

	p := &Produto{
		Nome:         nome,
		Categoria:    categoria,
		ValorAtacado: valorAtacado,
		ValorVarejo:  valorVarejo,
	}

	if err := s.repo.Criar(p); err != nil {
		return nil, errors.New("erro ao criar produto")
	}

	return p, nil
}

func (s *Service) BuscarPorID(id uint) (*Produto, error) {
	p, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, errors.New("produto não encontrado")
	}
	return p, nil
}

func (s *Service) Listar() ([]Produto, error) {
	return s.repo.Listar()
}

func (s *Service) ListarPorCategoria(categoria string) ([]Produto, error) {
	return s.repo.ListarPorCategoria(categoria)
}

func (s *Service) Atualizar(id uint, nome, categoria string, valorAtacado, valorVarejo float64) (*Produto, error) {
	p, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, errors.New("produto não encontrado")
	}

	if valorAtacado < 0 || valorVarejo < 0 {
		return nil, errors.New("valores não podem ser negativos")
	}

	// Verifica se já existe outro produto com mesmo nome e categoria
	existente, err := s.repo.BuscarPorNomeECategoria(nome, categoria)
	if err == nil && existente.ID != id {
		return nil, errors.New("já existe um produto com esse nome e categoria")
	}

	p.Nome = nome
	p.Categoria = categoria
	p.ValorAtacado = valorAtacado
	p.ValorVarejo = valorVarejo

	if err := s.repo.Atualizar(p); err != nil {
		return nil, errors.New("erro ao atualizar produto")
	}

	return p, nil
}

func (s *Service) Deletar(id uint) error {
	_, err := s.repo.BuscarPorID(id)
	if err != nil {
		return errors.New("produto não encontrado")
	}

	// Não permite deletar produto que tem itens no estoque
	temItens, err := s.repo.PossuiItensNoEstoque(id)
	if err != nil {
		return errors.New("erro ao verificar estoque do produto")
	}
	if temItens {
		return errors.New("não é possível deletar um produto que possui itens no estoque")
	}

	return s.repo.Deletar(id)
}