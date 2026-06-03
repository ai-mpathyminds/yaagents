# store-go — YAAgents ecommerce reference example (Go)

Go mirror of [`examples/store/`](../store/) using
[sdk-go](https://pkg.go.dev/github.com/ai-mpathyminds/yaagents-sdk-go) and `net/http`.

The `store-go` service exposes one agentic endpoint:

```
POST /products/{id}/recommendations
```

---

## Quick start (Docker Compose)

> **SECURITY NOTE — demo token only.**
> `GATEWAY_JWT_SECRET=demo-secret-not-for-production` is hard-coded for local demos.

```bash
cd examples/store-go
docker compose up --build
```

### Get recommendations

```bash
curl -s -X POST http://localhost:8120/products/p-1/recommendations \
  -H "Content-Type: application/json" \
  -d '{"limit": 3, "exclude_purchased": true}' \
  | python3 -m json.tool
```

### Personalised recommendations

```bash
curl -s -X POST http://localhost:8120/products/p-1/recommendations \
  -H "Content-Type: application/json" \
  -H "X-Customer-Id: c-1" \
  -d '{"limit": 3}' \
  | python3 -m json.tool
```

---

## Make it real — AI-tool skills

Open `examples/store/skills/<your-ai-tool>.md` for starter prompts that extend
the mock recommender into a real LLM-driven agent.

---

## See also

- [`examples/store/`](../store/) — Python mirror using sdk-fastapi
- [Quick Start](https://ai-mpathyminds.github.io/yaagents/tutorials/quick-start/)
- [Profile Spec](https://ai-mpathyminds.github.io/yaagents/profile/)
