current bug is I can send this
```{
    "user_id":  "invitteid",
    "group_id": "9Mufv6K3WbZ5oPP3aLuD",
    "accepted": true
}```

but i can use the default url which would creat a group under the invitees id when this is meant to accept

The way to get around this is to make a create-group endpoint and have the default respond with a access error


- in swift (or react??? or Go???) need to autheticate?? (i think this is the right way to say it)
- then i guess the uid from the autheticated user comes into handy when i tie together the frontend and backend because we pass the uid in the authroizoer?
    then on the backend side of things I have a function that autheticates the token and gets the uid from it which then can be used as the uid to fulfill whatever task im doing?



gcloud functions deploy create-group \
  --gen2 \
  --runtime go123 \
  --region us-central1 \
  --entry-point GroupHandler \
  --trigger-http \
  --set-env-vars GOOGLE_CLOUD_PROJECT=roommates-473217 \
  --allow-unauthenticated