package main

import (
	"super-br/config"
	"super-br/db"
	"super-br/internal/domain/produto"
	"super-br/internal/domain/usuario"
	"super-br/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	database := db.Connect(cfg)
	db.Seed(database)

	// Usuario
	usuarioRepo := usuario.NewRepository(database)
	usuarioService := usuario.NewService(usuarioRepo)
	usuarioHandler := usuario.NewHandler(usuarioService)

	// Produto
	produtoRepo := produto.NewRepository(database)
	produtoService := produto.NewService(produtoRepo)
	produtoHandler := produto.NewHandler(produtoService)

	r := gin.Default()

	// Rotas públicas
	public := r.Group("/api")
	{
		public.POST("/auth/login", usuarioHandler.Login)
		public.POST("/auth/primeiro-acesso", usuarioHandler.PrimeiroAcesso)
	}

	// Rotas protegidas
	protected := r.Group("/api")
	protected.Use(middleware.Autenticar())
	{
		// Auth
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
	}

	r.Run(":" + cfg.ServerPort)
}