package usuario

import (
	"errors"
	"super-br/internal/auth"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Login(login, senha string) (string, bool, error) {
	u, err := s.repo.BuscarPorLogin(login)
	if err != nil {
		return "", false, errors.New("usuário não encontrado")
	}

	if !u.Ativo {
		return "", false, errors.New("usuário inativo")
	}

	if u.PrimeiroAcesso {
		return "", true, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Senha), []byte(senha)); err != nil {
		return "", false, errors.New("senha incorreta")
	}

	token, err := auth.GerarToken(u.ID, u.Login, u.Perfil)
	if err != nil {
		return "", false, errors.New("erro ao gerar token")
	}

	return token, false, nil
}

func (s *Service) PrimeiroAcesso(login, novaSenha string) (string, error) {
	u, err := s.repo.BuscarPorLogin(login)
	if err != nil {
		return "", errors.New("usuário não encontrado")
	}

	if !u.PrimeiroAcesso {
		return "", errors.New("usuário já realizou o primeiro acesso")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("erro ao criptografar senha")
	}

	u.Senha = string(hash)
	u.PrimeiroAcesso = false

	if err := s.repo.Atualizar(u); err != nil {
		return "", errors.New("erro ao salvar senha")
	}

	token, err := auth.GerarToken(u.ID, u.Login, u.Perfil)
	if err != nil {
		return "", errors.New("erro ao gerar token")
	}

	return token, nil
}

func (s *Service) Criar(nome, login, perfil string) (*Usuario, error) {
	_, err := s.repo.BuscarPorLogin(login)
	if err == nil {
		return nil, errors.New("login já está em uso")
	}

	u := &Usuario{
		Nome:           nome,
		Login:          login,
		Perfil:         perfil,
		PrimeiroAcesso: true,
		Ativo:          true,
	}

	if err := s.repo.Criar(u); err != nil {
		return nil, errors.New("erro ao criar usuário")
	}

	return u, nil
}

func (s *Service) Atualizar(id uint, nome, perfil string) (*Usuario, error) {
	u, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	// Não permite alterar o superadmin
	if u.Perfil == "superadmin" {
		return nil, errors.New("não é permitido alterar o super admin")
	}

	u.Nome = nome
	u.Perfil = perfil

	if err := s.repo.Atualizar(u); err != nil {
		return nil, errors.New("erro ao atualizar usuário")
	}

	return u, nil
}

func (s *Service) Desativar(id uint) error {
	u, err := s.repo.BuscarPorID(id)
	if err != nil {
		return errors.New("usuário não encontrado")
	}

	if u.Perfil == "superadmin" {
		return errors.New("não é permitido desativar o super admin")
	}

	u.Ativo = false

	return s.repo.Atualizar(u)
}

func (s *Service) ResetarSenha(id uint) error {
	u, err := s.repo.BuscarPorID(id)
	if err != nil {
		return errors.New("usuário não encontrado")
	}

	if u.Perfil == "superadmin" {
		return errors.New("não é permitido resetar a senha do super admin")
	}

	u.Senha = ""
	u.PrimeiroAcesso = true

	return s.repo.Atualizar(u)
}

func (s *Service) Me(id uint) (*Usuario, error) {
	u, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}
	return u, nil
}

func (s *Service) Listar() ([]Usuario, error) {
	return s.repo.Listar()
}

func (s *Service) Ativar(id uint) error {
	u, err := s.repo.BuscarPorID(id)
	if err != nil {
		return errors.New("usuário não encontrado")
	}

	if u.Perfil == "superadmin" {
		return errors.New("não é permitido alterar o super admin")
	}

	u.Ativo = true

	return s.repo.Atualizar(u)
}