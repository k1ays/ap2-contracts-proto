"""
generate_logs.py
Generates synthetic server access logs for an e-commerce API service.
Logs simulate realistic traffic growth over 12 months.
"""

import random
import datetime
import json
import os

LOG_DIR = os.path.join(os.path.dirname(__file__), "..", "data")
LOG_FILE = os.path.join(LOG_DIR, "server_access.log")

ENDPOINTS = [
    "/api/products",
    "/api/products/{id}",
    "/api/cart",
    "/api/cart/checkout",
    "/api/users/login",
    "/api/users/register",
    "/api/orders",
    "/api/orders/{id}",
    "/api/search",
    "/api/recommendations",
]

STATUS_CODES = [200, 200, 200, 200, 200, 201, 301, 400, 404, 500]
METHODS = ["GET", "GET", "GET", "POST", "POST", "PUT", "DELETE"]

# Base daily requests and monthly growth rate
BASE_DAILY_REQUESTS = 5000
MONTHLY_GROWTH_RATE = 0.12  # 12% month-over-month


def generate_logs(months=12):
    """Generate synthetic server logs spanning the given number of months."""
    os.makedirs(LOG_DIR, exist_ok=True)
    start_date = datetime.datetime(2025, 1, 1)
    entries = []

    for month_offset in range(months):
        current_month = start_date + datetime.timedelta(days=30 * month_offset)
        growth_factor = (1 + MONTHLY_GROWTH_RATE) ** month_offset
        daily_requests = int(BASE_DAILY_REQUESTS * growth_factor)

        # Generate logs for ~30 days in each month
        for day in range(30):
            current_date = current_month + datetime.timedelta(days=day)
            # Add some daily variance (+/- 20%)
            day_requests = int(daily_requests * random.uniform(0.8, 1.2))

            for _ in range(day_requests):
                hour = random.choices(
                    range(24),
                    weights=[
                        1, 1, 1, 1, 1, 2, 4, 7, 9, 10,
                        10, 9, 8, 8, 9, 10, 10, 9, 7, 5,
                        4, 3, 2, 1,
                    ],
                )[0]
                minute = random.randint(0, 59)
                second = random.randint(0, 59)
                timestamp = current_date.replace(
                    hour=hour, minute=minute, second=second
                )

                entry = {
                    "timestamp": timestamp.isoformat(),
                    "method": random.choice(METHODS),
                    "endpoint": random.choice(ENDPOINTS),
                    "status_code": random.choice(STATUS_CODES),
                    "response_time_ms": random.randint(10, 2000),
                    "client_ip": f"192.168.{random.randint(1,254)}.{random.randint(1,254)}",
                    "user_agent": random.choice([
                        "Mozilla/5.0", "Chrome/120", "Safari/17",
                        "PostmanRuntime/7.36", "curl/8.4",
                    ]),
                }
                entries.append(entry)

    # Sort by timestamp
    entries.sort(key=lambda x: x["timestamp"])

    with open(LOG_FILE, "w") as f:
        for entry in entries:
            f.write(json.dumps(entry) + "\n")

    print(f"Generated {len(entries):,} log entries across {months} months")
    print(f"Log file saved to: {LOG_FILE}")
    return LOG_FILE


if __name__ == "__main__":
    generate_logs()
