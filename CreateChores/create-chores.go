package chores

import (
	"fmt"
	"log"
	"os"
	"net/http"
	"context"
	"encoding/json"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)
var firestoreClient *firestore.Client


func createChore(ctx context.Context, docID string, userID string) error{
	choreRef := firestoreClient.Collection("groups").Doc(docID).Collection("chores").Doc(docID)

	groupEntry := map[string]interface{} {
		"created_by" : userID,
	}

	_, err := choreRef.Set(ctx, groupEntry, firestore.MergeAll)

	if err != nil {
		log.Printf("failed to save group data to group %s", docID)
		return fmt.Errorf("failed to create group for user_id %s", userID)
	}
	
	log.Printf("Successfully set group %s profile", docID)
	return nil
}

// save the group to firstore with inital user id 
func saveGroupToFirestore(ctx context.Context, client *firestore.Client, groupID string, groupInfo map[string]interface{}) (string, error) {
	choreRef := client.Collection("groups").Doc(groupID).Collection("chores")
	doc, _, err := choreRef.Add(ctx, groupInfo)
	if err != nil {
		return "", fmt.Errorf("failed to save group: %v", err)
	}
	return doc.ID, nil
}



// Handler to create a group
// required fields: group_id
func CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed; only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	var RequestBody struct {
		UserID				string  `json:"user_id"`		// required
		GroupID				string  `json:"group_id"`		// required
	}

	// If decoding the request body fails
    if err := json.NewDecoder(r.Body).Decode(&RequestBody); err != nil {
        http.Error(w, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
        return
    }

	if RequestBody.UserID == "" {
		http.Error(w, "One or more required group fields are missing", http.StatusBadRequest)
		return
	}


	groupInfo := map[string]interface{}{
		"created_at" : firestore.ServerTimestamp,
	}


	docID,err := saveGroupToFirestore(ctx, firestoreClient, RequestBody.GroupID, groupInfo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating group id: %v", err), http.StatusInternalServerError)
		return
	}
	// later i need to add a validation to check if the user id is a string or not
	err = createChore(ctx, docID, RequestBody.UserID)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error saving user id %s to group: %v", RequestBody.UserID, err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Group %s created successfully"}`, docID)
}

func init() {
	ctx := context.Background()
    projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	var err error
	firestoreClient, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to initialize Firestore client: %v", err)
	}	
	functions.HTTP("createGroupHandler", CreateGroupHandler)
}




