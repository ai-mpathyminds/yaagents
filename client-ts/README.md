# @aimpathyminds/yaagents-client

TypeScript client for the [YAAgents](https://github.com/ai-mpathyminds/yaagents) Agentic REST Profile v0.2.

- Zero runtime dependencies — uses the global `fetch` API (Node ≥ 18 / all modern browsers)
- Dual ESM + CJS bundle with shipped `.d.ts` types
- Discriminated-union result style (`result.type === "clarification_required"`) and exception-style (`client.strict()`) helpers
- Declares YAAgents Profile v0.2 support

## Install

```sh
npm install @aimpathyminds/yaagents-client@0.2.0
```

Or with pnpm:

```sh
pnpm add @aimpathyminds/yaagents-client@0.2.0
```

## Quick start

```ts
import { YaAgentsClient, PROFILE_VERSION } from "@aimpathyminds/yaagents-client";

const client = new YaAgentsClient({
  baseUrl: "https://api.example.com",
  token: process.env.YAAGENTS_TOKEN!,
  tenantId: "tenant-1",
});

// Result-style (discriminated union — never throws)
const result = await client.campaigns.byId("c1").optimizations().create({ goal: "ctr" });
if (result.type === "clarification_required") {
  console.log("Needs input:", result.requiredInputs);
}

// Exception-style (strict mode — throws typed AgenticErrorBase subclass)
const strict = client.strict();
const data = await strict.campaigns.byId("c1").optimizations().create({ goal: "ctr" });
```

## License

Apache-2.0 — see [../LICENSE](../LICENSE).
