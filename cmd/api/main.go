package main

import (
	"super-br/config"
	"super-br/db"
	"super-br/internal/domain/usuario"
	"super-br/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	database := db.Connect(cfg)
	db.Seed(database)

	usuarioRepo := usuario.NewRepository(database)
	usuarioService := usuario.NewService(usuarioRepo)
	usuarioHandler := usuario.NewHandler(usuarioService)

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
	}

	r.Run(":" + cfg.ServerPort)
}