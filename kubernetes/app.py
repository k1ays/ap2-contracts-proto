"""
app.py
Simple e-commerce API service for Kubernetes deployment.
Includes CPU-intensive endpoint for load testing HPA behavior.
"""

from flask import Flask, jsonify
import math
import time

app = Flask(__name__)


@app.route("/health")
def health():
    return jsonify({"status": "healthy"}), 200


@app.route("/api/products")
def get_products():
    # Simulate some processing
    products = [
        {"id": i, "name": f"Product {i}", "price": round(9.99 + i * 5.5, 2)}
        for i in range(1, 21)
    ]
    return jsonify(products), 200


@app.route("/api/products/<int:product_id>")
def get_product(product_id):
    return jsonify({
        "id": product_id,
        "name": f"Product {product_id}",
        "price": round(9.99 + product_id * 5.5, 2),
        "description": "A sample product for the e-commerce platform.",
    }), 200


@app.route("/api/cart")
def get_cart():
    return jsonify({"items": [], "total": 0.0}), 200


@app.route("/api/search")
def search():
    # CPU-intensive operation to trigger HPA scaling
    result = 0
    for i in range(1, 50000):
        result += math.sqrt(i) * math.sin(i)
    return jsonify({"results": [], "computation": result}), 200


@app.route("/api/orders")
def get_orders():
    return jsonify({"orders": []}), 200


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080)
