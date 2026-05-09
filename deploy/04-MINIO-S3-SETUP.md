# 🗄️ Persistência de Dados (MinIO / S3)

O pacote Go `internal/storage/s3.go` foi projetado para funcionar com a Amazon S3 API, mas para um projeto Open Source local ou em provedores Bare Metal mais baratos, instalar o **MinIO** é a escolha definitiva.

O MinIO é compatível 100% com as APIs do Amazon S3, mas ele roda localmente no seu próprio servidor.

## Subindo o MinIO via Docker no Servidor

Como a nuvem S3 cuida de dados permanentes e não processa "código de usuários" que possa causar invasão de sandbox, é seguro (e recomendado) rodá-lo dentro de um contêiner Docker regular ao lado do Crolab.

```bash
docker run -d \
  --name minio-storage \
  -p 9000:9000 \
  -p 9001:9001 \
  -v /mnt/data/crolab_storage:/data \
  -e "MINIO_ROOT_USER=crolab_admin" \
  -e "MINIO_ROOT_PASSWORD=sua_senha_secreta_123" \
  minio/minio server /data --console-address ":9001"
```

### Explicando as Portas:
- **Porta 9000:** Essa é a API S3 oficial. O Crolab Backend (Go) usará essa porta para sincronizar os Notebooks (`.ipynb`).
- **Porta 9001:** O Console visual do MinIO. Você pode acessar no navegador para gerenciar graficamente os Buckets e baixar os backups dos clientes caso necessário.

## Variáveis de Ambiente do Crolab

Para que o Crolab consiga utilizar essa conexão na produção, você precisará declarar as chaves antes de ligar o daemon do Go:

```bash
export CROLAB_S3_ENDPOINT="127.0.0.1:9000"
export CROLAB_S3_ACCESS_KEY="crolab_admin"
export CROLAB_S3_SECRET_KEY="sua_senha_secreta_123"
export CROLAB_S3_BUCKET="crolab-notebooks"

# Em seguida, rode o servidor:
crolab cloud-serve start
```

Graças à função `EnsureBucket` que programamos no `s3.go`, você não precisa criar o bucket "crolab-notebooks" no MinIO. O Go fará o *provisionamento dinâmico* no primeiro acesso automaticamente!

---
**Próximo Passo:** Gerenciar tudo na mão fica insano quando você tem dezenas de instâncias. Vamos usar Orquestração: [05-KUBERNETES-DEPLOY.md](05-KUBERNETES-DEPLOY.md).
