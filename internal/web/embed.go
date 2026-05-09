package web

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed all:dist
var distFiles embed.FS

// FrontendHandler retorna o handler HTTP responsável por servir os arquivos estáticos do React embutidos.
// Este handler abstrai a separação de deploy: rodar o Go executa a interface sem dependência de NodeJS.
func FrontendHandler() http.Handler {
	// Obtem a subpasta dist que o Vite cria após o npm run build
	dist, err := fs.Sub(distFiles, "dist")
	if err != nil {
		log.Fatal("Falha ao embutir os arquivos do Frontend. Rode 'npm run build' primeiro.", err)
	}

	return http.FileServer(http.FS(dist))
}
