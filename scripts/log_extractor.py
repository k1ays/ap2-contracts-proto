"""
log_extractor.py
Extracts and parses server access logs, aggregating traffic metrics
by month. Simulates the role of an ELK (Elasticsearch/Kibana) pipeline
for log analysis.
"""

import json
import os
from collections import defaultdict
from datetime import datetime

LOG_FILE = os.path.join(os.path.dirname(__file__), "..", "data", "server_access.log")


def extract_logs(log_path=LOG_FILE):
    """Parse JSON log entries and return a list of structured records."""
    records = []
    with open(log_path, "r") as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            record = json.loads(line)
            record["timestamp"] = datetime.fromisoformat(record["timestamp"])
            records.append(record)
    print(f"Extracted {len(records):,} log entries from {log_path}")
    return records


def aggregate_monthly_traffic(records):
    """Aggregate request counts and response metrics per month."""
    monthly = defaultdict(lambda: {
        "requests": 0,
        "errors": 0,
        "total_response_time": 0,
    })

    for rec in records:
        key = rec["timestamp"].strftime("%Y-%m")
        monthly[key]["requests"] += 1
        if rec["status_code"] >= 400:
            monthly[key]["errors"] += 1
        monthly[key]["total_response_time"] += rec["response_time_ms"]

    result = []
    for month_key in sorted(monthly.keys()):
        data = monthly[month_key]
        result.append({
            "month": month_key,
            "total_requests": data["requests"],
            "error_count": data["errors"],
            "error_rate": round(data["errors"] / data["requests"] * 100, 2),
            "avg_response_time_ms": round(
                data["total_response_time"] / data["requests"], 1
            ),
        })

    return result


def print_monthly_report(monthly_data):
    """Print a formatted monthly traffic report."""
    print("\n" + "=" * 75)
    print("MONTHLY TRAFFIC REPORT")
    print("=" * 75)
    print(
        f"{'Month':<12} {'Requests':>12} {'Errors':>10} "
        f"{'Error %':>10} {'Avg RT (ms)':>14}"
    )
    print("-" * 75)
    for m in monthly_data:
        print(
            f"{m['month']:<12} {m['total_requests']:>12,} {m['error_count']:>10,} "
            f"{m['error_rate']:>9.2f}% {m['avg_response_time_ms']:>14.1f}"
        )
    print("=" * 75)


if __name__ == "__main__":
    records = extract_logs()
    monthly = aggregate_monthly_traffic(records)
    print_monthly_report(monthly)
