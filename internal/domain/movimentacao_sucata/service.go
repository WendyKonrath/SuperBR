package movimentacao_sucata

import (
	"errors"
	"time"
)

// Service contém a lógica de negócio do domínio de movimentação de sucata.
type Service struct {
	repo *Repository
}

// NewService cria o service com o repositório injetado.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ListarTodas retorna todas as movimentações de sucata registradas.
func (s *Service) ListarTodas() ([]MovimentacaoSucata, error) {
	return s.repo.ListarTodas()
}

// ListarPorSucata retorna o histórico de movimentações de um tipo de sucata específico.
func (s *Service) ListarPorSucata(sucataID uint) ([]MovimentacaoSucata, error) {
	return s.repo.ListarPorSucata(sucataID)
}

// ListarPorTipo retorna movimentações filtradas por tipo.
// Tipos válidos: "entrada_sucata" ou "saida_sucata".
func (s *Service) ListarPorTipo(tipo string) ([]MovimentacaoSucata, error) {
	if tipo != "entrada_sucata" && tipo != "saida_sucata" {
		return nil, errors.New("tipo inválido — use 'entrada_sucata' ou 'saida_sucata'")
	}
	return s.repo.ListarPorTipo(tipo)
}

// ListarPorPeriodo retorna movimentações dentro de um intervalo de datas.
func (s *Service) ListarPorPeriodo(inicio, fim time.Time) ([]MovimentacaoSucata, error) {
	if fim.Before(inicio) {
		return nil, errors.New("data de fim deve ser posterior à data de início")
	}
	return s.repo.ListarPorPeriodo(inicio, fim)
}