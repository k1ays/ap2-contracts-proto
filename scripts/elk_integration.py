"""
elk_integration.py
Demonstrates integration with Elasticsearch for log indexing and querying.
In production, logs would be shipped via Filebeat/Logstash to Elasticsearch
and visualized in Kibana.

This script simulates the ELK pipeline by:
1. Indexing log entries into an Elasticsearch-compatible structure
2. Running aggregation queries for monthly traffic analysis
3. Generating a Kibana-style dashboard summary

Note: Uses local JSON processing to demonstrate the logic.
      In production, replace with elasticsearch-py client calls.
"""

import json
import os
from collections import defaultdict
from datetime import datetime

LOG_FILE = os.path.join(os.path.dirname(__file__), "..", "data", "server_access.log")


def create_es_index_mapping():
    """Define the Elasticsearch index mapping for server logs."""
    mapping = {
        "mappings": {
            "properties": {
                "timestamp": {"type": "date"},
                "method": {"type": "keyword"},
                "endpoint": {"type": "keyword"},
                "status_code": {"type": "integer"},
                "response_time_ms": {"type": "integer"},
                "client_ip": {"type": "ip"},
                "user_agent": {"type": "keyword"},
            }
        },
        "settings": {
            "number_of_shards": 1,
            "number_of_replicas": 1,
            "index.lifecycle.name": "server-logs-policy",
        },
    }
    print("Elasticsearch Index Mapping:")
    print(json.dumps(mapping, indent=2))
    return mapping


def simulate_es_aggregation(log_path=LOG_FILE):
    """
    Simulate an Elasticsearch aggregation query:
    GET /server-logs/_search
    {
      "size": 0,
      "aggs": {
        "monthly_traffic": {
          "date_histogram": {
            "field": "timestamp",
            "calendar_interval": "month"
          },
          "aggs": {
            "avg_response_time": { "avg": { "field": "response_time_ms" } },
            "error_count": {
              "filter": { "range": { "status_code": { "gte": 400 } } }
            },
            "top_endpoints": {
              "terms": { "field": "endpoint", "size": 5 }
            }
          }
        }
      }
    }
    """
    monthly = defaultdict(lambda: {
        "doc_count": 0,
        "total_rt": 0,
        "errors": 0,
        "endpoints": defaultdict(int),
    })

    with open(log_path, "r") as f:
        for line in f:
            rec = json.loads(line.strip())
            month_key = rec["timestamp"][:7]
            monthly[month_key]["doc_count"] += 1
            monthly[month_key]["total_rt"] += rec["response_time_ms"]
            if rec["status_code"] >= 400:
                monthly[month_key]["errors"] += 1
            monthly[month_key]["endpoints"][rec["endpoint"]] += 1

    # Format as Elasticsearch aggregation response
    buckets = []
    for month in sorted(monthly.keys()):
        data = monthly[month]
        top_ep = sorted(
            data["endpoints"].items(), key=lambda x: -x[1]
        )[:5]
        buckets.append({
            "key_as_string": month,
            "doc_count": data["doc_count"],
            "avg_response_time": {
                "value": round(data["total_rt"] / data["doc_count"], 1),
            },
            "error_count": {"doc_count": data["errors"]},
            "top_endpoints": {
                "buckets": [
                    {"key": ep, "doc_count": count} for ep, count in top_ep
                ]
            },
        })

    return {"aggregations": {"monthly_traffic": {"buckets": buckets}}}


def print_kibana_dashboard(agg_result):
    """Print a Kibana-style dashboard summary."""
    buckets = agg_result["aggregations"]["monthly_traffic"]["buckets"]

    print("\n" + "=" * 70)
    print("KIBANA DASHBOARD: Monthly Traffic Overview")
    print("=" * 70)

    for b in buckets:
        print(f"\n  Month: {b['key_as_string']}")
        print(f"  Total Requests:     {b['doc_count']:>10,}")
        print(f"  Avg Response Time:  {b['avg_response_time']['value']:>10.1f} ms")
        print(f"  Error Count:        {b['error_count']['doc_count']:>10,}")
        print("  Top Endpoints:")
        for ep in b["top_endpoints"]["buckets"][:3]:
            print(f"    - {ep['key']}: {ep['doc_count']:,} hits")
        print("  " + "-" * 50)


if __name__ == "__main__":
    print("=" * 70)
    print("ELK STACK INTEGRATION DEMO")
    print("=" * 70)

    create_es_index_mapping()
    print("\nRunning Elasticsearch aggregation query...")
    result = simulate_es_aggregation()
    print_kibana_dashboard(result)
