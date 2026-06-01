# yaagents-fastapi

FastAPI SDK for the **YAAgents Agentic REST Profile v0.2**.

Supports YAAgents Profile v0.2 — see `spec/agentic-rest-profile.md`.

## Install

```
pip install yaagents-fastapi==0.2.0
```

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
