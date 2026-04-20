package zip_test

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/crolab/core/internal/cli"
	"github.com/crolab/core/internal/node"
)

func TestZipDirMultipleFiles(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "a.py"), []byte("print('a')"), 0644)
	os.WriteFile(filepath.Join(tmp, "b.py"), []byte("print('b')"), 0644)
	os.MkdirAll(filepath.Join(tmp, "sub"), 0755)
	os.WriteFile(filepath.Join(tmp, "sub", "c.txt"), []byte("hello"), 0644)

	payload, err := cli.ZipDir(tmp)
	if err != nil {
		t.Fatalf("ZipDir falhou: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		t.Fatalf("zip inválido: %v", err)
	}

	names := map[string]bool{}
	for _, f := range reader.File {
		names[f.Name] = true
	}

	if !names["a.py"] || !names["b.py"] {
		t.Errorf("arquivos raiz faltando: %v", names)
	}
	if !names["sub/c.txt"] && !names[filepath.Join("sub", "c.txt")] {
		t.Errorf("subdiretório faltando: %v", names)
	}
}

func TestZipDirEmpty(t *testing.T) {
	tmp := t.TempDir()
	payload, err := cli.ZipDir(tmp)
	if err != nil {
		t.Fatalf("ZipDir em pasta vazia falhou: %v", err)
	}
	if len(payload) == 0 {
		t.Error("payload vazio para pasta vazia")
	}
}

func TestZipDirNonExistent(t *testing.T) {
	_, err := cli.ZipDir("/caminho/que/nao/existe")
	if err == nil {
		t.Error("deveria falhar com caminho inexistente")
	}
}

func TestUnzipDirValid(t *testing.T) {
	// Create a zip
	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "data.txt"), []byte("conteudo"), 0644)
	payload, _ := cli.ZipDir(src)

	// Unzip
	dest := t.TempDir()
	err := node.UnzipDir(payload, dest)
	if err != nil {
		t.Fatalf("UnzipDir falhou: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dest, "data.txt"))
	if err != nil {
		t.Fatalf("arquivo não extraído: %v", err)
	}
	if string(data) != "conteudo" {
		t.Errorf("conteúdo errado: %s", data)
	}
}

func TestUnzipDirInvalidPayload(t *testing.T) {
	dest := t.TempDir()
	err := node.UnzipDir([]byte("lixo"), dest)
	if err == nil {
		t.Error("deveria falhar com payload inválido")
	}
}

func TestZipUnzipRoundTrip(t *testing.T) {
	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "script.py"), []byte("import torch\nprint('ok')"), 0644)
	os.MkdirAll(filepath.Join(src, "models"), 0755)
	os.WriteFile(filepath.Join(src, "models", "weights.bin"), []byte{0xDE, 0xAD, 0xBE, 0xEF}, 0644)

	payload, err := cli.ZipDir(src)
	if err != nil {
		t.Fatalf("zip falhou: %v", err)
	}

	dest := t.TempDir()
	err = node.UnzipDir(payload, dest)
	if err != nil {
		t.Fatalf("unzip falhou: %v", err)
	}

	// Verify script
	data, _ := os.ReadFile(filepath.Join(dest, "script.py"))
	if string(data) != "import torch\nprint('ok')" {
		t.Errorf("script corrompido: %s", data)
	}

	// Verify binary file
	bin, _ := os.ReadFile(filepath.Join(dest, "models", "weights.bin"))
	if len(bin) != 4 || bin[0] != 0xDE {
		t.Error("arquivo binário corrompido no roundtrip")
	}
}
