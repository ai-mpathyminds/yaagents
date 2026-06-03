// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 AimpathyMinds

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// Product represents a store product from data/products.json.
type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
	Brand    string  `json:"brand"`
}

// Customer represents a store customer from data/customers.json.
type Customer struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Purchased []string `json:"purchased"`
}

// dataDir returns the path to the data/ directory adjacent to the source file.
func dataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "data")
}

func loadProducts() ([]Product, error) {
	data, err := os.ReadFile(filepath.Join(dataDir(), "products.json"))
	if err != nil {
		return nil, err
	}
	var products []Product
	return products, json.Unmarshal(data, &products)
}

func loadCustomers() ([]Customer, error) {
	data, err := os.ReadFile(filepath.Join(dataDir(), "customers.json"))
	if err != nil {
		return nil, err
	}
	var customers []Customer
	return customers, json.Unmarshal(data, &customers)
}

func findProduct(products []Product, id string) *Product {
	for i := range products {
		if products[i].ID == id {
			return &products[i]
		}
	}
	return nil
}

func findCustomer(customers []Customer, id string) *Customer {
	for i := range customers {
		if customers[i].ID == id {
			return &customers[i]
		}
	}
	return nil
}
