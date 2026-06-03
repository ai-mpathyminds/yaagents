# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""yaagents ecommerce store example service.

Exposes POST /products/{id}/recommendations per YAAgents Profile v0.3.
Returns mock recommendations from the same product category.

To extend with a real LLM, use the prompts in examples/store/skills/.
"""

__version__ = "0.3.0"
