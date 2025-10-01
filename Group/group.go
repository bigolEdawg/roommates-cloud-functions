package group

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

var firestoreClient *firestore.Client


// callerUID returns the acting user ID.
// Dev mode: if DEV_BYPASS_AUTH=1, use X-Dev-UID or DEV_DEFAULT_UID.
// Prod mode: require "Authorization: Bearer <token>" (plug Firebase here later).
// func callerUID(ctx context.Context, r *http.Request) (string, error) {
//   if os.Getenv("DEV_BYPASS_AUTH") == "1" {
//     if uid := r.Header.Get("X-Dev-UID"); uid != "" { return uid, nil }
//     if uid := os.Getenv("DEV_DEFAULT_UID"); uid != "" { return uid, nil }
//     return "", error.New("dev bypass enabled, but no X-Dev-UID or DEV_DEFAULT_UID provided")
//   }

//   // Production path: require an ID token (wire Firebase here later)
//   ah := r.Header.Get("Authorization")
//   if !strings.HasPrefix(ah, "Bearer ") {
//     return "", error.New("missing bearer token")
//   }
//   idToken := strings.TrimPrefix(ah, "Bearer ")

//   // TODO (when ready): verify with Firebase Admin:
//   // tok, err := authClient.VerifyIDToken(ctx, idToken)
//   // if err != nil { return "", err }
//   // return tok.UID, nil

//   // Temporary guard so you don't accidentally run "prod path" without verification:
//   return "", errors.New("token verification not configured")
// }


// Handler
func GroupHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
	// Extract the type of ID and the actual ID from the URL path
	pathSegments := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")

    // /group/
    // /group/entries
    // /group/status
    // /group/accept
    // /group/invite

    log.Print(len(pathSegments))
	log.Print(pathSegments)
	editType := pathSegments[0]
	if editType != "" {

        
        log.Printf("The edit type is %s", editType)

        switch{
			case editType == "invite":
				invite(ctx, w, r)
			case editType == "accept":
				acceptGroupInvite(ctx, w, r)
			default:
                http.Error(w, "Invalid resource. Refer to README.md for valid resources", http.StatusBadRequest)

        }
		return
	} else if editType == "" {  // check if we just have /group
        createGroupHandler(ctx, w, r)
    } else {
        http.Error(w, "Invalid endpoint. Expected format: /group or /group/{group_feature}", http.StatusBadRequest)
    }
}

// Initialize Firestore client and HTTP handler
func init() {
	ctx := context.Background()
	var err error
    projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	firestoreClient, err = firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}

	functions.HTTP("GroupHandler", GroupHandler)
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

