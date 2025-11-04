# Desafio Open Telemetry - Sistema de Temperatura por CEP

**Objetivo:** Sistema em Go que recebe um CEP, identifica a cidade e retorna o clima atual (temperatura em graus celsius, fahrenheit e kelvin) juntamente com a cidade. O sistema implementa OTEL (Open Telemetry) e Zipkin para tracing distribuído.

## Arquitetura

O projeto é composto por dois serviços:

- **Serviço A** (porta 8080): Responsável por receber o input do CEP via POST e encaminhar para o Serviço B
- **Serviço B** (porta 8081): Responsável pela orquestração (consulta ViaCEP + WeatherAPI e retorna temperaturas)
- **Zipkin** (porta 9411): Serviço de observabilidade para visualizar traces distribuídos

## Requisitos Implementados

### ✅ Serviço A (Input)
- Recebe POST em `/submit` com JSON: `{ "cep": "29902555" }`
- Valida se o CEP contém exatamente 8 dígitos numéricos
- Retorna 422 com `{"message":"invalid zipcode"}` se inválido
- Encaminha requisição válida para o Serviço B via HTTP
- Propaga contexto de tracing (OTEL)

### ✅ Serviço B (Orquestração)
- Recebe GET em `/weather/{zipcode}`
- Valida formato do CEP (8 dígitos)
- Consulta ViaCEP para obter localização
- Consulta WeatherAPI para obter temperatura
- Converte temperaturas:
  - Fahrenheit: `F = C * 1.8 + 32`
  - Kelvin: `K = C + 273`
- Retorna respostas adequadas:
  - **200**: `{ "city": "São Paulo", "temp_C": 28.5, "temp_F": 83.3, "temp_K": 301.5 }`
  - **422**: `{"message":"invalid zipcode"}` (formato inválido)
  - **404**: `{"message":"can not find zipcode"}` (CEP não encontrado)

### ✅ OTEL + Zipkin
- Tracing distribuído entre Serviço A e Serviço B
- Spans para medir tempo de:
  - `viacep.Lookup` - Busca de CEP
  - `weather.CurrentTempC` - Busca de temperatura
- Exportação para Zipkin

## Como Rodar o Projeto

### Pré-requisitos

- Docker e Docker Compose instalados
- Chave de API do WeatherAPI (gratuita em https://www.weatherapi.com/)

### 1. Configurar Variável de Ambiente

Crie um arquivo `.env` na raiz do projeto com sua chave da WeatherAPI:

```bash
WEATHERAPI_KEY=sua_chave_aqui
```

Ou exporte diretamente no terminal:

```bash
export WEATHERAPI_KEY=sua_chave_aqui
```

### 2. Subir os Serviços com Docker Compose

```bash
docker compose up --build
```

Isso irá iniciar:
- **Zipkin** em http://localhost:9411
- **Serviço B** em http://localhost:8081
- **Serviço A** em http://localhost:8080

### 3. Testar a Aplicação

#### Testar via Serviço A (fluxo completo):

```bash
curl -X POST http://localhost:8080/submit \
  -H 'Content-Type: application/json' \
  -d '{"cep":"01310100"}'
```

**Resposta esperada** (sucesso):
```json
{
  "city": "São Paulo",
  "temp_C": 23.5,
  "temp_F": 74.3,
  "temp_K": 296.5
}
```

#### Testar CEP inválido:

```bash
curl -X POST http://localhost:8080/submit \
  -H 'Content-Type: application/json' \
  -d '{"cep":"123"}'
```

**Resposta esperada**:
```json
{"message":"invalid zipcode"}
```
**Status Code**: 422

#### Testar Serviço B diretamente:

```bash
curl http://localhost:8081/weather/01310100
```

### 4. Visualizar Traces no Zipkin

1. Acesse http://localhost:9411
2. Clique em "Run Query" para ver os traces
3. Clique em um trace para ver detalhes dos spans:
   - Span do Serviço A (`submit`)
   - Span do Serviço B (`service-b-router`)
   - Span da busca de CEP (`viacep.Lookup`)
   - Span da busca de temperatura (`weather.CurrentTempC`)

## Estrutura do Projeto

```
.
├── cmd/
│   ├── server/          # Serviço B (main.go)
│   └── service-a/       # Serviço A (main.go)
├── internal/
│   ├── http/           # Handlers HTTP
│   ├── otel/           # Configuração OpenTelemetry
│   ├── service/        # Lógica de negócio
│   ├── types/          # Tipos de dados
│   ├── viacep/         # Cliente ViaCEP (com span)
│   └── weather/        # Cliente WeatherAPI (com span)
├── docker-compose.yml  # Orquestração dos serviços
├── Dockerfile          # Build multi-stage
└── README.md
```

## Desenvolvimento Local (sem Docker)

### 1. Compilar o projeto:

```bash
go build ./...
```

### 2. Rodar Zipkin localmente:

```bash
docker run -d -p 9411:9411 openzipkin/zipkin:2.23
```

### 3. Exportar variáveis:

```bash
export WEATHERAPI_KEY=sua_chave_aqui
export ZIPKIN_URL=http://localhost:9411/api/v2/spans
```

### 4. Rodar Serviço B:

```bash
go run cmd/server/main.go
```

### 5. Em outro terminal, rodar Serviço A:

```bash
export SERVICE_B_URL=http://localhost:8081
go run cmd/service-a/main.go
```

## Variáveis de Ambiente

### Serviço A
- `PORT` - Porta do serviço (padrão: 8080)
- `SERVICE_B_URL` - URL base do Serviço B (padrão: http://service-b:8081)
- `ZIPKIN_URL` - URL do Zipkin (padrão: http://zipkin:9411/api/v2/spans)

### Serviço B
- `PORT` - Porta do serviço (padrão: 8081)
- `WEATHERAPI_KEY` - **Obrigatória** - Chave da API WeatherAPI
- `ZIPKIN_URL` - URL do Zipkin (padrão: http://zipkin:9411/api/v2/spans)

## Endpoints

### Serviço A

| Método | Endpoint  | Descrição                    |
|--------|-----------|------------------------------|
| POST   | /submit   | Recebe CEP e encaminha para B |

### Serviço B

| Método | Endpoint           | Descrição                       |
|--------|--------------------|---------------------------------|
| GET    | /weather/{zipcode} | Retorna clima para o CEP        |
| GET    | /health            | Health check                    |
| GET    | /ready             | Readiness check                 |

## Tecnologias Utilizadas

- **Go 1.24**
- **OpenTelemetry** - Instrumentação e tracing
- **Zipkin** - Backend de observabilidade
- **ViaCEP API** - Consulta de CEPs
- **WeatherAPI** - Consulta de temperatura
- **Docker & Docker Compose** - Containerização

## Observabilidade

O projeto implementa tracing distribuído completo:

1. **Serviço A** cria um span para a requisição `/submit`
2. **Propagação** do contexto de trace via headers HTTP
3. **Serviço B** extrai o contexto e continua o trace
4. **Spans filhos** são criados para:
   - Busca de CEP no ViaCEP
   - Busca de temperatura no WeatherAPI
5. Todos os spans são exportados para o **Zipkin**

## Troubleshooting

### Erro: "missing WEATHERAPI_KEY"
Configure a variável de ambiente `WEATHERAPI_KEY` antes de iniciar o Serviço B.

### Erro: "can not find zipcode"
- Verifique se o CEP existe e está no formato correto (8 dígitos)
- Teste o CEP diretamente no ViaCEP: https://viacep.com.br/ws/01310100/json/

### Traces não aparecem no Zipkin
- Verifique se o Zipkin está rodando em http://localhost:9411
- Verifique os logs dos serviços para erros de conexão
- Confirme que `ZIPKIN_URL` está configurada corretamente