package cloud

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// KernelInstance representa um kernel vivo e persistente num container da infraestrutura Lab.
type KernelInstance struct {
	Language string
	Cmd      *exec.Cmd
	Stdin    io.WriteCloser
	Stdout   io.ReadCloser
	Mutex    sync.Mutex
	LastExec time.Time
}

var (
	kernelProxies = map[string]string{
		"python": pythonProxy,
		"node":   nodeProxy,
	}

	kernelImages = map[string]string{
		"python": "python:3.12-slim",
		"node":   "node:20-slim",
		"bash":   "ubuntu:22.04",
	}

	// Active Kernels multiplexados por token 
	// Em um ambiente escalado multi-tenant, podemos suportar N kernels.
	activeKernels = make(map[string]*KernelInstance)
	kernelsMutex  sync.Mutex
)

// spawnDockerKernel constrói a chamada nativa do Docker acatando as flags pedidas de SRE OOM/CPU e GPU
func spawnDockerKernel(lang string) (*exec.Cmd, error) {
	image, ok := kernelImages[lang]
	if !ok {
		return nil, fmt.Errorf("runtime não suportado: %s", lang)
	}

	proxyScript := kernelProxies[lang]

	// Flags brutas de Docker SRE
	// --memory 4g: Hard Limit
	// --cpus 2: Thread/Process limit
	// --gpus all: Aloca todo o barramento das placas nativas conectadas
	args := []string{
		"run", "--rm", "-i",
		"--memory", "4g",
		"--cpus", "2",
		/* "--gpus", "all", ---> Descomente depois pra evitar crash no docker engine sem nvidia-runtime em dev_linux_base */
	}

	// Adota o proxy
	args = append(args, image)
	switch lang {
	case "python":
		args = append(args, "python3", "-u", "-c", proxyScript)
	case "node":
		args = append(args, "node", "-e", proxyScript)
	case "bash":
		args = append(args, "bash", "-c", bashProxy)
	}

	return exec.Command("docker", args...), nil
}

// connectOrCreateKernel recicla instâncias ou sube um kernel Docker stateful caso inexistente ou morto.
func connectOrCreateKernel(ws *websocket.Conn, sessionID, lang string, userID string) (*KernelInstance, error) {
	kernelsMutex.Lock()
	defer kernelsMutex.Unlock()

	kernelID := sessionID + "_" + lang
	if kern, exists := activeKernels[kernelID]; exists {
		// Checar se o comando ainda tá rodando..
		if kern.Cmd != nil && kern.Cmd.ProcessState == nil {
			kern.LastExec = time.Now()
			return kern, nil
		}
		// Morreu, remover e recriar
		delete(activeKernels, kernelID)
	}

	cmd, err := spawnDockerKernel(lang)
	if err != nil {
		return nil, err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("falha ao invocar Docker kernel: %w", err)
	}

	kernel := &KernelInstance{
		Language: lang,
		Cmd:      cmd,
		Stdin:    stdin,
		Stdout:   stdout,
		LastExec: time.Now(),
	}

	activeKernels[kernelID] = kernel

	// Rotina assíncrona para ler Stdout continuamente (streaming do kernel stateful) e rotear WS
	go readKernelStdout(kernel, ws, kernelID)

	return kernel, nil
}

func readKernelStdout(kernel *KernelInstance, ws *websocket.Conn, kernelID string) {
	scanner := bufio.NewScanner(kernel.Stdout)
	for scanner.Scan() {
		line := scanner.Text()
		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(line), &msg); err == nil {
			if stream, ok := msg["stream"]; ok {
				safeWsSend(ws, map[string]interface{}{
					"type":    stream,
					"data":    msg["data"],
					"cell_id": msg["cell_id"],
				})
			} else if status, ok := msg["status"]; ok {
				if status == "end" {
					safeWsSend(ws, map[string]interface{}{
						"type":      "exit",
						"exit_code": 0,
						"cell_id":   msg["cell_id"],
					})
				} else if status == "start" {
					safeWsSend(ws, map[string]interface{}{
						"type":    "status",
						"data":    "[Native SRE] Docker Engine executando celular...\n",
						"cell_id": msg["cell_id"],
					})
				}
			}
		} else {
			// Plaintext falback print
			safeWsSend(ws, map[string]interface{}{
				"type": "stdout",
				"data": line + "\n",
			})
		}
	}

	log.Printf("Kernel Docker %s finalizado.", kernelID)
	kernelsMutex.Lock()
	delete(activeKernels, kernelID)
	kernelsMutex.Unlock()
}

func safeWsSend(ws *websocket.Conn, payload map[string]interface{}) {
	if ws != nil {
		websocket.JSON.Send(ws, payload)
	}
}

// Watchdog: Timeout idle cleaning
func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			kernelsMutex.Lock()
			now := time.Now()
			for id, kern := range activeKernels {
				if now.Sub(kern.LastExec) > 30*time.Minute {
					log.Printf("[Kernel Watchdog] Matando container idle: %s", id)
					if kern.Cmd != nil && kern.Cmd.Process != nil {
						kern.Cmd.Process.Kill()
					}
					delete(activeKernels, id)
				}
			}
			kernelsMutex.Unlock()
		}
	}()
}

// PROXIES GLOBAIS NATIVOS (.PY e .JS)

const pythonProxy = `
import sys, json, traceback

env = {'__name__': '__main__'}

class StreamProxy:
    def __init__(self, name):
        self.name = name
        self.real = sys.__stdout__
    def write(self, data):
        if data:
            self.real.write(json.dumps({"stream": self.name, "data": data, "cell_id": env.get('__CELL_ID__', '')}) + "\n")
            self.real.flush()
    def flush(self): pass

sys.stdout = StreamProxy("stdout")
sys.stderr = StreamProxy("stderr")

for line in sys.__stdin__:
    try:
        req = json.loads(line)
        cell_id = req.get("cell_id", "")
        env['__CELL_ID__'] = cell_id
        
        sys.__stdout__.write(json.dumps({"status": "start", "cell_id": cell_id}) + "\n")
        sys.__stdout__.flush()
        try:
            exec(req.get("code", ""), env)
        except Exception:
            traceback.print_exc(file=sys.stderr)
        finally:
            sys.__stdout__.write(json.dumps({"status": "end", "cell_id": cell_id}) + "\n")
            sys.__stdout__.flush()
    except Exception:
        pass
`

const nodeProxy = `
const vm = require('vm');
const context = vm.createContext({
	console: Object.assign({}, console),
	require: require,
	setTimeout: setTimeout,
	clearTimeout: clearTimeout,
	setInterval: setInterval,
	clearInterval: clearInterval,
	process: process,
	Buffer: Buffer
});

let currentCellId = '';

['stdout', 'stderr'].forEach(stream => {
	const orig = process[stream].write;
	process[stream].write = function(chunk, encoding, callback) {
		const outBytes = Buffer.isBuffer(chunk) ? chunk.toString('utf8') : chunk;
		process._rawDebug(JSON.stringify({stream: stream, data: outBytes, cell_id: currentCellId}));
		return true; // prevent default write
	};
});

context.console.log = (...args) => process.stdout.write(args.join(' ') + '\n');
context.console.error = (...args) => process.stderr.write(args.join(' ') + '\n');

const readline = require('readline');
const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

rl.on('line', (line) => {
	try {
		const req = JSON.parse(line);
		currentCellId = req.cell_id || '';
		process._rawDebug(JSON.stringify({status: 'start', cell_id: currentCellId}));
		
		try {
			vm.runInContext(req.code || '', context);
		} catch (err) {
			process.stderr.write(err.stack || String(err));
		} finally {
			process._rawDebug(JSON.stringify({status: 'end', cell_id: currentCellId}));
		}
	} catch (e) {
	}
});
`

// bashProxy usa leitura linha-a-linha de JSON via sed/awk (sem dependência de python3 ou jq na imagem ubuntu)
var bashProxy = "while IFS= read -r line; do\n" +
	"  cell_id=$(echo \"$line\" | sed -n 's/.*\"cell_id\":\"\\([^\"]*\\)\".*/\\1/p')\n" +
	"  code=$(echo \"$line\" | sed -n 's/.*\"code\":\"\\(.*\\)\"/\\1/p' | sed 's/\\\\n/\\n/g')\n" +
	"  echo \"{\\\"status\\\":\\\"start\\\",\\\"cell_id\\\":\\\"$cell_id\\\"}\"\n" +
	"  output=$(eval \"$code\" 2>&1) || true\n" +
	"  if [ -n \"$output\" ]; then\n" +
	"    echo \"$output\" | while IFS= read -r oline; do\n" +
	"      printf '{\\\"stream\\\":\\\"stdout\\\",\\\"data\\\":\\\"%s\\\\n\\\",\\\"cell_id\\\":\\\"%s\\\"}\\n' \"$oline\" \"$cell_id\"\n" +
	"    done\n" +
	"  fi\n" +
	"  echo \"{\\\"status\\\":\\\"end\\\",\\\"cell_id\\\":\\\"$cell_id\\\"}\"\n" +
	"done\n"

// RestartKernel mata o kernel ativo de uma sessão e força recriação no próximo request.
func RestartKernel(sessionID, lang string) {
	kernelsMutex.Lock()
	defer kernelsMutex.Unlock()
	kernelID := sessionID + "_" + lang
	if kern, exists := activeKernels[kernelID]; exists {
		log.Printf("[Kernel] Restart solicitado: %s", kernelID)
		if kern.Stdin != nil {
			kern.Stdin.Close()
		}
		if kern.Cmd != nil && kern.Cmd.Process != nil {
			kern.Cmd.Process.Kill()
		}
		delete(activeKernels, kernelID)
	}
}
