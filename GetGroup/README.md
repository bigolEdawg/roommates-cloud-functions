This function is used to get the chores of a group
some needed functionality:
    - security check, so authorize the user, check against false groups. 
gcloud functions deploy create-group \
  --gen2 \
  --runtime go123 \
  --region us-central1 \
  --entry-point GroupHandler \
  --trigger-http \
  --set-env-vars GOOGLE_CLOUD_PROJECT=roommates-473217 \
  --allow-unauthenticated