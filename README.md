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

The current implementation is a good prototype. To make the pipeline truly production-ready, the following improvements are critical:

### Application Resilience and Reliability

* **Dead Letter Queue (DLQ):** Implement a dedicated **Dead Letter Queue (DLQ)** topic in Kafka. If the consumer encounters a non-recoverable error (e.g., malformed JSON), it should send the failed message to the DLQ for later inspection and reprocessing.
* **Graceful Shutdown:** Implement signal handling in Go services (listening for `SIGTERM`/`SIGINT`) to ensure resources (Kafka writers/readers, database connections) are closed cleanly, preventing data corruption or connection leaks.
* **Backoff Policies:** Integrate a **backoff and retry mechanism** for all external connections (Kafka and MongoDB). If a connection fails, the application should wait for an increasing duration before retrying, instead of crashing or spamming the resource.

### Data Integrity (Kafka)

* **Required Acks:** Configure the Kafka Producer to use **`RequiredAcks: AllISRs`** (or its equivalent in the chosen library) to ensure high durability and guarantee that messages are replicated before being acknowledged.

### Security and Configuration

* **Secure Environment:** Currently, the Go services use hardcoded credentials alongside environment variables passed from the `.env`. This should be improved by eliminating **all hardcoded credentials** and reading them purely from environment variables to simplify rotation and secure deployment.
* **Credentials**: Change all default credentials defined in the `.env` file to unique, secure passwords.

### Observability Detail

* **Kafka JMX:** Add a **JMX Exporter** sidecar container to the Kafka service. This will allow Prometheus to scrape crucial JVM and internal operational metrics (e.g., consumer lag, partition health) that are currently unavailable.
