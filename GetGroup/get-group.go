package getgroup

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"

    "cloud.google.com/go/firestore"
    "github.com/GoogleCloudPlatform/functions-framework-go/functions"
    "google.golang.org/api/iterator"
)
var firestoreClient *firestore.Client

// Group struct
type groupData struct {
    GroupName     string `json:group_name`
    GroupID       string `json:"group_id"`
}

// Get chore data from a group
func getChoreFromGroup(ctx context.Context, client *firestore.Client, groupID string) ([]choreData, error) {
    var chores []choreData

    iter := client.Collection("groups").Doc(groupID).Collection("chores").Documents(ctx)
	fmt.Printf("In the getChoreFromGroup function ")
    defer iter.Stop()

    for {
        doc, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("error reading chores for group %s: %v", groupID, err)
        }

        c, err := getChoreDoc(doc)
        if err != nil {
            return nil, err
        }

        chores = append(chores, c)
    }

    return chores, nil
}

func getChoreDoc(doc *firestore.DocumentSnapshot) (choreData, error) {
    var c choreData

    data := doc.Data()

    if v, ok := data["user_id"].(string); ok {
        c.UserID = v
    }
    if v, ok := data["group_id"].(string); ok {
        c.GroupID = v
    }
    if v, ok := data["chore_name"].(string); ok {
        c.ChoreName = v
    }
    if v, ok := data["chore_details"].(string); ok {
        c.ChoreDetails = v
    }
    if v, ok := data["chore_due_date"].(string); ok {
        c.ChoreDueDate = v
    }
    if v, ok := data["chore_frequency"].(string); ok {
        c.ChoreFreq = v
    }
    if v, ok := data["chore_assignee"].(string); ok {
        c.ChoreAssignee = v
    }

    return c, nil
}

// Handler: GET /getchore
func GetChoreHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
	fmt.Printf("In the getChoreFromGroup function")
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed. Only GET is supported", http.StatusMethodNotAllowed)
        return
    }

    // var request struct {
    //     GroupID string `json:"group_id"`
    // }

    // if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
    //     http.Error(w, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
    //     return
    // }

    // if request.GroupID == "" {
    //     http.Error(w, "group_id is required", http.StatusBadRequest)
    //     return
    // }

	
	groupID := r.URL.Query().Get("group_id")

	if groupID == "" {
		http.Error(w, "group_id is required", http.StatusBadRequest)
		return
	}

    chores, err := getChoreFromGroup(ctx, firestoreClient, groupID)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error fetching chores: %v", err), http.StatusInternalServerError)
        return
    }

    // Return all chores
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(chores)
}

func init() {
    ctx := context.Background()
    projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

    var err error
    firestoreClient, err = firestore.NewClient(ctx, projectID)
    if err != nil {
        log.Fatalf("Failed to initialize Firestore client: %v", err)
    }

    functions.HTTP("GetChoreHandler", GetChoreHandler)
}
