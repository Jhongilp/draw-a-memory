# Draw a Memory - Cloud Deployment Guide

## Prerequisites

1. **Google Cloud Account** with billing enabled
2. **gcloud CLI** installed and authenticated
3. **Clerk Account** for authentication

## Setup Steps

### 1. Create Google Cloud Project

```bash
# Set your project ID
export PROJECT_ID=draw-a-memory
export REGION=us-central1

# Create project (if new)
gcloud projects create $PROJECT_ID

# Set as default
gcloud config set project $PROJECT_ID

# Enable required APIs
gcloud services enable \
  cloudsql.googleapis.com \
  storage.googleapis.com \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  secretmanager.googleapis.com
```

### 2. Create Cloud SQL Instance

```bash
# Create PostgreSQL instance
gcloud sql instances create draw-a-memory-db \
  --database-version=POSTGRES_15 \
  --tier=db-f1-micro \
  --region=$REGION \
  --root-password=YOUR_SECURE_PASSWORD

# Create database
gcloud sql databases create draw_a_memory \
  --instance=draw-a-memory-db

# Create user
gcloud sql users create app_user \
  --instance=draw-a-memory-db \
  --password=YOUR_USER_PASSWORD
```

### 3. Create GCS Bucket

```bash
# Create bucket for photos (private by default)
export BUCKET_NAME=${PROJECT_ID}-photos

gcloud storage buckets create gs://$BUCKET_NAME \
  --location=$REGION \
  --uniform-bucket-level-access

# Set lifecycle rule to delete old soft-deleted photos (optional)
cat > lifecycle.json << EOF
{
  "lifecycle": {
    "rule": [
      {
        "action": {"type": "Delete"},
        "condition": {
          "age": 90,
          "matchesPrefix": ["deleted/"]
        }
      }
    ]
  }
}
EOF
gcloud storage buckets update gs://$BUCKET_NAME --lifecycle-file=lifecycle.json
```

### 4. Configure Secrets

```bash
# Store secrets in Secret Manager
echo -n "postgres://app_user:YOUR_USER_PASSWORD@/draw_a_memory?host=/cloudsql/${PROJECT_ID}:${REGION}:draw-a-memory-db" | \
  gcloud secrets create draw-a-memory-db-url --data-file=-

echo -n "sk_live_YOUR_CLERK_SECRET" | \
  gcloud secrets create clerk-secret-key --data-file=-

echo -n "https://your-clerk-instance.clerk.accounts.dev/.well-known/jwks.json" | \
  gcloud secrets create clerk-jwks-url --data-file=-

echo -n "YOUR_GEMINI_API_KEY" | \
  gcloud secrets create gemini-api-key --data-file=-

echo -n "https://your-frontend-domain.com" | \
  gcloud secrets create frontend-url --data-file=-
```

### 5. Grant Permissions

```bash
# Get Cloud Run service account
export SA_EMAIL="${PROJECT_ID}@appspot.gserviceaccount.com"

# Grant Secret Manager access
gcloud secrets add-iam-policy-binding draw-a-memory-db-url \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/secretmanager.secretAccessor"

# Repeat for other secrets...

# Grant Cloud SQL access
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/cloudsql.client"

# Grant GCS access
gcloud storage buckets add-iam-policy-binding gs://$BUCKET_NAME \
  --member="serviceAccount:${SA_EMAIL}" \
  --role="roles/storage.objectAdmin"
```

### 6. Deploy

```bash
# Update cloudbuild.yaml with your values
# Then submit build:
gcloud builds submit \
  --config=cloudbuild.yaml \
  --substitutions=_REGION=$REGION,_GCS_BUCKET=$BUCKET_NAME,_CLOUD_SQL_INSTANCE=${PROJECT_ID}:${REGION}:draw-a-memory-db
```

### 7. Get Service URL

```bash
gcloud run services describe draw-a-memory-api \
  --region=$REGION \
  --format='value(status.url)'
```

## Local Development with Cloud SQL Auth Proxy

```bash
# Download and run Cloud SQL Auth Proxy
./cloud-sql-proxy ${PROJECT_ID}:${REGION}:draw-a-memory-db

# In another terminal, run the server
DATABASE_URL="postgres://app_user:password@localhost:5432/draw_a_memory" go run .
```

## Costs Estimate (Monthly)

| Service | Tier | Estimated Cost |
|---------|------|----------------|
| Cloud SQL | db-f1-micro | ~$7 |
| Cloud Storage | Standard | ~$0.02/GB |
| Cloud Run | Free tier (2M requests) | $0 - $5 |
| **Total** | | **~$10-15/month** |

## Security Checklist

- [x] Private GCS bucket (no public access)
- [x] Signed URLs for photo access (15 min expiry)
- [x] Clerk JWT validation on all endpoints
- [x] Per-user photo isolation
- [x] Soft delete with retention period
- [x] HTTPS only (Cloud Run enforced)
- [x] Non-root container user
- [ ] Enable Cloud Armor for DDoS protection (optional)
- [ ] Set up Cloud Monitoring alerts
