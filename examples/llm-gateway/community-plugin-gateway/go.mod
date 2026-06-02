module github.com/ai-mpathyminds/yaagents/examples/llm-gateway/community-plugin-gateway

go 1.25.0

require (
	github.com/ai-mpathyminds/yaagents/examples/llm-gateway/community-plugin v0.0.0-unpublished
	github.com/ai-mpathyminds/yaagents/gateway v0.0.0-unpublished
)

replace (
	github.com/ai-mpathyminds/yaagents/examples/llm-gateway/community-plugin => ../community-plugin
	github.com/ai-mpathyminds/yaagents/gateway => ../../../gateway
)
