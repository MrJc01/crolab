package unit

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/crolab/core/internal/cli"
)

func TestZipDirCreatesValidArchive(t *testing.T) {
	// Setup: Criar pasta mock e arquivo com peso
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("Payload seguro do Crolab Engine"), 0644)

	// Act: Comprimir pasta via algoritmo publico
	payload, err := cli.ZipDir(tmpDir)
	if err != nil {
		t.Fatalf("ZipDir falhou inesperadamente: %v", err)
	}

	if len(payload) == 0 {
		t.Fatalf("Payload comprimido retornou bytes vazios (0)")
	}

	// Assert: Analisar integridade do ZIP nativamente (Prevenção O(1))
	reader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		t.Fatalf("Binário do Zip não é válido ou está corrompido: %v", err)
	}

	found := false
	for _, f := range reader.File {
		if f.Name == "test.txt" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Arquivo subjacente não foi encontrado dentro do Zip Payload")
	}
}
