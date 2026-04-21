package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/crolab/core/internal/cli"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Gerencia o Portal Web interface",
}

var webStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Inicia o servidor e abre o Portal Web proxy (use -d para background)",
	Run: runWeb,
}

var webStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Desliga o serviço web local",
	Run: func(cmd *cobra.Command, args []string) {
		cli.DaemonStop("web")
	},
}

func init() {
	webStartCmd.Flags().BoolP("daemon", "d", false, "Rodar em background via Daemon")
	webStartCmd.Flags().String("port", ":8855", "Porta para hospedar o cliente")
	
	webCmd.AddCommand(webStartCmd)
	webCmd.AddCommand(webStopCmd)
	rootCmd.AddCommand(webCmd)
}

func runWeb(cmd *cobra.Command, args []string) {
	if cli.Daemonize("web") {
		return
	}

	port, _ := cmd.Flags().GetString("port")
	
	// Start Local Background Reverse Proxy with dynamic target loading
	go startDynamicProxy(port)
	
	fmt.Println("🌐 Abrindo Portal Crolab Desktop App...")
	time.Sleep(1 * time.Second)
	open.Run("http://127.0.0.1" + port)
	
	select {} // block main thread forever
}

func startDynamicProxy(port string) {
	fmt.Println("🌟 Portal local offline. Iniciando Web Proxy Node silenciosamente...")

	cli.InitConfig()
	cfg, err := cli.LoadConfig()
	var hubURL string
	if err == nil {
		hubURL = cfg.DefaultServer
	}
	target, err := url.Parse(hubURL)
	if err != nil {
		target, _ = url.Parse("https://api.crom.cloud") 
	}

	mux := http.NewServeMux()
	
	// Dynamic Proxy handling logic simplified for snippet size...
	var proxyTarget = target
	
	proxy := httputil.NewSingleHostReverseProxy(proxyTarget)
	
	mux.HandleFunc("/local-api/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]string{"target": proxyTarget.String()})
			return
		}
		if r.Method == http.MethodPost {
			var b struct { Target string `json:"target"` }
			json.NewDecoder(r.Body).Decode(&b)
			nt, err := url.Parse(b.Target)
			if err == nil {
				proxyTarget = nt
				proxy.Director = func(req *http.Request) {
					req.URL.Scheme = proxyTarget.Scheme
					req.URL.Host = proxyTarget.Host
					req.Host = proxyTarget.Host
				}
				cli.AddServer("web-target", b.Target, "", "crom", 1)
				log.Printf("⚙️ Proxy Local reconfigurado para -> %s", b.Target)
			}
			w.WriteHeader(200)
		}
	})
	
	mux.Handle("/", proxy)

	log.Printf("💻 Servidor UI inicializado na porta %s (Proxying %s)\n", port, target)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Erro no proxy local: %v\n", err)
	}
}
