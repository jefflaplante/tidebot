## Seattle Tidebot

### Passing Environment variables to docker container
`docker run --env-file=env_file_name tidebot:v0.1.0`

### Deploy to GCP CloudRun
`gcloud builds submit --tag gcr.io/PROJECT_ID/tidebot`

`gcloud beta run deploy --image gcr.io/PROJECT_ID/tidebot --platform managed`

`gcloud beta run services update tidebot --update-env-vars KEY1=VALUE1,KEY2=VALUE2`

`gcloud iam service-accounts create svc-tidebot --display-name "Service Account Tidebot"`

`gcloud beta run services add-iam-policy-binding tidebot \
    --member=serviceAccount:svc-tidebot@[PROJECT_ID].iam.gserviceaccount.com \
    --role=roles/run.invoker`
