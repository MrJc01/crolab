package security_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/crolab/core/internal/lab"
)

func TestLabPathTraversal(t *testing.T) {
	// Setup Lab Server space
	workDir := t.TempDir()

	// Esconde um segredo forgado fora do sandbox para tentarmos ler
	secretPath := filepath.Join(t.TempDir(), "secret.txt")
	os.WriteFile(secretPath, []byte("TOP_SECRET_HOST_KEY"), 0644)

	// Inicia o laboratório
	go func() {
		// Mock start
		lab.StartLabServer(":19999", workDir, "./web")
	}()

	time.Sleep(100 * time.Millisecond) // Boot wait

	// Rota do Lab internal (exposta via mux dele)
	reqUrl := "http://localhost:19999/api/file?path=../../../../../" + filepath.Base(filepath.Dir(secretPath)) + "/secret.txt"

	resp, err := http.Get(reqUrl)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Fatalf("CRITICAL VULNERABILITY: Path Traversal succeeded via /api/file, returning 200 OK for out-of-sandbox file.")
	}

	if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusBadRequest {
		t.Logf("Warning: Expected 403 Forbidden, got %d. But content was not served.", resp.StatusCode)
	}

	// Segundo teste: SetDir escapulindo
	// Tentar forçar o working directory para a raiz do root.
	payload := `{"path": "../../../../../../"}`
	req, _ := http.NewRequest("POST", "http://localhost:19999/api/setdir", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	
	resp2, _ := http.DefaultClient.Do(req)
	defer resp2.Body.Close()
	
	if resp2.StatusCode == http.StatusOK {
		t.Fatalf("CRITICAL VULNERABILITY: Managed to SetDir to outside boundary.")
	}
}
