# Waypoints2-V2

## Overview

Waypoints2 is a containerized web application vulnerability scanner developed by BlackBird Security. It performs template-based security testing against target URLs using configurable detection patterns and validation engines.

## Deployment

### RabbitMQ

#### Requirement

Install Docker Desktop and run it.

#### Docker Build

```bash
cd ./docker/rabbitmq
docker compose up -d --build
```

This backend communicate with the microservice using RabbitMQ.

This is the RabbitMQ admin webpage [http://localhost:15672](http://localhost:15672).
