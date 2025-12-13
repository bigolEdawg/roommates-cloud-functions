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
    GroupName     string `json:"group_name"`
    GroupID       string `json:"group_id"`
    UserID        string `json:"user_id"`
}

// Get group data from a group
func getGroupFromMyGroups(ctx context.Context, client *firestore.Client, uid string) ([]groupData, error) {
    var groups []groupData

    iter := client.Collection("users").Doc(uid).Collection("my_groups").Documents(ctx)
	fmt.Printf("In the getGroupFromMyGroups function ")
    defer iter.Stop()

    for {
        doc, err := iter.Next()
        if err == iterator.Done {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("error reading groups for group %s: %v", uid, err)
        }

        g, err := getGroupDoc(doc)
        if err != nil {
            return nil, err
        }

        groups = append(groups, g)
    }

    return groups, nil
}

func getGroupDoc(doc *firestore.DocumentSnapshot) (groupData, error) {
    var g groupData

    data := doc.Data()

    if v, ok := data["user_id"].(string); ok {
        g.UserID = v
    }
    if v, ok := data["group_id"].(string); ok {
        g.GroupID = v
    }
    if v, ok := data["group_name"].(string); ok {
        g.GroupName = v
    }

    return g, nil
}

// Handler: GET /getgroup
func GetGroupHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
	fmt.Printf("In the getGroupFromMyGroups function")
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed. Only GET is supported", http.StatusMethodNotAllowed)
        return
    }

	// groupID := r.URL.Query().Get("group_id")

	// if groupID == "" {
	// 	http.Error(w, "group_id is required", http.StatusBadRequest)
	// 	return
	// }


    // temp uid implement auth header check
    uid := "" 
    groups, err := getGroupFromMyGroups(ctx, firestoreClient, uid)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error fetching groups: %v", err), http.StatusInternalServerError)
        return
    }

    // Return all groups
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(groups)
}

func init() {
    ctx := context.Background()
    projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

    var err error
    firestoreClient, err = firestore.NewClient(ctx, projectID)
    if err != nil {
        log.Fatalf("Failed to initialize Firestore client: %v", err)
    }

    functions.HTTP("GetGroupHandler", GetGroupHandler)
}
