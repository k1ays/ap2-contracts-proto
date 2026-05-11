"""
generate_report.py
Generates the final PDF report for Assignment 6: Automation in SRE,
Capacity Planning and Load Testing.
"""

import os
from fpdf import FPDF

SCREENSHOTS_DIR = os.path.join(os.path.dirname(__file__), "screenshots")
OUTPUT_PDF = os.path.join(os.path.dirname(__file__), "Assignment_6_Report.pdf")


class SREReport(FPDF):
    def header(self):
        self.set_font("Helvetica", "B", 10)
        self.cell(0, 8, "Assignment 6: Automation in SRE", align="C", new_x="LMARGIN", new_y="NEXT")
        self.line(10, self.get_y(), 200, self.get_y())
        self.ln(3)

    def footer(self):
        self.set_y(-15)
        self.set_font("Helvetica", "I", 8)
        self.cell(0, 10, f"Page {self.page_no()}/{{nb}}", align="C")

    def chapter_title(self, title):
        self.set_font("Helvetica", "B", 14)
        self.set_fill_color(230, 240, 255)
        self.cell(0, 10, title, new_x="LMARGIN", new_y="NEXT", fill=True)
        self.ln(4)

    def section_title(self, title):
        self.set_font("Helvetica", "B", 11)
        self.cell(0, 7, title, new_x="LMARGIN", new_y="NEXT")
        self.ln(2)

    def body_text(self, text):
        self.set_font("Helvetica", "", 10)
        self.multi_cell(0, 5.5, text)
        self.ln(2)

    def code_block(self, code):
        self.set_font("Courier", "", 8)
        self.set_fill_color(245, 245, 245)
        lines = code.strip().split("\n")
        for line in lines:
            self.cell(0, 4.5, "  " + line, new_x="LMARGIN", new_y="NEXT", fill=True)
        self.ln(3)

    def add_image_safe(self, path, w=170):
        if os.path.exists(path):
            self.image(path, w=w)
            self.ln(5)
        else:
            self.body_text(f"[Image not found: {path}]")


def generate_report():
    pdf = SREReport()
    pdf.alias_nb_pages()
    pdf.set_auto_page_break(auto=True, margin=20)

    # ---- TITLE PAGE ----
    pdf.add_page()
    pdf.ln(40)
    pdf.set_font("Helvetica", "B", 24)
    pdf.cell(0, 15, "Assignment 6", align="C", new_x="LMARGIN", new_y="NEXT")
    pdf.set_font("Helvetica", "B", 18)
    pdf.cell(0, 12, "Automation in SRE,", align="C", new_x="LMARGIN", new_y="NEXT")
    pdf.cell(0, 12, "Capacity Planning and Load Testing", align="C", new_x="LMARGIN", new_y="NEXT")
    pdf.ln(15)
    pdf.set_font("Helvetica", "", 12)
    pdf.cell(0, 8, "Course: Introduction to SRE", align="C", new_x="LMARGIN", new_y="NEXT")
    pdf.cell(0, 8, "Form: Individual Project Report", align="C", new_x="LMARGIN", new_y="NEXT")
    pdf.ln(20)
    pdf.set_font("Helvetica", "I", 10)
    pdf.cell(0, 8, "Objective: Understand the significance of automation in SRE by predicting", align="C", new_x="LMARGIN", new_y="NEXT")
    pdf.cell(0, 8, "traffic growth, setting up autoscaling infrastructure, and evaluating", align="C", new_x="LMARGIN", new_y="NEXT")
    pdf.cell(0, 8, "system performance under simulated loads.", align="C", new_x="LMARGIN", new_y="NEXT")

    # ---- STEP 1: PREDICTIVE ANALYSIS ----
    pdf.add_page()
    pdf.chapter_title("Step 1: Predictive Analysis (Python & ELK)")

    pdf.section_title("1.1 Log Generation Script (generate_logs.py)")
    pdf.body_text(
        "A synthetic log generator was developed to create realistic server access logs "
        "for an e-commerce API service. The script generates JSON-formatted log entries "
        "spanning 12 months with configurable growth parameters:\n\n"
        "- Base daily requests: 5,000\n"
        "- Monthly growth rate: 12%\n"
        "- Daily variance: +/- 20%\n"
        "- Realistic hourly distribution (peak hours 9-17)\n"
        "- 10 API endpoints simulating e-commerce operations\n\n"
        "Total generated: 3,610,150 log entries across 12 months."
    )

    pdf.code_block(
        "# Key configuration parameters\n"
        "BASE_DAILY_REQUESTS = 5000\n"
        "MONTHLY_GROWTH_RATE = 0.12  # 12% month-over-month\n"
        "\n"
        "# Log entry structure\n"
        '{\n'
        '  "timestamp": "2025-01-15T14:32:18",\n'
        '  "method": "GET",\n'
        '  "endpoint": "/api/products",\n'
        '  "status_code": 200,\n'
        '  "response_time_ms": 145,\n'
        '  "client_ip": "192.168.42.15",\n'
        '  "user_agent": "Chrome/120"\n'
        '}'
    )

    pdf.section_title("1.2 Log Extraction & Monthly Aggregation (log_extractor.py)")
    pdf.body_text(
        "The log extractor parses JSON log entries and computes monthly aggregations "
        "including total requests, error counts, error rates, and average response times."
    )
    pdf.body_text("Monthly Traffic Report:")
    pdf.code_block(
        "Month         Requests     Errors    Error %    Avg RT (ms)\n"
        "---------------------------------------------------------------------------\n"
        "2025-01        159,517     47,821     29.98%        1006.3\n"
        "2025-02        162,781     49,191     30.22%        1003.7\n"
        "2025-03        184,175     55,303     30.03%        1004.9\n"
        "2025-04        206,229     61,740     29.94%        1006.2\n"
        "2025-05        240,966     72,096     29.92%        1005.2\n"
        "2025-06        258,765     77,748     30.05%        1006.3\n"
        "2025-07        311,291     93,714     30.10%        1006.0\n"
        "2025-08        347,100    103,916     29.94%        1005.2\n"
        "2025-09        372,781    111,720     29.97%        1004.8\n"
        "2025-10        453,193    135,911     29.99%        1004.6\n"
        "2025-11        482,426    144,886     30.03%        1004.9\n"
        "2025-12        430,926    129,465     30.04%        1004.8"
    )

    pdf.section_title("1.3 ELK Stack Integration (elk_integration.py)")
    pdf.body_text(
        "An Elasticsearch integration script demonstrates the full ELK pipeline:\n\n"
        "1. Index Mapping: Defines Elasticsearch mappings for log fields (date, keyword, "
        "integer, ip types) with ILM policy for log rotation.\n\n"
        "2. Aggregation Queries: Simulates ES date_histogram aggregation to compute "
        "monthly traffic with nested sub-aggregations for avg response time, error count, "
        "and top endpoints.\n\n"
        "3. Kibana Dashboard: Formats output as a Kibana-style dashboard showing key "
        "metrics per month.\n\n"
        "In production, Filebeat would ship logs to Logstash for parsing, then to "
        "Elasticsearch for indexing, and Kibana would provide real-time visualizations."
    )

    pdf.section_title("1.4 Traffic Forecasting (traffic_forecast.py)")
    pdf.body_text(
        "The forecasting module calculates month-over-month growth rates and uses "
        "exponential trend fitting (numpy polyfit on log-transformed data) to predict "
        "traffic for the next 6 months.\n\n"
        "Growth Analysis Results:\n"
        "- Simple Average Growth Rate: 9.81%\n"
        "- Fitted Exponential Growth Rate: 11.55%\n\n"
        "6-Month Traffic Forecast:"
    )
    pdf.code_block(
        "Month         Forecasted Requests\n"
        "-----------------------------------\n"
        "  2026-01                 569,909\n"
        "  2026-02                 635,715\n"
        "  2026-03                 709,119\n"
        "  2026-04                 790,999\n"
        "  2026-05                 882,334\n"
        "  2026-06                 984,215"
    )

    pdf.section_title("Forecast Visualization")
    forecast_img = os.path.join(SCREENSHOTS_DIR, "traffic_forecast.png")
    pdf.add_image_safe(forecast_img, w=170)

    # ---- STEP 2: KUBERNETES HPA ----
    pdf.add_page()
    pdf.chapter_title("Step 2: Infrastructure Scaling (Kubernetes HPA)")

    pdf.section_title("2.1 Application Deployment")
    pdf.body_text(
        "A Flask-based e-commerce API was containerized and deployed to Kubernetes. "
        "The application includes:\n\n"
        "- /health - Health check endpoint for readiness/liveness probes\n"
        "- /api/products - Product listing\n"
        "- /api/products/<id> - Product details\n"
        "- /api/cart - Shopping cart\n"
        "- /api/search - CPU-intensive search (triggers HPA scaling)\n"
        "- /api/orders - Order history"
    )

    pdf.section_title("2.2 Deployment Configuration (deployment.yaml)")
    pdf.code_block(
        "apiVersion: apps/v1\n"
        "kind: Deployment\n"
        "metadata:\n"
        "  name: ecommerce-api\n"
        "spec:\n"
        "  replicas: 2\n"
        "  selector:\n"
        "    matchLabels:\n"
        "      app: ecommerce-api\n"
        "  template:\n"
        "    spec:\n"
        "      containers:\n"
        "        - name: ecommerce-api\n"
        "          image: ecommerce-api:latest\n"
        "          ports:\n"
        "            - containerPort: 8080\n"
        "          resources:\n"
        "            requests:\n"
        '              cpu: "100m"\n'
        '              memory: "128Mi"\n'
        "            limits:\n"
        '              cpu: "500m"\n'
        '              memory: "256Mi"\n'
        "          readinessProbe:\n"
        "            httpGet:\n"
        "              path: /health\n"
        "              port: 8080"
    )

    pdf.section_title("2.3 Horizontal Pod Autoscaler (hpa.yaml)")
    pdf.body_text(
        "The HPA is configured with autoscaling/v2 API to scale based on "
        "CPU utilization with an 80% threshold:"
    )
    pdf.code_block(
        "apiVersion: autoscaling/v2\n"
        "kind: HorizontalPodAutoscaler\n"
        "metadata:\n"
        "  name: ecommerce-api-hpa\n"
        "spec:\n"
        "  scaleTargetRef:\n"
        "    apiVersion: apps/v1\n"
        "    kind: Deployment\n"
        "    name: ecommerce-api\n"
        "  minReplicas: 2\n"
        "  maxReplicas: 10\n"
        "  metrics:\n"
        "    - type: Resource\n"
        "      resource:\n"
        "        name: cpu\n"
        "        target:\n"
        "          type: Utilization\n"
        "          averageUtilization: 80\n"
        "  behavior:\n"
        "    scaleUp:\n"
        "      stabilizationWindowSeconds: 30\n"
        "      policies:\n"
        "        - type: Pods\n"
        "          value: 2\n"
        "          periodSeconds: 60\n"
        "    scaleDown:\n"
        "      stabilizationWindowSeconds: 120\n"
        "      policies:\n"
        "        - type: Pods\n"
        "          value: 1\n"
        "          periodSeconds: 60"
    )

    pdf.section_title("2.4 HPA Scaling Behavior")
    pdf.body_text(
        "Key HPA configuration decisions:\n\n"
        "- CPU threshold of 80%: Triggers scaling before pods become fully saturated, "
        "maintaining headroom for traffic bursts.\n\n"
        "- Min replicas = 2: Ensures high availability even during low traffic.\n\n"
        "- Max replicas = 10: Caps scaling to prevent runaway costs while providing "
        "sufficient capacity for peak loads.\n\n"
        "- Scale-up stabilization (30s): Rapid response to traffic spikes by adding "
        "up to 2 pods per 60 seconds.\n\n"
        "- Scale-down stabilization (120s): Conservative downscaling to prevent "
        "flapping during variable traffic patterns."
    )

    pdf.section_title("2.5 Deployment Commands")
    pdf.code_block(
        "# Start Minikube\n"
        "minikube start --driver=docker\n\n"
        "# Build Docker image in Minikube\n"
        "eval $(minikube docker-env)\n"
        "docker build -t ecommerce-api:latest ./kubernetes/\n\n"
        "# Deploy application and HPA\n"
        "kubectl apply -f kubernetes/deployment.yaml\n"
        "kubectl apply -f kubernetes/hpa.yaml\n\n"
        "# Verify deployment\n"
        "kubectl get deployments\n"
        "kubectl get hpa\n"
        "kubectl get pods"
    )

    pdf.section_title("2.6 Expected kubectl get hpa output")
    pdf.code_block(
        "NAME                 REFERENCE                  TARGETS         MINPODS   MAXPODS   REPLICAS   AGE\n"
        "ecommerce-api-hpa    Deployment/ecommerce-api   <cpu>/80%       2         10        2          5m"
    )

    # ---- STEP 3: LOAD TESTING ----
    pdf.add_page()
    pdf.chapter_title("Step 3: Load Testing Simulation")

    pdf.section_title("3.1 Load Testing Tool: Locust")
    pdf.body_text(
        "Locust was chosen as the load testing framework for its Python-based "
        "scripting, real-time web dashboard, and ability to simulate realistic "
        "user behavior patterns."
    )

    pdf.section_title("3.2 Test Configuration (locustfile.py)")
    pdf.body_text(
        "Two user classes simulate different traffic patterns:\n\n"
        "ECommerceUser (normal traffic):\n"
        "- Browse products (weight: 5) - most common action\n"
        "- View product details (weight: 3)\n"
        "- Search products (weight: 2) - CPU intensive\n"
        "- View cart (weight: 2)\n"
        "- View orders (weight: 1)\n"
        "- Health check (weight: 1)\n"
        "- Wait time: 0.5-2.0 seconds between requests\n\n"
        "TrafficSpikeUser (spike simulation):\n"
        "- Heavy search requests (weight: 8) - targets CPU-intensive endpoint\n"
        "- Browse products (weight: 2)\n"
        "- Wait time: 0.1-0.5 seconds (aggressive)\n"
        "- User weight: 3x normal users"
    )

    pdf.section_title("3.3 Running Load Tests")
    pdf.code_block(
        "# Install Locust\n"
        "pip install locust\n\n"
        "# Run with web UI\n"
        "locust -f loadtest/locustfile.py --host=http://localhost:8080\n\n"
        "# Run headless for automated testing\n"
        "locust -f loadtest/locustfile.py --host=http://<service-ip> \\\n"
        "    --headless -u 500 -r 50 --run-time 5m"
    )

    pdf.section_title("3.4 Expected HPA Scaling Behavior During Load Test")
    pdf.body_text(
        "When the load test ramps up to 500 concurrent users hitting the "
        "CPU-intensive /api/search endpoint:\n\n"
        "1. t=0s: 2 pods running (minReplicas)\n"
        "2. t=30-60s: CPU utilization exceeds 80% threshold\n"
        "3. t=60-90s: HPA adds 2 pods (scale-up policy: 2 pods per 60s)\n"
        "4. t=90-150s: If CPU still >80%, HPA scales to 6 pods\n"
        "5. t=150-210s: Continues scaling up to max 10 pods if needed\n"
        "6. After load stops: 120s stabilization window before scale-down\n"
        "7. Gradual scale-down at 1 pod per 60 seconds to minReplicas=2"
    )

    pdf.section_title("3.5 Monitoring Commands")
    pdf.code_block(
        "# Watch HPA status in real time\n"
        "kubectl get hpa --watch\n\n"
        "# Monitor pod scaling\n"
        "kubectl get pods --watch\n\n"
        "# View HPA events\n"
        "kubectl describe hpa ecommerce-api-hpa\n\n"
        "# Expected output during load test:\n"
        "NAME                              READY   STATUS    RESTARTS   AGE\n"
        "ecommerce-api-7d8f9b6c4d-abc12   1/1     Running   0          5m\n"
        "ecommerce-api-7d8f9b6c4d-def34   1/1     Running   0          5m\n"
        "ecommerce-api-7d8f9b6c4d-ghi56   1/1     Running   0          2m\n"
        "ecommerce-api-7d8f9b6c4d-jkl78   1/1     Running   0          2m\n"
        "ecommerce-api-7d8f9b6c4d-mno90   1/1     Running   0          1m\n"
        "ecommerce-api-7d8f9b6c4d-pqr12   1/1     Running   0          1m"
    )

    # ---- STEP 4: ROI & INCIDENT DOCUMENTATION ----
    pdf.add_page()
    pdf.chapter_title("Step 4: ROI & Incident Documentation")

    pdf.section_title("4.1 Three Real-World Incidents Preventable by Automation")

    pdf.set_font("Helvetica", "B", 10)
    pdf.cell(0, 7, "Incident 1: Amazon Prime Day 2018 - Website Outage", new_x="LMARGIN", new_y="NEXT")
    pdf.body_text(
        "What happened: Amazon's website experienced significant outages during "
        "Prime Day 2018, one of their biggest sales events. The site was unable to "
        "handle the massive traffic surge, resulting in error pages for millions of "
        "customers during the first hour.\n\n"
        "How automation would prevent it:\n"
        "- Predictive autoscaling based on historical Prime Day data could have "
        "pre-provisioned sufficient infrastructure.\n"
        "- Load testing at projected peak volumes would have identified capacity gaps.\n"
        "- Automated traffic forecasting (like our Step 1) would have predicted the "
        "expected load and triggered preemptive scaling.\n\n"
        "Estimated impact: Amazon reportedly lost $72-99 million in sales during "
        "the outage window."
    )

    pdf.set_font("Helvetica", "B", 10)
    pdf.cell(0, 7, "Incident 2: GitLab Database Deletion (2017)", new_x="LMARGIN", new_y="NEXT")
    pdf.body_text(
        "What happened: A GitLab engineer accidentally deleted the production database "
        "during a maintenance operation. The manual backup and recovery processes failed, "
        "resulting in 6 hours of data loss for GitLab.com users.\n\n"
        "How automation would prevent it:\n"
        "- Automated backup verification scripts would have detected failing backups "
        "before the incident.\n"
        "- Infrastructure-as-Code with automated deployment pipelines would have "
        "prevented manual database operations.\n"
        "- Automated runbook execution would have guided the incident response with "
        "predefined recovery procedures.\n"
        "- SRE automation for canary deployments and staged rollouts would have "
        "limited the blast radius.\n\n"
        "Estimated impact: ~300GB of production data lost, 18 hours of recovery time, "
        "significant reputation damage."
    )

    pdf.set_font("Helvetica", "B", 10)
    pdf.cell(0, 7, "Incident 3: Cloudflare Outage (July 2019)", new_x="LMARGIN", new_y="NEXT")
    pdf.body_text(
        "What happened: A misconfigured WAF rule caused a massive spike in CPU "
        "utilization across Cloudflare's global network, taking down millions of "
        "websites for approximately 30 minutes.\n\n"
        "How automation would prevent it:\n"
        "- Automated canary deployment of WAF rules to a small percentage of traffic "
        "would have caught the CPU spike early.\n"
        "- HPA-style autoscaling (like our Step 2) with CPU-based triggers could have "
        "automatically scaled resources when utilization exceeded thresholds.\n"
        "- Automated load testing of rule changes in staging environments would have "
        "identified the resource consumption issue before production deployment.\n"
        "- Automated rollback triggers based on CPU/latency SLOs would have reverted "
        "the change within seconds.\n\n"
        "Estimated impact: Cloudflare handles ~10% of internet traffic; the outage "
        "affected millions of websites and their users worldwide."
    )

    pdf.section_title("4.2 ROI Analysis of SRE Automation")

    pdf.set_font("Helvetica", "B", 10)
    pdf.cell(0, 7, "Cost-Benefit Analysis", new_x="LMARGIN", new_y="NEXT")
    pdf.body_text(
        "Initial Investment (One-time costs):\n"
        "- ELK Stack setup and configuration: ~$15,000\n"
        "- Kubernetes cluster setup (managed K8s): ~$5,000\n"
        "- Load testing infrastructure (Locust/similar): ~$2,000\n"
        "- Monitoring and alerting (Prometheus/Grafana): ~$8,000\n"
        "- CI/CD pipeline automation: ~$10,000\n"
        "- SRE team training: ~$10,000\n"
        "Total Initial Investment: ~$50,000"
    )
    pdf.body_text(
        "Annual Operational Savings:\n"
        "- Reduced manual scaling operations: ~$30,000/year\n"
        "  (Eliminates 2+ hours/week of manual capacity management)\n"
        "- Decreased incident response time: ~$45,000/year\n"
        "  (MTTR reduction from 2 hours to 15 minutes average)\n"
        "- Prevention of major outages: ~$200,000/year\n"
        "  (Based on industry average $5,600/minute downtime cost,\n"
        "   preventing ~3 30-min incidents per year)\n"
        "- Improved developer productivity: ~$25,000/year\n"
        "  (Automated deployments, self-healing infrastructure)\n"
        "Total Annual Savings: ~$300,000/year"
    )
    pdf.body_text(
        "ROI Calculation:\n"
        "- Year 1 ROI: ($300,000 - $50,000) / $50,000 = 500%\n"
        "- Break-even point: ~2 months\n"
        "- 3-Year NPV (at 10% discount): ~$696,000\n\n"
        "Qualitative Benefits:\n"
        "- Improved system reliability (99.9% -> 99.95% availability)\n"
        "- Faster time to market for new features\n"
        "- Better capacity planning through data-driven decisions\n"
        "- Reduced on-call burden and engineer burnout\n"
        "- Proactive issue detection vs reactive firefighting"
    )

    pdf.section_title("4.3 Key Takeaways")
    pdf.body_text(
        "1. Automation in SRE is not just about reducing toil - it fundamentally "
        "changes how organizations handle reliability at scale.\n\n"
        "2. The combination of predictive analytics (Step 1), autoscaling (Step 2), "
        "and load testing (Step 3) creates a comprehensive automation framework that:\n"
        "   - Anticipates capacity needs before they become incidents\n"
        "   - Responds to traffic changes faster than human operators\n"
        "   - Validates system behavior under stress conditions\n\n"
        "3. The ROI of SRE automation strongly favors investment, with typical "
        "payback periods under 6 months and cumulative savings growing each year "
        "as the automation reduces manual overhead and prevents costly outages."
    )

    # ---- CONCLUSION ----
    pdf.add_page()
    pdf.chapter_title("Conclusion")
    pdf.body_text(
        "This project demonstrated the full lifecycle of SRE automation for an "
        "e-commerce API service:\n\n"
        "Step 1 - Predictive Analysis: Built a complete log generation, extraction, "
        "and forecasting pipeline using Python. The exponential growth model projected "
        "traffic reaching ~984,000 monthly requests by June 2026, representing an "
        "11.55% compound monthly growth rate. Integration with the ELK stack was "
        "demonstrated for production-grade log analysis.\n\n"
        "Step 2 - Kubernetes HPA: Configured a Horizontal Pod Autoscaler with an 80% "
        "CPU utilization threshold, enabling automatic scaling between 2-10 replicas. "
        "The scaling behavior includes rapid scale-up (2 pods/60s) for traffic spikes "
        "and conservative scale-down (1 pod/60s with 120s stabilization) to prevent "
        "flapping.\n\n"
        "Step 3 - Load Testing: Developed Locust-based load tests simulating both "
        "normal user behavior and traffic spikes, targeting CPU-intensive endpoints "
        "to exercise the HPA scaling logic.\n\n"
        "Step 4 - ROI & Incidents: Analyzed three real-world incidents (Amazon, "
        "GitLab, Cloudflare) that could have been prevented or mitigated by SRE "
        "automation, and calculated an ROI of 500% in the first year with a 2-month "
        "payback period.\n\n"
        "The project demonstrates that investing in SRE automation provides both "
        "quantitative benefits (cost savings, reduced downtime) and qualitative "
        "improvements (team morale, system reliability, faster feature delivery)."
    )

    # ---- APPENDIX: FILE LISTING ----
    pdf.add_page()
    pdf.chapter_title("Appendix: Project File Structure")
    pdf.code_block(
        "sre-assignment/\n"
        "|-- scripts/\n"
        "|   |-- generate_logs.py      # Synthetic log generator\n"
        "|   |-- log_extractor.py      # Log parsing and monthly aggregation\n"
        "|   |-- traffic_forecast.py   # Growth analysis and 6-month forecast\n"
        "|   |-- elk_integration.py    # ELK stack integration demo\n"
        "|-- kubernetes/\n"
        "|   |-- app.py                # Flask e-commerce API\n"
        "|   |-- Dockerfile            # Container image definition\n"
        "|   |-- requirements.txt      # Python dependencies\n"
        "|   |-- deployment.yaml       # K8s Deployment + Service\n"
        "|   |-- hpa.yaml              # Horizontal Pod Autoscaler\n"
        "|-- loadtest/\n"
        "|   |-- locustfile.py         # Locust load testing script\n"
        "|-- screenshots/\n"
        "|   |-- traffic_forecast.png  # Forecast visualization\n"
        "|-- data/\n"
        "|   |-- server_access.log     # Generated log data\n"
        "|-- generate_report.py        # This report generator\n"
        "|-- Assignment_6_Report.pdf   # Final report"
    )

    # Save
    pdf.output(OUTPUT_PDF)
    print(f"Report saved to: {OUTPUT_PDF}")


if __name__ == "__main__":
    generate_report()
