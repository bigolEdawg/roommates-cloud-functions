package group

import (
	"fmt"
	"log"
	// "os"
	"net/http"
	"context"
	"encoding/json"

	"cloud.google.com/go/firestore"
	// "github.com/GoogleCloudPlatform/functions-framework-go/functions"
)
// var firestoreClient *firestore.Client


/*
	- needs to take user id some how so that a user can create a group

	- create a members collection


	*/

func createGroup(ctx context.Context, docID string, userID string) error{
	groupRef := firestoreClient.Collection("groups").Doc(docID)

	groupEntry := map[string]interface{} {
		"created_by" : userID,
	}

	_, err := groupRef.Set(ctx, groupEntry, firestore.MergeAll)

	if err != nil {
		log.Printf("failed to save group data to group %s", docID)
		return fmt.Errorf("failed to create group for user_id %s", userID)
	}
	
	log.Printf("Successfully set group %s profile", docID)
	return nil
}

// save the group to firstore with inital user id 
func saveGroupToFirestore(ctx context.Context, client *firestore.Client, groupInfo map[string]interface{}) (string, error) {
	trackRef := client.Collection("groups")
	doc, _, err := trackRef.Add(ctx, groupInfo)
	if err != nil {
		return "", fmt.Errorf("failed to save group: %v", err)
	}
	return doc.ID, nil
}



// Handler to create a group
// required fields: group_id
func createGroupHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed; only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	var RequestBody struct {
		UserID				string  `json:"user_id"`		// required	
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
		// "user_id" : RequestBody.UserID, // placeholder until auth is added
		"created_at" : firestore.ServerTimestamp,
	}


	docID,err := saveGroupToFirestore(ctx, firestoreClient, groupInfo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating group id: %v", err), http.StatusInternalServerError)
		return
	}
	// later i need to add a validation to check if the user id is a string or not
	err = createGroup(ctx, docID, RequestBody.UserID)

	if err != nil {
		http.Error(w, fmt.Sprintf("Error saving user id %s to group: %v", RequestBody.UserID, err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Group %s created successfully"}`, docID)
	
}

// func init() {
// 	ctx := context.Background()
//     projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
// 	var err error
// 	firestoreClient, err = firestore.NewClient(ctx, projectID)
// 	if err != nil {
// 		log.Fatalf("Failed to initialize Firestore client: %v", err)
// 	}	
// 	functions.HTTP("createGroupHandler", createGroupHandler)
// }




