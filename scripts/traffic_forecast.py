"""
traffic_forecast.py
Calculates average monthly traffic growth from historical data and
forecasts expected traffic for the next 6 months.
Uses numpy for regression and matplotlib for visualization.
"""

import os
import sys
import numpy as np

# Ensure matplotlib uses a non-interactive backend for headless rendering
import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker

from log_extractor import extract_logs, aggregate_monthly_traffic

SCREENSHOT_DIR = os.path.join(os.path.dirname(__file__), "..", "screenshots")


def calculate_growth_rates(monthly_data):
    """Calculate month-over-month growth rates."""
    growth_rates = []
    for i in range(1, len(monthly_data)):
        prev = monthly_data[i - 1]["total_requests"]
        curr = monthly_data[i]["total_requests"]
        rate = (curr - prev) / prev
        growth_rates.append(rate)
    return growth_rates


def forecast_traffic(monthly_data, forecast_months=6):
    """Forecast traffic using exponential trend fitting."""
    requests = [m["total_requests"] for m in monthly_data]
    months_idx = np.arange(len(requests))

    # Fit exponential growth: log(y) = a + b*x
    log_requests = np.log(requests)
    coeffs = np.polyfit(months_idx, log_requests, 1)
    b, a = coeffs  # slope and intercept

    monthly_growth = np.exp(b) - 1  # convert to percentage

    # Generate forecast
    future_idx = np.arange(len(requests), len(requests) + forecast_months)
    forecast_values = np.exp(a + b * future_idx)

    # Generate month labels for forecast
    last_month = monthly_data[-1]["month"]
    year, month = int(last_month[:4]), int(last_month[5:])
    forecast_labels = []
    for _ in range(forecast_months):
        month += 1
        if month > 12:
            month = 1
            year += 1
        forecast_labels.append(f"{year}-{month:02d}")

    forecast_data = []
    for label, value in zip(forecast_labels, forecast_values):
        forecast_data.append({
            "month": label,
            "forecasted_requests": int(value),
        })

    return forecast_data, monthly_growth


def print_forecast(monthly_data, forecast_data, growth_rates, avg_growth):
    """Print forecast results."""
    print("\n" + "=" * 60)
    print("TRAFFIC GROWTH ANALYSIS")
    print("=" * 60)

    print("\nMonth-over-Month Growth Rates:")
    print("-" * 40)
    for i, rate in enumerate(growth_rates):
        m1 = monthly_data[i]["month"]
        m2 = monthly_data[i + 1]["month"]
        print(f"  {m1} -> {m2}: {rate:+.2%}")

    simple_avg = np.mean(growth_rates)
    print(f"\n  Simple Average Growth Rate: {simple_avg:.2%}")
    print(f"  Fitted Exponential Growth Rate: {avg_growth:.2%}")

    print("\n" + "=" * 60)
    print("6-MONTH TRAFFIC FORECAST")
    print("=" * 60)
    print(f"{'Month':<12} {'Forecasted Requests':>20}")
    print("-" * 35)
    for f in forecast_data:
        print(f"  {f['month']:<12} {f['forecasted_requests']:>18,}")
    print("=" * 60)


def plot_forecast(monthly_data, forecast_data):
    """Create a visualization of historical and forecasted traffic."""
    os.makedirs(SCREENSHOT_DIR, exist_ok=True)

    hist_months = [m["month"] for m in monthly_data]
    hist_requests = [m["total_requests"] for m in monthly_data]

    fc_months = [f["month"] for f in forecast_data]
    fc_requests = [f["forecasted_requests"] for f in forecast_data]

    all_months = hist_months + fc_months
    all_requests = hist_requests + fc_requests

    fig, ax = plt.subplots(figsize=(14, 7))

    ax.plot(
        hist_months, hist_requests,
        "bo-", linewidth=2, markersize=8, label="Historical Traffic",
    )
    ax.plot(
        fc_months, fc_requests,
        "r^--", linewidth=2, markersize=8, label="Forecasted Traffic",
    )

    # Shade forecast region
    ax.axvspan(
        len(hist_months) - 0.5, len(all_months) - 0.5,
        alpha=0.1, color="red", label="Forecast Window",
    )

    ax.set_xlabel("Month", fontsize=12)
    ax.set_ylabel("Total Requests", fontsize=12)
    ax.set_title(
        "E-Commerce API Traffic: Historical & 6-Month Forecast",
        fontsize=14, fontweight="bold",
    )
    ax.legend(fontsize=11)
    ax.yaxis.set_major_formatter(ticker.FuncFormatter(lambda x, _: f"{x:,.0f}"))
    plt.xticks(rotation=45, ha="right")
    plt.tight_layout()
    plt.grid(True, alpha=0.3)

    output_path = os.path.join(SCREENSHOT_DIR, "traffic_forecast.png")
    plt.savefig(output_path, dpi=150, bbox_inches="tight")
    print(f"\nForecast chart saved to: {output_path}")
    plt.close()
    return output_path


def main():
    print("Loading and analyzing server logs...")
    records = extract_logs()
    monthly_data = aggregate_monthly_traffic(records)

    growth_rates = calculate_growth_rates(monthly_data)
    forecast_data, avg_growth = forecast_traffic(monthly_data)

    print_forecast(monthly_data, forecast_data, growth_rates, avg_growth)
    plot_forecast(monthly_data, forecast_data)


if __name__ == "__main__":
    main()
