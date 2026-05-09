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
		log.Println("[Aviso] Pasta 'dist' do Frontend não encontrada. O React não foi embutido. Para dev local, use npm run dev.")
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Frontend não compilado no Go. Acesse a porta do npm run dev."))
		})
	}

	return http.FileServer(http.FS(dist))
}
