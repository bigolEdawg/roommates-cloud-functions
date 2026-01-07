package inviteuser

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"

    "cloud.google.com/go/firestore"
    "github.com/GoogleCloudPlatform/functions-framework-go/functions"
    // "google.golang.org/api/iterator"
)

var firestoreClient *firestore.Client

// this code will send invites to people so they can join the roommate group


// users/invites/invite id
func acceptInvite(ctx context.Context, client *firestore.Client, uid string, data map[string]interface{}) error {
	// send an invite to the user

	// get the doc ref for the user
	userRef := client.Collection("users").Doc(uid).Collection("invites")

	_, err := userRef.Set(ctx, data)

	if err != nil {
		return fmt.Errorf("error setting document data")
	}

	return nil
}

func editGroup(ctx context.Context, client *firestore.Client, uid string) error{
	
	// groupRef()
	
	return nil
}


func AcceptInviteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var RequestBody struct {
		UserID		string		`json:"user-id"`
	}

	// unpack json into request body

	if err := json.NewDecoder(r.Body).Decode(&RequestBody); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
		return
	}

	// add group id
	// user id
	// timestamp
	data := map[string]interface{}{
		"group-id" 		: 		"",
		"user-id" 		: 		"",
		"created-at" 	: 		"",
	}
	
	err := sendInvite(ctx, firestoreClient, RequestBody.UserID, data)

	if err != nil {
		return
	}

	// return success message

	ms := map[string]interface{}{
		"message" : "sucessfully invited user to group",
	}


    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ms)

}

func init() {
	ctx := context.Background()
    projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	var err error
	firestoreClient, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to initialize Firestore client: %v", err)
	}	
	functions.HTTP("AcceptInviteHandler", AcceptInviteHandler)
}




