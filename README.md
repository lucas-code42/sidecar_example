# Como rodar um Sidecar no Kubernetes: Um guia prático 🚀

Se você está começando com Kubernetes e já ouviu falar em **Sidecar Containers**, mas ainda não entendeu direito como funciona na prática, esse artigo é para você! Vamos criar um exemplo simples e direto ao ponto para ilustrar o conceito. 💡

---

## 🔹 O que é Kubernetes?

Kubernetes (ou **k8s**) é um sistema que ajuda a gerenciar aplicações em contêineres de forma automatizada. Ele lida com a **implantação**, **escalabilidade** e **execução** dos contêineres, sem que você precise se preocupar com servidores manualmente.

Se você já usou Docker, sabe que ele roda um contêiner isolado. Mas e quando precisamos gerenciar vários contêineres que precisam trabalhar juntos? Kubernetes entra exatamente aqui!

---

## 🔹 O que é um Pod no Kubernetes?

No Kubernetes, um **Pod** é a menor unidade que pode ser implantada. Ele pode conter um ou mais contêineres que:

- **Compartilham a mesma rede** (se comunicam via `localhost`);
- **Podem compartilhar volumes** para trocar arquivos;
- **Têm o mesmo ciclo de vida** (são iniciados e finalizados juntos).

Pensa no Pod como um **mini-servidor** que agrupa processos que precisam rodar juntos.

---

## 🔹 O que é um Sidecar Container?

Um **Sidecar** é um contêiner auxiliar que roda dentro do mesmo Pod que a aplicação principal, ajudando com alguma funcionalidade extra. Em nosso caso, ele fará a conversão de strings para **Base64**.

📌 **Nosso exemplo:**
1. Criamos um Pod com **dois contêineres**:
   - Um **API (base64-http)** que recebe strings e precisa convertê-las para Base64.
   - Um **Sidecar** que faz essa conversão.
2. O Sidecar compila seu próprio binário e o disponibiliza para a API dentro de um **volume compartilhado**.
3. A API principal chama o Sidecar como um executável local, sem precisar conhecê-lo diretamente.

---

## 🔹 Outros exemplos de Sidecar

O nosso exemplo usa um Sidecar como um binário executável, mas esse não é o único jeito de usá-los! Aqui estão outras aplicações comuns:

- **Proxy reverso** → Gerencia tráfego entre microserviços.
- **Coleta de logs** → Envia logs da aplicação para um sistema central.
- **Armazenamento de secrets** → Gerencia credenciais sem expô-las diretamente à aplicação.

Cada caso de uso pode exigir uma abordagem diferente!

---

## 🔹 Como funciona o `pod.yaml`?

O arquivo `pod.yaml` define como o Kubernetes deve criar e gerenciar nosso Pod.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: base64-pod
spec:
  volumes:
    - name: shared-bin
      emptyDir: {}

  containers:
    - name: base64-http
      image: base64-http:latest
      imagePullPolicy: Never
      ports:
        - containerPort: 8080
      volumeMounts:
        - mountPath: /shared-bin
          name: shared-bin

    - name: sidecar
      image: sidecar:latest
      imagePullPolicy: Never
      volumeMounts:
        - mountPath: /shared-bin
          name: shared-bin
      command: ["/bin/sh", "-c", "cp /sidecar /shared-bin/sidecar && chmod +x /shared-bin/sidecar && tail -f /dev/null"]
```

### 🔍 Explicação:
- **Volumes e VolumeMounts:**
  - Criamos um volume chamado `shared-bin` com `emptyDir: {}`. Isso significa que esse volume será um diretório compartilhado entre os contêineres do Pod e existirá **somente enquanto o Pod estiver rodando**.
  - O volume é montado em **ambos os contêineres** (`base64-http` e `sidecar`) no caminho `/shared-bin`, permitindo que o binário gerado pelo Sidecar fique acessível para a aplicação principal.

- **`imagePullPolicy: Never`** → Esse campo instrui o Kubernetes a **não tentar puxar a imagem do registro (como Docker Hub)**, pois estamos usando imagens locais no Minikube. Se esse campo não estivesse definido, o Kubernetes tentaria baixar a imagem, o que poderia causar falhas se a imagem não estivesse publicada em um repositório.

- **O `sidecar` copia seu binário para o volume compartilhado** → O comando:
  ```sh
  command: ["/bin/sh", "-c", "cp /sidecar /shared-bin/sidecar && chmod +x /shared-bin/sidecar && tail -f /dev/null"]
  ```
  faz três coisas:
  1. Copia o binário `/sidecar` para o volume compartilhado (`/shared-bin/sidecar`).
  2. Dá permissão de execução (`chmod +x`) para que a aplicação principal possa rodá-lo.
  3. Mantém o contêiner vivo com `tail -f /dev/null`, evitando que ele seja finalizado e entre em estado de CrashLoopBackOff.

---

## 🔹 Código da aplicação principal (`base64-http`)

Essa API recebe um JSON com um texto e chama o Sidecar via `exec.Command`.

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type Request struct {
	Data string `json:"data"`
}

type Response struct {
	Encoded string `json:"encoded"`
}

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	var req Request

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cmd := exec.Command("/shared-bin/sidecar", req.Data)
	output, err := cmd.Output()
	if err != nil {
		http.Error(w, "Error executing sidecar", http.StatusInternalServerError)
		return
	}

	res := Response{Encoded: strings.TrimSpace(string(output))}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func main() {
	http.HandleFunc("/encode", encodeHandler)

	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

```

---

## 🔹 Código do Sidecar

O Sidecar é um simples programa CLI que recebe uma string e retorna sua versão codificada em Base64.

```go
package main

import (
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./sidecar <string>")
		os.Exit(1)
	}

	input := os.Args[1]
	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	fmt.Println(encoded)
}
```

---

## 🔹 Dockerfiles

### **Aplicação Principal (`base64-http`)**
```dockerfile
FROM golang:1.23 AS builder
WORKDIR /app
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o app main.go

FROM alpine
WORKDIR /
COPY --from=builder /app/app /app
RUN chmod +x /app
CMD ["/app"]
```

### **Sidecar**
```dockerfile
FROM golang:1.23 AS builder
WORKDIR /app
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o sidecar main.go

FROM alpine
WORKDIR /
COPY --from=builder /app/sidecar /sidecar
RUN chmod +x /sidecar
CMD ["/sidecar"]
```

---

## 🔹 O que é o Minikube e por que usamos?

O **Minikube** é uma ferramenta que permite rodar um cluster Kubernetes localmente. Ele simula um ambiente real, perfeito para testes antes de enviar para produção.

### **Passos para rodar no Minikube**
```sh
make up            # Inicia o Minikube, carrega as imagens docker no minikube e aplica o Pod
make build         # Constrói as imagens docker
make restart       # Reinicia o Pod
make port-forward  # Conecta uma porta do pod com uma porta da máquina local
```

Agora podemos testar com:
```sh
curl -X POST "http://localhost:8080/encode" \
-H "Content-Type: application/json" \
-d '{"data": "hello-world"}'
```

Saída esperada:
```json
{"encoded":"aGVsbG8td29ybGQ="}
```

---

## 🔹 Conclusão

Neste artigo, vimos como criar e rodar um **Sidecar Container no Kubernetes** usando o Minikube. Esse padrão é muito útil para modularizar aplicações e torná-las mais flexíveis! Agora que você entendeu o conceito, pode experimentar outras abordagens, como transformar o Sidecar em um microserviço HTTP. 🚀

