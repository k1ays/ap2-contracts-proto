"""
locustfile.py
Load testing script for the e-commerce API using Locust.
Simulates realistic user traffic patterns to trigger Kubernetes HPA scaling.

Usage:
    locust -f locustfile.py --host=http://localhost:8080
    locust -f locustfile.py --host=http://<service-ip> --headless \
        -u 500 -r 50 --run-time 5m
"""

from locust import HttpUser, task, between, events
import logging


class ECommerceUser(HttpUser):
    """Simulates a typical e-commerce API user."""

    wait_time = between(0.5, 2.0)

    @task(5)
    def browse_products(self):
        """Most common action: browsing products."""
        self.client.get("/api/products")

    @task(3)
    def view_product(self):
        """View individual product details."""
        product_id = self.environment.runner.user_count % 20 + 1
        self.client.get(f"/api/products/{product_id}")

    @task(2)
    def search_products(self):
        """Search — CPU-intensive endpoint that triggers HPA."""
        self.client.get("/api/search")

    @task(2)
    def view_cart(self):
        """Check shopping cart."""
        self.client.get("/api/cart")

    @task(1)
    def view_orders(self):
        """Check order history."""
        self.client.get("/api/orders")

    @task(1)
    def health_check(self):
        """Occasional health check."""
        self.client.get("/health")


class TrafficSpikeUser(HttpUser):
    """Simulates a traffic spike with aggressive requests to CPU-heavy endpoints."""

    wait_time = between(0.1, 0.5)
    weight = 3  # 3x as many spike users

    @task(8)
    def heavy_search(self):
        """Hit the CPU-intensive search endpoint aggressively."""
        self.client.get("/api/search")

    @task(2)
    def browse(self):
        self.client.get("/api/products")
