// Package main é o ponto de entrada da API do sistema Super BR Estoque.
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
	"super-br/internal/domain/venda"
	"super-br/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	database := db.Connect(cfg)
	db.Seed(database)

	// Repositórios
	usuarioRepo := usuario.NewRepository(database)
	produtoRepo := produto.NewRepository(database)
	estoqueRepo := estoque.NewRepository(database)
	movRepo := movimentacao.NewRepository(database)
	sucataRepo := sucata.NewRepository(database)
	vendaRepo := venda.NewRepository(database)

	// Services
	usuarioService := usuario.NewService(usuarioRepo, cfg.JWTSecret)
	produtoService := produto.NewService(produtoRepo)
	estoqueService := estoque.NewService(estoqueRepo, produtoRepo, movRepo)
	sucataService := sucata.NewService(sucataRepo)
	vendaService := venda.NewService(vendaRepo, estoqueRepo, produtoRepo, movRepo)

	// Handlers
	usuarioHandler := usuario.NewHandler(usuarioService)
	produtoHandler := produto.NewHandler(produtoService)
	estoqueHandler := estoque.NewHandler(estoqueService)
	sucataHandler := sucata.NewHandler(sucataService)
	vendaHandler := venda.NewHandler(vendaService)

	r := gin.Default()

	// Rotas publicas
	public := r.Group("/api")
	{
		public.POST("/auth/login", usuarioHandler.Login)
		public.POST("/auth/primeiro-acesso", usuarioHandler.PrimeiroAcesso)
	}

	// Rotas protegidas
	protected := r.Group("/api")
	protected.Use(middleware.Autenticar(cfg.JWTSecret))
	{
		protected.GET("/auth/me", usuarioHandler.Me)

		// Usuarios
		protected.GET("/usuarios", middleware.ExigirPerfil("admin"), usuarioHandler.Listar)
		protected.POST("/usuarios", middleware.ExigirPerfil("admin"), usuarioHandler.Criar)
		protected.PUT("/usuarios/:id", middleware.ExigirPerfil("admin"), usuarioHandler.Atualizar)
		protected.PATCH("/usuarios/:id/desativar", middleware.ExigirPerfil("admin"), usuarioHandler.Desativar)
		protected.PATCH("/usuarios/:id/ativar", middleware.ExigirPerfil("admin"), usuarioHandler.Ativar)
		protected.PATCH("/usuarios/:id/resetar-senha", middleware.ExigirPerfil("admin"), usuarioHandler.ResetarSenha)

		// Produtos
		protected.GET("/produtos", produtoHandler.Listar)
		protected.GET("/produtos/:id", produtoHandler.BuscarPorID)
		protected.POST("/produtos", middleware.ExigirPerfil("admin"), produtoHandler.Criar)
		protected.PUT("/produtos/:id", middleware.ExigirPerfil("admin"), produtoHandler.Atualizar)
		protected.DELETE("/produtos/:id", middleware.ExigirPerfil("admin"), produtoHandler.Deletar)

		// Estoque - itens individuais
		protected.GET("/estoque/itens", estoqueHandler.ListarItens)
		protected.GET("/estoque/itens/:id", estoqueHandler.BuscarItemPorID)
		protected.POST("/estoque/entrada", middleware.ExigirPerfil("admin"), estoqueHandler.EntradaEstoque)
		protected.POST("/estoque/saida", middleware.ExigirPerfil("admin"), estoqueHandler.SaidaEstoque)
		protected.PATCH("/estoque/itens/:id/devolver", middleware.ExigirPerfil("admin"), estoqueHandler.DevolverItem)
		protected.PATCH("/estoque/itens/:id/emprestar", middleware.ExigirPerfil("admin"), estoqueHandler.EmprestarItem)
		protected.PATCH("/estoque/itens/:id/devolver-emprestimo", middleware.ExigirPerfil("admin"), estoqueHandler.DevolverEmprestimo)

		// Estoque - resumo por produto
		protected.GET("/estoque", estoqueHandler.ListarEstoque)
		protected.GET("/estoque/produto/:produto_id", estoqueHandler.BuscarEstoquePorProduto)

		// Sucata
		protected.GET("/sucata", sucataHandler.Listar)
		protected.GET("/sucata/:tipo", sucataHandler.BuscarPorTipo)
		protected.POST("/sucata/tipos", middleware.ExigirPerfil("admin"), sucataHandler.CadastrarTipo)
		protected.PUT("/sucata/tipos/:id", middleware.ExigirPerfil("admin"), sucataHandler.AtualizarTipo)
		protected.POST("/sucata/entrada", middleware.ExigirPerfil("admin"), sucataHandler.EntradaSucata)
		protected.POST("/sucata/saida", middleware.ExigirPerfil("admin"), sucataHandler.SaidaSucata)

		// Vendas
		// Criar: vendas e admin podem abrir uma venda
		// Confirmar/Cancelar: somente admin — acao critica que baixa estoque definitivamente
		protected.POST("/vendas", middleware.ExigirPerfil("admin", "vendas"), vendaHandler.CriarVenda)
		protected.GET("/vendas", middleware.ExigirPerfil("admin", "financeiro", "vendas"), vendaHandler.Listar)
		protected.GET("/vendas/:id", middleware.ExigirPerfil("admin", "financeiro", "vendas"), vendaHandler.BuscarPorID)
		protected.PATCH("/vendas/:id/confirmar", middleware.ExigirPerfil("admin"), vendaHandler.ConfirmarVenda)
		protected.PATCH("/vendas/:id/cancelar", middleware.ExigirPerfil("admin"), vendaHandler.CancelarVenda)

		// Dominios futuros:
		// /notificacoes  -> notificacao.Handler
		// /relatorios    -> relatorio.Handler
		// /movimentacoes -> movimentacao.Handler
	}

	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatal("Erro ao iniciar servidor: ", err)
	}
}