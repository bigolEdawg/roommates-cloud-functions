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

/*

	Goal:
	- create a chore within the group collection 
	
	Requirements:
	- user id (this will evetually need to be varified
	 with auth instead of just passing in the user id in the body)
	
	- group id is needed to know which group is needed
	
	- Chore name
	- Chore details
	- Chore due date
	- Chore frequency (daily, weekly, monthly)
	- Chore Assigned to (user id)
	- Chore status (not started, in progress, completed, overdue)
 */

// save the group to firstore with inital user id 
func saveChoreToFirestore(ctx context.Context, client *firestore.Client, groupID string, choreInfo map[string]interface{}) (string, error) {
	choreRef := client.Collection("groups").Doc(groupID).Collection("chores")
	doc, _, err := choreRef.Add(ctx, choreInfo)
	if err != nil {
		return "", fmt.Errorf("failed to save group: %v", err)
	}
	return doc.ID, nil
}



// Handler to create a group
// required fields: group_id
func AddChoreHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed; only POST is supported", http.StatusMethodNotAllowed)
		return
	}

	// - Chore name
	// - Chore details
	// - Chore due date
	// - Chore frequency (daily, weekly, monthly)
	// - Chore Assigned to (user id)
	// - Chore status (not started, in progress, completed, overdue)
	// completedAt

	var RequestBody struct {
		UserID				string  `json:"user_id"`		// required
		GroupID				string  `json:"group_id"`		// required
		ChoreName			string	`json:"chore_name"`
		ChoreDetails		string	`json:"chore_details"`
		ChoreDueDate		string	`json:"chore_due_date"`
		ChoreFrequency		string	`json:"chore_frequency"`
		ChoreAssignee		string	`json:"chore_assignee"`
		// ChoreStatus			string	`json:"chore_status"`
	}

	// If decoding the request body fails
    if err := json.NewDecoder(r.Body).Decode(&RequestBody); err != nil {
        http.Error(w, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
        return
    }

	if RequestBody.UserID == "" || RequestBody.GroupID == "" || RequestBody.ChoreName == "" ||RequestBody.ChoreDueDate == "" || RequestBody.ChoreFrequency == "" {
		http.Error(w, "One or more required group fields are missing", http.StatusBadRequest)
		return
	}


	choreInfo := map[string]interface{}{
		"created_at" 		: firestore.ServerTimestamp,
		"group_id"			: RequestBody.GroupID,
		"chore_name"		: RequestBody.ChoreName,
		"chore_details"		: RequestBody.ChoreDetails,
		"chore_due_date"	: RequestBody.ChoreDueDate,
		"chore_frequency"	: RequestBody.ChoreFrequency,
		"chore_assignee"	: RequestBody.ChoreAssignee,
		"created_by"		: RequestBody.UserID,
		"chore_status"		: "not started",
		"completed_at"		: "NA",
	}


	docID, err := saveChoreToFirestore(ctx, firestoreClient, RequestBody.GroupID, choreInfo)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating group id: %v", err), http.StatusInternalServerError)
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
	functions.HTTP("AddChoreHandler", AddChoreHandler)
}




