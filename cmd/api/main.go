// Package main é o ponto de entrada da API do sistema Super BR Estoque.
// Responsável por inicializar configurações, banco de dados, repositórios,
// services, handlers e registrar todas as rotas HTTP.
package main

import (
	"log"

	"super-br/config"
	"super-br/db"
	"super-br/internal/domain/estoque"
	"super-br/internal/domain/movimentacao"
	"super-br/internal/domain/produto"
	"super-br/internal/domain/sucata"
	"super-br/internal/domain/usuario"
	"super-br/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Carrega configurações do .env
	cfg := config.Load()

	// 2. Conecta ao banco e executa AutoMigrate
	database := db.Connect(cfg)

	// 3. Garante que o superadmin existe
	db.Seed(database)

	// ─── Repositórios ────────────────────────────────────────────────────────

	usuarioRepo := usuario.NewRepository(database)
	produtoRepo := produto.NewRepository(database)
	estoqueRepo := estoque.NewRepository(database)
	movRepo := movimentacao.NewRepository(database)
	sucataRepo := sucata.NewRepository(database)

	// ─── Services ────────────────────────────────────────────────────────────
	// O JWTSecret é injetado diretamente — nenhum service ou auth chama os.Getenv.

	usuarioService := usuario.NewService(usuarioRepo, cfg.JWTSecret)
	produtoService := produto.NewService(produtoRepo)

	// EstoqueService recebe movRepo para registrar movimentações de forma type-safe.
	estoqueService := estoque.NewService(estoqueRepo, produtoRepo, movRepo)

	sucataService := sucata.NewService(sucataRepo)

	// ─── Handlers ────────────────────────────────────────────────────────────

	usuarioHandler := usuario.NewHandler(usuarioService)
	produtoHandler := produto.NewHandler(produtoService)
	estoqueHandler := estoque.NewHandler(estoqueService)
	sucataHandler := sucata.NewHandler(sucataService)

	// ─── Roteador ────────────────────────────────────────────────────────────

	// gin.Default() inclui logger e recovery automático de panics.
	// Em produção, considere configurar middlewares explicitamente para
	// controlar o que é logado (evitar expor dados sensíveis nos logs).
	r := gin.Default()

	// ─── Rotas Públicas (sem autenticação) ───────────────────────────────────

	public := r.Group("/api")
	{
		// Autenticação
		public.POST("/auth/login", usuarioHandler.Login)
		public.POST("/auth/primeiro-acesso", usuarioHandler.PrimeiroAcesso)
	}

	// ─── Rotas Protegidas (requerem JWT válido) ───────────────────────────────
	// O secret JWT é injetado via closure — não usa os.Getenv no middleware.

	protected := r.Group("/api")
	protected.Use(middleware.Autenticar(cfg.JWTSecret))
	{
		// ── Auth ──────────────────────────────────────────────────────────────
		protected.GET("/auth/me", usuarioHandler.Me)

		// ── Usuários (somente admin e superadmin) ─────────────────────────────
		protected.GET("/usuarios", middleware.ExigirPerfil("admin"), usuarioHandler.Listar)
		protected.POST("/usuarios", middleware.ExigirPerfil("admin"), usuarioHandler.Criar)
		protected.PUT("/usuarios/:id", middleware.ExigirPerfil("admin"), usuarioHandler.Atualizar)
		protected.PATCH("/usuarios/:id/desativar", middleware.ExigirPerfil("admin"), usuarioHandler.Desativar)
		protected.PATCH("/usuarios/:id/ativar", middleware.ExigirPerfil("admin"), usuarioHandler.Ativar)
		protected.PATCH("/usuarios/:id/resetar-senha", middleware.ExigirPerfil("admin"), usuarioHandler.ResetarSenha)

		// ── Produtos ──────────────────────────────────────────────────────────
		// Leitura: todos os perfis autenticados
		// Escrita: somente admin
		protected.GET("/produtos", produtoHandler.Listar)
		protected.GET("/produtos/:id", produtoHandler.BuscarPorID)
		protected.POST("/produtos", middleware.ExigirPerfil("admin"), produtoHandler.Criar)
		protected.PUT("/produtos/:id", middleware.ExigirPerfil("admin"), produtoHandler.Atualizar)
		protected.DELETE("/produtos/:id", middleware.ExigirPerfil("admin"), produtoHandler.Deletar)

		// ── Estoque — itens individuais ───────────────────────────────────────
		// Leitura: todos os perfis
		// Movimentação: somente admin
		protected.GET("/estoque/itens", estoqueHandler.ListarItens)
		protected.GET("/estoque/itens/:id", estoqueHandler.BuscarItemPorID)
		protected.POST("/estoque/entrada", middleware.ExigirPerfil("admin"), estoqueHandler.EntradaEstoque)
		protected.POST("/estoque/saida", middleware.ExigirPerfil("admin"), estoqueHandler.SaidaEstoque)
		protected.PATCH("/estoque/itens/:id/devolver", middleware.ExigirPerfil("admin"), estoqueHandler.DevolverItem)
		protected.PATCH("/estoque/itens/:id/emprestar", middleware.ExigirPerfil("admin"), estoqueHandler.EmprestarItem)
		protected.PATCH("/estoque/itens/:id/devolver-emprestimo", middleware.ExigirPerfil("admin"), estoqueHandler.DevolverEmprestimo)

		// ── Estoque — resumo por produto ──────────────────────────────────────
		protected.GET("/estoque", estoqueHandler.ListarEstoque)
		protected.GET("/estoque/produto/:produto_id", estoqueHandler.BuscarEstoquePorProduto)

		// ── Sucata ────────────────────────────────────────────────────────────
		// Leitura: todos os perfis
		// Movimentação e cadastro de tipos: somente admin
		protected.GET("/sucata", sucataHandler.Listar)
		protected.GET("/sucata/:tipo", sucataHandler.BuscarPorTipo)
		protected.POST("/sucata/tipos", middleware.ExigirPerfil("admin"), sucataHandler.CadastrarTipo)
		protected.PUT("/sucata/tipos/:id", middleware.ExigirPerfil("admin"), sucataHandler.AtualizarTipo)
		protected.POST("/sucata/entrada", middleware.ExigirPerfil("admin"), sucataHandler.EntradaSucata)
		protected.POST("/sucata/saida", middleware.ExigirPerfil("admin"), sucataHandler.SaidaSucata)

		// ── Domínios futuros (serão adicionados conforme implementação) ────────
		// /vendas      → venda.Handler
		// /notificacoes → notificacao.Handler
		// /relatorios   → relatorio.Handler
		// /movimentacoes → movimentacao.Handler
	}

	// Inicia o servidor na porta configurada no .env
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Erro ao iniciar servidor: ", err)
	}
}