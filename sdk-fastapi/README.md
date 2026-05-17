# yaagents-fastapi

FastAPI SDK for the **YAAgents Agentic REST Profile v0.1**.

Supports YAAgents Profile v0.1 — see `spec/agentic-rest-profile.md`.

## Quick start

```python
from yaagents_fastapi import AgenticResponse

@app.post("/campaigns/{id}/optimizations")
def optimize(id: str) -> Response:
    return AgenticResponse.accepted(
        operation_id="op-001",
        status_url=f"/campaigns/{id}/optimizations/op-001/status",
        correlation_id=ctx.correlation_id,
        request_id=ctx.request_id,
    )
```

Source: https://github.com/ai-mpathyminds/yaagents
