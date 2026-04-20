package security_test

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"

	"github.com/crolab/core/internal/node"
)

func TestZipSlipPrevention(t *testing.T) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Cria um arquivo malicioso tentando escapar do diretório
	w, err := zipWriter.Create("../../../etc/passwd")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("maliciou payload"))

	// Cria um arquivo normal
	w2, _ := zipWriter.Create("main.py")
	w2.Write([]byte("print('ok')"))

	zipWriter.Close()

	payload := buf.Bytes()
	destDir := t.TempDir()

	err = node.UnzipDir(payload, destDir)
	if err == nil {
		t.Fatalf("Vulnerabilidade Crítica! UnzipDir aceitou um path absoluto (ZipSlip) e não retornou erro.")
	}

	if !strings.Contains(err.Error(), "zipslip") && !strings.Contains(err.Error(), "caminho suspeito") {
		t.Errorf("Erro retornado não identificou ZipSlip. Erro real: %v", err)
	}
}
