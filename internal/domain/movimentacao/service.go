package movimentacao

import (
	"errors"
	"time"
)

// Service contém a lógica de negócio do domínio de movimentação.
type Service struct {
	repo *Repository
}

// NewService cria o service com o repositório injetado.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ListarTodas retorna todas as movimentações registradas.
func (s *Service) ListarTodas() ([]Movimentacao, error) {
	return s.repo.ListarTodas()
}

// ListarPorItem retorna o histórico de movimentações de um item específico.
func (s *Service) ListarPorItem(itemID uint) ([]Movimentacao, error) {
	return s.repo.ListarPorItem(itemID)
}

// ListarPorProduto retorna o histórico de movimentações de todos os itens
// de um produto específico.
func (s *Service) ListarPorProduto(produtoID uint) ([]Movimentacao, error) {
	return s.repo.ListarPorProduto(produtoID)
}

// ListarPorTipo retorna movimentações filtradas por tipo.
// Tipos válidos: "entrada" ou "saida".
func (s *Service) ListarPorTipo(tipo string) ([]Movimentacao, error) {
	if tipo != "entrada" && tipo != "saida" {
		return nil, errors.New("tipo inválido — use 'entrada' ou 'saida'")
	}
	return s.repo.ListarPorTipo(tipo)
}

// ListarPorPeriodo retorna movimentações dentro de um intervalo de datas.
// Ambas as datas são obrigatórias e fim deve ser posterior a inicio.
func (s *Service) ListarPorPeriodo(inicio, fim time.Time) ([]Movimentacao, error) {
	if fim.Before(inicio) {
		return nil, errors.New("data de fim deve ser posterior à data de início")
	}
	return s.repo.ListarPorPeriodo(inicio, fim)
}