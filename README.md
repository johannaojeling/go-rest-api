# Go REST API

This project contains a REST API built with Gin and Gorm. It is hosted on Google Cloud Run and persists data in Cloud
SQL (PostgreSQL).

## Pre-requisites

- Go version 1.18
- Docker
- Gcloud SDK

## Development

### Setup

Install dependencies

```bash
go mod download
```

### Testing

Run unit tests

```bash
go test ./...
```

### Running the application

Set environment variables

| Variable    | Description                                            |
|-------------|--------------------------------------------------------|
| DB_USER     | Database user                                          |
| DB_PASSWORD | Database password                                      |
| DB_NAME     | Database name                                          |
| DB_PORT     | Database port                                          |
| DB_DRIVER   | Database driver. If not set, will use `postgres`       |
| PORT        | Port for web server. If not set, will listen on `8080` |

Run PostgreSQL with Docker

```bash
docker run -d --name postgres \
-p ${DB_PORT}:5432 \
-e POSTGRES_USER=${DB_USER} \
-e POSTGRES_PASSWORD=${DB_PASSWORD} \
-e POSTGRES_DB=${DB_NAME} \
postgres
```

Run application

```bash
# Set database URL
export DB_URL="host=localhost port=${DB_PORT} user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} sslmode=disable"

# Run server
go run main.go
```

### Calling the API

Set URL

```bash
export URL="localhost:${PORT}"
```

Call API

```bash
# Create new user
curl -i -X POST ${URL}/users/ \
-H "Content-Type: application/json" \
-d '{"first_name":"Jane","last_name":"Doe","email":"jane.doe@mail.com"}'

# Get user with id 'abc123'
curl -i -X GET ${URL}/users/abc123

# Get all users
curl -i -X GET ${URL}/users/

# Update user with id 'abc123'
curl -i -X PUT ${URL}/users/abc123 \
-H "Content-Type: application/json" \
-d '{"first_name":"Jane","last_name":"Doe","email":"jane@mail.com"}'

# Delete user with id 'abc123'
curl -i -X DELETE ${URL}/users/abc123
```

## Deployment

### Deploying to Cloud Run

Set environment variables

| Variable           | Description                                                                                                                                             |
|--------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------|
| PROJECT            | GCP project                                                                                                                                             |
| REGION             | Region                                                                                                                                                  |
| CLOUD_SQL_INSTANCE | Cloud SQL instance name                                                                                                                                 |
| DB_ROOT_PASSWORD   | Database root password                                                                                                                                  |
| DB_USER            | Database user                                                                                                                                           |
| DB_PASSWORD        | Database password                                                                                                                                       |
| DB_NAME            | Database name                                                                                                                                           |
| DB_DRIVER          | Database driver. If not set, will use `postgres`                                                                                                        |
| DB_URL_SECRET      | Secret Manager secret that will contain the database URL                                                                                                |
| CLOUD_BUILD_BUCKET | Bucket used for Cloud Build                                                                                                                             |
| CLOUD_RUN_SA       | Email of service account used for Cloud Run. Needs the roles:<br/><ul><li>`roles/cloudsql.admin`</li><li>`roles/secretmanager.secretAccessor`</li></ul> |
| IMAGE              | Image tag                                                                                                                                               |
| SERVICE            | Service name                                                                                                                                            |

Create Cloud SQL instance

```bash
# Create Cloud SQL instance
gcloud sql instances create ${CLOUD_SQL_INSTANCE} \
--project=${PROJECT} \
--region=${REGION} \
--database-version=POSTGRES_14 \
--tier=db-f1-micro \
--root-password=${DB_ROOT_PASSWORD}

# Create database user
gcloud sql users create ${DB_USER} \
--project=${PROJECT} \
--instance=${CLOUD_SQL_INSTANCE} \
--password=${DB_PASSWORD}

# Create database
gcloud sql databases create $DB_NAME \
--project=${PROJECT} \
--instance=${CLOUD_SQL_INSTANCE}
```

Create Secret Manager secret for the database URL

```bash
# Set database URL
export DB_URL="host=${PROJECT}:${REGION}:${CLOUD_SQL_INSTANCE} port=5432 user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_NAME} sslmode=disable"

# Create secret
echo $DB_URL | gcloud secrets create ${DB_URL_SECRET} \
--data-file=- \
--replication-policy=user-managed \
--locations=${REGION}
```

Build container image with Cloud Build

```bash
gcloud builds submit . \
--project=${PROJECT} \
--config=./cloudbuild.yaml \
--gcs-source-staging-dir=gs://${CLOUD_BUILD_BUCKET}/staging \
--substitutions=_PROJECT=${PROJECT},_LOGS_BUCKET=${CLOUD_BUILD_BUCKET},_IMAGE=${IMAGE}
```

Deploy to Cloud Run

```bash
# Deploy to Cloud Run
gcloud run deploy ${SERVICE} \
--project=${PROJECT} \
--region=${REGION} \
--image=eu.gcr.io/${PROJECT}/${IMAGE} \
--service-account=${CLOUD_RUN_SA} \
--add-cloudsql-instances=${INSTANCE_CONNECTION_NAME} \
--update-env-vars=DRIVER=cloudsqlpostgres \
--update-secrets=DB_URL=${DB_URL_SECRET}:latest \
--platform=managed \
--no-allow-unauthenticated
```

### Calling the API

Set URL and token

```bash
export URL=$(gcloud run services describe ${SERVICE} \
--project=${PROJECT} \
--region=${REGION} \
--format='value(status.url)')

export TOKEN=$(gcloud auth print-identity-token)
```

Call API

```bash
# Create new user
curl -i -X POST ${URL}/users/ \
-H "Authorization: Bearer ${TOKEN}" \
-H "Content-Type: application/json" \
-d '{"first_name":"Jane","last_name":"Doe","email":"jane.doe@mail.com"}'

# Get user with id 'abc123'
curl -i -X GET ${URL}/users/abc123 \
-H "Authorization: Bearer ${TOKEN}"

# Get all users
curl -i -X GET ${URL}/users/ \
-H "Authorization: Bearer ${TOKEN}"

# Update user with id 'abc123'
curl -i -X PUT ${URL}/users/abc123 \
-H "Authorization: Bearer ${TOKEN}" \
-H "Content-Type: application/json" \
-d '{"first_name":"Jane","last_name":"Doe","email":"jane@mail.com"}'

# Delete user with id 'abc123'
curl -i -X DELETE ${URL}/users/abc123 \
-H "Authorization: Bearer ${TOKEN}"
```
