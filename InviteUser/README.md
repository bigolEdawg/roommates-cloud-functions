# Invite User - Roommates Group User Managemnt 

## Overview

`invite-users` is a Go package that provides a Google Cloud Function for creating and managing users within each group.  
It allows admin users to invite other users to the current group.

### Key features include:

- Invite users to group

## Prerequisites

- **Go 1.22+:** Ensure Go is installed on your system.  
- **Google Cloud SDK:** Required if deploying as a Google Cloud Function.  
- **Firestore:** The function expects an existing Firestore setup with a `groups` collection.  

## Installation

Clone the repository and install the required dependencies:

```bash
go get -u github.com/your-username/roommates-cloud-functions/invite-users
```



## Usage
### API Endpoints

This package provides an endpoint to handle chore-related operations.

### Add a Chore to a Group

- Endpoint: /add-chore

- Method: `POST`

- Request Body (JSON):

    - `user_id`(string, required): The user creating the chore.

    - `group_id` (string, required): The group ID the chore belongs to.

    - `chore_name` (string, required): The name of the chore.

    - `chore_details` (string, optional): Extra details or description.

    - `chore_due_date` (string, required): Due date for the chore.
    - `chore_frequency` (string, required): Frequency of the chore (`daily`, `weekly`, `monthly`).

    - `chore_assignee` (string, optional): The user assigned to the chore.

**Example**:
```bash
curl -X POST "https://REGION-PROJECT_ID.cloudfunctions.net/invite-user" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "group_id": "group456",
  }'
```

**Example Response**:
```bash
{
  "message": "Group 8fhd72Ks9 created successfully"
}
```

## Error Handling

The API handles errors in the following scenarios:

1. Missing Required Fields
    If one or more required fields are missing, the API returns:
    ```json
    {
        "error": "One or more required group fields are missing"
    }
    ```


2. Invalid Method
    If the endpoint is called with a non-POST request:
    ```json
    {
        "error": "Method not allowed; only POST is supported"
    }
    ```

3. Firestore Errors
    If Firestore fails to save the chore:
    ```json
    {
        "error": "Error creating group id: <specific error>"
    }
    ```

## Deployment

### Deploying as a Google Cloud Function

1. **Install Google Cloud SDK**:

    ```bash
    curl https://sdk.cloud.google.com | bash
    exec -l $SHELL
    gcloud init
    ```

2. **Authenticate with GCP**:

    ```bash
    gcloud auth login
    gcloud config set project your-gcp-project-id
    ```

3. **Deploy the Function**  
   Ensure the Go file is in the same directory you are deploying from. Use:

    ```bash
    gcloud functions deploy invite-user \
      --gen2 \
      --runtime go122 \
      --trigger-http \
      --allow-unauthenticated \
      --entry-point AddChoreHandler \
      --region="your-region"
    ```

4. **Test the Deployment**  
   Once deployed, test the function with the generated URL:

    ```bash
    curl -X POST "https://REGION-PROJECT_ID.cloudfunctions.net/invite-user" \
      -H "Content-Type: application/json" \
      -d '{
        "user_id": "user123",
        "group_id": "group456",
      }'
    ```

---

### Local Testing

To test locally:

1. **Run the server**  
   Ensure you have a `main.go` file. Run with:

    ```bash
    FUNCTION_TARGET=AddChoreHandler LOCAL_ONLY=true go run cmd/main.go
    ```

2. **Test the API locally**:

    ```bash
    curl -X POST "http://127.0.0.1:8080" \
      -H "Content-Type: application/json" \
      -d '{"user_id":"user123","group_id":"group456"}'
    ```

---

## Contribution

Contributions are welcome! Feel free to fork the repository and submit a pull request.

### Steps to Contribute

1. Fork the repository.  
2. Create a new branch for your feature or bugfix.  
3. Write your code and tests.  
4. Submit a pull request.  

---

## License

This package is open-source under the MIT License.
