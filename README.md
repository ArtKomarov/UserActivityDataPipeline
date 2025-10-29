# Real-Time Data Pipeline: User Activity Processing

This project implements a scalable real-time data pipeline using Golang services, Kafka, MongoDB, Prometheus, and Grafana, all containerized with Docker Compose. It fulfills the requirements for real-time streaming, transformation, storage, and full observability.

## 1. Architecture Overview

The pipeline processes user activity events and is entirely self-contained for local deployment.

The pipeline processes user activity events and is entirely self-contained for local deployment.

| Component | Role | Technology/Service |
| :--- | :--- | :--- |
| **Docker Compose** | **Local Orchestration & Networking** | YAML/Docker |
| **Kafka/ZK** | Message broker for reliable event streaming. | Confluent Platform |
| **Producer** | Simulates and publishes user activity events at irregular intervals. | Golang |
| **Consumer** | Reads events, **calculates processing latency (transformation)**, and persists data. | Golang |
| **MongoDB** | Persistence layer for processed events. | MongoDB |
| **Observability** | Monitoring metrics and integrated alerting. | Prometheus & Grafana |
| **Provisioning** | **Automated setup of Data Sources, Dashboards, and Alerts.** | Grafana Provisioning |

## 2. Prerequisites

You must have the following installed to run the solution:

* **Docker** (Includes Docker Compose)

## 3. Setup and Run

1. **Create .env File:** Before running, create a file named `.env` in the root directory with the following content. These values are used to configure service credentials and are externalized from the `docker-compose.yml`.

    ```env
    MONGO_INITDB_ROOT_USERNAME=user
    MONGO_INITDB_ROOT_PASSWORD=password

    ME_CONFIG_MONGODB_AUTH_USERNAME=user
    ME_CONFIG_MONGODB_AUTH_PASSWORD=password

    ME_CONFIG_BASICAUTH_USERNAME=user
    ME_CONFIG_BASICAUTH_PASSWORD=pass

    GF_SECURITY_ADMIN_USER=admin
    GF_SECURITY_ADMIN_PASSWORD=admin
    ```

2. **Start the Stack:** Build the Go services (Producer/Consumer) and launch all containers in detached mode:

    ```bash
    docker compose up --build -d
    ```

3. **Verify Status:** Ensure all services, especially `mongodb`, are running:

    ```bash
    docker compose ps
    ```

## 4. Component Access

| Component | Purpose | URL | Credentials |
| :--- | :--- | :--- | :--- |
| **Grafana** | Monitoring Dashboards and Alert Management | `http://localhost:3000` | **User/Pass:** `admin`/`admin` |
| **Mongo-Express** | Web UI to inspect persisted data | `http://localhost:8088` | **Web UI Login:** `user`/`pass` |
| **Prometheus** | Metric collection and querying endpoint | `http://localhost:9090` | N/A |

## 5. Key Functionality & Verification

| Requirement | Implementation Detail | Verification Method |
| :--- | :--- | :--- |
| **Data Flow** | Producer $\rightarrow$ Kafka Topic (`user_events`) $\rightarrow$ Consumer $\rightarrow$ MongoDB. | Check `docker logs consumer -f`. |
| **Transformation** | Consumer calculates **End-to-End Latency** (`latency_ms`) for every event. | Inspect documents in Mongo-Express (`user_data_db.events`). |
| **Monitoring** | Grafana is provisioned with dashboards for event rates and ratio. | Access Grafana and navigate to the **Dashboards** section. |
| **Alerting** | An alert, **"Consumed to produced ratio is low"**, is provisioned to detect consumer lag/failure. | Check Grafana's **Alerting** tab for the active rule. |

## 6. Future Improvements

This implementation provides a robust baseline. For production readiness and advanced functionality, consider these improvements:

* **Security:** Change all default credentials defined in the `.env` file to unique, secure passwords.
* **Reliability (DLQ):** Implement a **Dead Letter Queue (DLQ)** in Kafka to capture and handle events that fail processing (e.g., malformed JSON).
* **Aggregation:** Introduce proper **windowed aggregation** logic in the Go Consumer (e.g., hourly count of unique users) to fulfill more complex analytical requirements.
* **Metrics Detail:** Integrate a JMX Exporter alongside Kafka to enable Prometheus to scrape deeper operational metrics.
