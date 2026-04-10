package produto

import "errors"

// Service contém a lógica de negócio do domínio de produtos.
type Service struct {
	repo *Repository
}

// NewService cria o service com o repositório injetado.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Criar cadastra um novo produto no catálogo.
// Impede a criação de produtos com a mesma combinação de nome+categoria.
func (s *Service) Criar(nome, categoria string, valorAtacado, valorVarejo float64) (*Produto, error) {
	if valorAtacado < 0 || valorVarejo < 0 {
		return nil, errors.New("valores não podem ser negativos")
	}

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

// BuscarPorID retorna um produto pelo seu ID.
func (s *Service) BuscarPorID(id uint) (*Produto, error) {
	p, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, errors.New("produto não encontrado")
	}
	return p, nil
}

// Listar retorna todos os produtos do catálogo.
func (s *Service) Listar() ([]Produto, error) {
	return s.repo.Listar()
}

// ListarPorCategoria retorna produtos filtrados por categoria (ex: "60Ah").
func (s *Service) ListarPorCategoria(categoria string) ([]Produto, error) {
	return s.repo.ListarPorCategoria(categoria)
}

// Atualizar modifica os dados de um produto existente.
// Impede colisão de nome+categoria com outro produto diferente.
func (s *Service) Atualizar(id uint, nome, categoria string, valorAtacado, valorVarejo float64) (*Produto, error) {
	p, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, errors.New("produto não encontrado")
	}

	if valorAtacado < 0 || valorVarejo < 0 {
		return nil, errors.New("valores não podem ser negativos")
	}

	existente, err := s.repo.BuscarPorNomeECategoria(nome, categoria)
	if err == nil && existente.ID != id {
		return nil, errors.New("já existe outro produto com esse nome e categoria")
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

// Deletar remove um produto do catálogo.
// Bloqueia a remoção se o produto ainda possuir itens no estoque.
func (s *Service) Deletar(id uint) error {
	_, err := s.repo.BuscarPorID(id)
	if err != nil {
		return errors.New("produto não encontrado")
	}

	temItens, err := s.repo.PossuiItensNoEstoque(id)
	if err != nil {
		return errors.New("erro ao verificar estoque do produto")
	}
	if temItens {
		return errors.New("não é possível excluir um produto que possui itens no estoque")
	}

	return s.repo.Deletar(id)
}