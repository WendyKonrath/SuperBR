package usuario

import (
	"errors"
	"super-br/internal/auth"

	"golang.org/x/crypto/bcrypt"
)

// custo de hashing para bcrypt. Valor 12 é recomendado para produção;
// DefaultCost (10) é aceitável mas 12 adiciona resistência extra a brute-force.
const bcryptCost = 12

// Service contém a lógica de negócio do domínio de usuários.
type Service struct {
	repo      *Repository
	jwtSecret string
}

// NewService cria o service injetando o repositório e o secret JWT.
// O secret deve vir de config.Config.JWTSecret, nunca de os.Getenv diretamente.
func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret}
}

// Login valida as credenciais e retorna um token JWT em caso de sucesso.
// Retorna (token, primeiroAcesso, erro).
// Se primeiroAcesso for true, o token virá vazio — o usuário deve definir sua senha antes.
func (s *Service) Login(login, senha string) (string, bool, error) {
	u, err := s.repo.BuscarPorLogin(login)
	if err != nil {
		// Mensagem genérica para não revelar se o login existe ou não.
		return "", false, errors.New("credenciais inválidas")
	}

	if !u.Ativo {
		return "", false, errors.New("usuário inativo — contate o administrador")
	}

	// Usuário no primeiro acesso não tem senha definida ainda.
	if u.PrimeiroAcesso {
		return "", true, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Senha), []byte(senha)); err != nil {
		return "", false, errors.New("credenciais inválidas")
	}

	token, err := auth.GerarToken(u.ID, u.Login, u.Perfil, s.jwtSecret)
	if err != nil {
		return "", false, errors.New("erro interno ao gerar token")
	}

	return token, false, nil
}

// PrimeiroAcesso define a senha inicial do usuário e retorna um token JWT.
// Só pode ser chamado uma vez por usuário — após isso, PrimeiroAcesso fica false.
func (s *Service) PrimeiroAcesso(login, novaSenha string) (string, error) {
	u, err := s.repo.BuscarPorLogin(login)
	if err != nil {
		return "", errors.New("usuário não encontrado")
	}

	if !u.PrimeiroAcesso {
		return "", errors.New("usuário já realizou o primeiro acesso")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(novaSenha), bcryptCost)
	if err != nil {
		return "", errors.New("erro ao processar senha")
	}

	u.Senha = string(hash)
	u.PrimeiroAcesso = false

	if err := s.repo.Atualizar(u); err != nil {
		return "", errors.New("erro ao salvar nova senha")
	}

	token, err := auth.GerarToken(u.ID, u.Login, u.Perfil, s.jwtSecret)
	if err != nil {
		return "", errors.New("erro interno ao gerar token")
	}

	return token, nil
}

// Criar cadastra um novo usuário com PrimeiroAcesso=true.
// O usuário precisará definir sua senha no primeiro login.
func (s *Service) Criar(nome, login, perfil string) (*Usuario, error) {
	// Verifica duplicidade de login.
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

// Atualizar altera nome e perfil de um usuário existente.
// Não é permitido alterar o superadmin por esta rota.
func (s *Service) Atualizar(id uint, nome, perfil string) (*Usuario, error) {
	u, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}

	if u.Perfil == "superadmin" {
		return nil, errors.New("não é permitido alterar o superadmin")
	}

	u.Nome = nome
	u.Perfil = perfil

	if err := s.repo.Atualizar(u); err != nil {
		return nil, errors.New("erro ao atualizar usuário")
	}

	return u, nil
}

// Desativar bloqueia o acesso de um usuário sem removê-lo do banco.
func (s *Service) Desativar(id uint) error {
	u, err := s.repo.BuscarPorID(id)
	if err != nil {
		return errors.New("usuário não encontrado")
	}

	if u.Perfil == "superadmin" {
		return errors.New("não é permitido desativar o superadmin")
	}

	u.Ativo = false
	return s.repo.Atualizar(u)
}

// Ativar reativa um usuário previamente desativado.
func (s *Service) Ativar(id uint) error {
	u, err := s.repo.BuscarPorID(id)
	if err != nil {
		return errors.New("usuário não encontrado")
	}

	if u.Perfil == "superadmin" {
		return errors.New("não é permitido alterar o superadmin")
	}

	u.Ativo = true
	return s.repo.Atualizar(u)
}

// ResetarSenha limpa a senha e força o usuário a definir uma nova no próximo acesso.
func (s *Service) ResetarSenha(id uint) error {
	u, err := s.repo.BuscarPorID(id)
	if err != nil {
		return errors.New("usuário não encontrado")
	}

	if u.Perfil == "superadmin" {
		return errors.New("não é permitido resetar a senha do superadmin")
	}

	u.Senha = ""
	u.PrimeiroAcesso = true
	return s.repo.Atualizar(u)
}

// Me retorna os dados do usuário autenticado pelo ID extraído do token.
func (s *Service) Me(id uint) (*Usuario, error) {
	u, err := s.repo.BuscarPorID(id)
	if err != nil {
		return nil, errors.New("usuário não encontrado")
	}
	return u, nil
}

// Listar retorna todos os usuários do sistema.
func (s *Service) Listar() ([]Usuario, error) {
	return s.repo.Listar()
}