# SPDX-License-Identifier: Apache-2.0
# Copyright 2026 AimpathyMinds

"""Static data loaders for the yaagents store example.

Reads products.json and customers.json from examples/store/data/.
In a real service these would come from a database or external API.
"""

from __future__ import annotations

import json
from pathlib import Path

DATA_DIR = Path(__file__).resolve().parents[2] / "data"


def load_products() -> list[dict]:
    with open(DATA_DIR / "products.json") as f:
        return json.load(f)


def load_customers() -> list[dict]:
    with open(DATA_DIR / "customers.json") as f:
        return json.load(f)


def get_product(product_id: str) -> dict | None:
    return next((p for p in load_products() if p["id"] == product_id), None)


def get_customer(customer_id: str) -> dict | None:
    return next((c for c in load_customers() if c["id"] == customer_id), None)
