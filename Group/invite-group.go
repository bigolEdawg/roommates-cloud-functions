package group

import (
	"context"
	"log"
	"net/http"
	"encoding/json"
	"fmt"
	// "os"
	// "strings"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"cloud.google.com/go/firestore"
	// "github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func invite(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var requestBody struct {
		GroupID		string `json:"group_id"`
		Invitee		string `json:"invitee"`
		UserID		string `json:"user_id"`
	}

	// Parse the incoming JSON body
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		http.Error(w, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
		return fmt.Errorf("error parsing request body: %v", err)
	}

	// Ensure the Invitee is provided
	if requestBody.Invitee == "" {
		log.Printf("Missing required field: invitee")
		http.Error(w, "Missing required field: invitee", http.StatusBadRequest)
		return fmt.Errorf("invitee is required")
	}
 
	// Check if the Invitee is already in the group contest
	groupRef := firestoreClient.Collection("groups").Doc(requestBody.GroupID).Collection("members")
	query := groupRef.Where("user_id", "==", requestBody.Invitee).Limit(1)

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Failed to query Firestore: %v", err)
		http.Error(w, fmt.Sprintf("Failed to query Firestore: %v", err), http.StatusInternalServerError)
		return fmt.Errorf("error querying Firestore: %v", err)
	}

	if len(docs) > 0 {
		// Invitee is already in the group
		log.Printf("Invitee %s is already in the group", requestBody.Invitee)
		http.Error(w, fmt.Sprintf("Invitee %s is already in the group", requestBody.Invitee), http.StatusBadRequest)
		return fmt.Errorf("invitee already in group")
	}

	// Invitee isn't in the group, proceed with the invite
	inviteRef := firestoreClient.Collection("users").Doc(requestBody.Invitee).Collection("group_invites").Doc(requestBody.GroupID)

	// Save the invitation
	log.Printf("Saving invite for user %s to group %s", requestBody.Invitee, requestBody.GroupID)
	_, err = inviteRef.Set(ctx, map[string]interface{}{
		"group_id":		requestBody.GroupID,
		"status":		"pending",
		"sent_at":		firestore.ServerTimestamp,
		"sent_from":	requestBody.UserID,
	})

	if err != nil {
		log.Printf("Failed to save invite: %v", err)
		http.Error(w, fmt.Sprintf("Failed to save invite: %v", err), http.StatusInternalServerError)
		return fmt.Errorf("error saving invite to Firestore: %v", err)
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Invite sent successfully",
		"group_id": requestBody.GroupID,
	})

	return nil
}


func acceptGroupInvite(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed; only POST is supported", http.StatusMethodNotAllowed)
		return fmt.Errorf("method not allowed")
	}
	// uid, err := callerUID(r.Context(), r)
	// if err != nil {
	// 	http.Error(w, "unauthorized", http.StatusUnauthorized)
	// 	return err
	// }	

	type reqBody struct {
		UserID   string `json:"user_id"`
		GroupID  string `json:"group_id"`
		Accepted bool   `json:"accepted"`
	}

	// Prefer the request context for cancellation/timeouts
	ctx = r.Context()
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req reqBody
	if err := dec.Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
		return err
	}
	if req.UserID == "" || req.GroupID == "" {
		http.Error(w, "user_id and group_id are required", http.StatusBadRequest)
		return fmt.Errorf("missing required fields")
	}

	inviteRef := firestoreClient.Collection("users").Doc(req.UserID).Collection("group_invites").Doc(req.GroupID)
	groupRef := firestoreClient.Collection("groups").Doc(req.GroupID)
	memberRef := groupRef.Collection("members").Doc(req.UserID)
	myGroupsRef := firestoreClient.Collection("users").Doc(req.UserID).Collection("my_groups").Doc(req.GroupID)
	userRef := firestoreClient.Collection("users").Doc(req.UserID)

	// Decline: just delete the invite and return.
	if !req.Accepted {
		_, err := inviteRef.Delete(ctx)
		// Treat not-found as success for idempotency
		if err != nil && status.Code(err) != codes.NotFound {
			http.Error(w, fmt.Sprintf("Failed to delete invite: %v", err), http.StatusInternalServerError)
			return err
		}
		writeJSON(w, map[string]string{
			"message":  fmt.Sprintf("User %s declined group %s", req.UserID, req.GroupID),
			"group_id": req.GroupID,
		})
		return nil
	}

	// Accept path: do it atomically and idempotently.
	err := firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// 1) Verify invite exists
		if _, err := tx.Get(inviteRef); err != nil {
			if status.Code(err) == codes.NotFound {
				// If invite missing but member already exists, treat as already accepted
				if mSnap, merr := tx.Get(memberRef); merr == nil && mSnap.Exists() {
					return nil
				}
				return fmt.Errorf("no invite found for user %s in group %s", req.UserID, req.GroupID)
			}
			return fmt.Errorf("failed reading invite: %w", err)
		}

		// 2) Fetch user_name (fallback to user_id)
		userName := req.UserID
		if uSnap, err := tx.Get(userRef); err == nil {
			if n, err := uSnap.DataAt("user_name"); err == nil {
				if s, ok := n.(string); ok && s != "" {
					userName = s
				}
			}
		}

		// 3) If already a member, just ensure mirrors are in place and delete invite
		if mSnap, err := tx.Get(memberRef); err == nil && mSnap.Exists() {
			_ = tx.Delete(inviteRef)
			if err := tx.Set(myGroupsRef, map[string]interface{}{
				"group_id":  "/groups/" + req.GroupID,
				"timestamp": firestore.ServerTimestamp,
			}, firestore.MergeAll); err != nil {
				return fmt.Errorf("failed to upsert my_groups: %w", err)
			}
			return nil
		}

		// 4) Add the member
		if err := tx.Set(memberRef, map[string]interface{}{
			"user_id":   req.UserID,
			"user_name": userName,
			"joined_at": firestore.ServerTimestamp,
			"added_by":  req.UserID, // or inviter if you track it
		}, firestore.MergeAll); err != nil {
			return fmt.Errorf("failed to add member: %w", err)
		}

		// 5) (Optional) increment a counter if you keep one
		// if err := tx.Update(groupRef, []firestore.Update{
		// 	{Path: "member_count", Value: firestore.Increment(1)},
		// }); err != nil {
		// 	return fmt.Errorf("failed to increment member_count: %w", err)
		// }

		// 6) Delete the invite (idempotent)
		if err := tx.Delete(inviteRef); err != nil && status.Code(err) != codes.NotFound {
			return fmt.Errorf("failed to delete invite: %w", err)
		}

		// 7) Mirror in user's my_groups (idempotent)
		if err := tx.Set(myGroupsRef, map[string]interface{}{
			"group_id":  "/groups/" + req.GroupID,
			"timestamp": firestore.ServerTimestamp,
		}, firestore.MergeAll); err != nil {
			return fmt.Errorf("failed to upsert my_groups: %w", err)
		}


		return nil
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to accept invite: %v", err), http.StatusBadRequest)
		return err
	}

	writeJSON(w, map[string]string{
		"message":  fmt.Sprintf("User %s accepted group %s", req.UserID, req.GroupID),
		"group_id": req.GroupID,
	})
	return nil
}

// helpers

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}












// // accept an invintation
// func acceptGroupInvite(ctx context.Context, w http.ResponseWriter, r *http.Request) (error){	
// 	var requestBody struct {
// 		UserID		string	`json:"user_id"`
// 		GroupID		string	`json:"group_id"`
// 		Accepted		bool	`json:"accepted"`
// 		// TrackID			string	`json:"track_id"`
// 	}
	
// 	defer r.Body.Close()
//     // Get the post request
//     if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
//         http.Error(w, fmt.Sprintf("Failed to parse request body: %v", err), http.StatusBadRequest)
//         return err
//     }

// 	var message string
// 	if requestBody.Accepted {
// 		err := member(ctx, requestBody.UserID, requestBody.GroupID)
// 		if err != nil {
// 			log.Printf("error saving User %s to groups %s: %v", requestBody.UserID, requestBody.GroupID, err)
// 			return fmt.Errorf("error saving User %s to groups %s: %v", requestBody.UserID, requestBody.GroupID, err)
// 		}

// 		// After the User makes their decision we can remove the invite
// 		_, err = firestoreClient.Collection("users").Doc(requestBody.UserID).Collection("group_invites").Doc(requestBody.GroupID).Delete(ctx)

// 		if err != nil {
// 			http.Error(w, fmt.Sprintf("Failed to remove invte to groups %s: %v", requestBody.GroupID, err), http.StatusInternalServerError)
// 			return err
// 		}
	
// 		if err := saveGroupIDToUserCollection(ctx, firestoreClient, requestBody.UserID, requestBody.GroupID); err != nil {
// 			http.Error(w, fmt.Sprintf("Failed to save to groups %s for user: %v", requestBody.GroupID, err), http.StatusInternalServerError)
// 			return err
// 		}
// 		message = "accepted"
// 		//fmt.Sprintf("set the status %s", status)
// 	} else {
// 		message = "rejected"
// 	}
	
// 	// Respond with success
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]string{
// 		"message":  fmt.Sprintf("User %s %s groups %s successfully",requestBody.UserID, message, requestBody.GroupID),
// 		"group_id": requestBody.GroupID,
// 	})


// 	log.Printf("User %s has %s groups %s", requestBody.UserID, message, requestBody.GroupID)
// 	return nil
// }


// // ------------------------------ Save Versus to User --------------------------------
// // Saves the groups ID in the User's "my_groups" subcollection
// func saveGroupIDToUserCollection(ctx context.Context, client *firestore.Client, UserID, GroupID string) error {
// 	// Reference to the user's "Myuser" subcollection with GroupID as the document ID
// 	UserRef := client.Collection("users").Doc(UserID).Collection("my_groups").Doc(GroupID)
// 	GroupID = fmt.Sprintf("/groups/%s", GroupID)

// 	_, err := UserRef.Set(ctx, map[string]interface{}{
// 		"group_id": GroupID,
// 		"timestamp": firestore.ServerTimestamp,
// 	})
// 	if err != nil {
// 		log.Printf("failed to add groups %s for User %s: %v", GroupID, UserID, err)
// 		return fmt.Errorf("failed to add groups %s for User %s: %v", GroupID, UserID, err)
// 	}
// 	return nil
// }


// func member(ctx context.Context, UserID, GroupID string) (error){
// 	UserRef := firestoreClient.Collection("users").Doc(UserID)
// 	UserDoc, err :=  UserRef.Get(ctx)
// 	if err != nil {
// 		log.Printf("failed to fetch User data for User %s: %v", UserID, err)
// 		return fmt.Errorf("failed to fetch User data for User %s: %v", UserID, err)
// 	}
	
// 	UserData := UserDoc.Data()
// 	UserName, UserNameExist := UserData["user_name"].(string)

// 	if !UserNameExist {
// 		log.Printf("Username not found for UserID: %s", UserID)
// 		return fmt.Errorf("username not found for UserID: %s", UserID)
// 	}


// 	// Save the member
// 	memberRef := firestoreClient.Collection("groups").Doc(GroupID).Collection("members").Doc(UserID)
// 	GroupID = fmt.Sprintf("/groups/%s", GroupID)

// 	// Use Set() to add the groups ID and timestamp
// 	_, err = memberRef.Set(ctx, map[string]interface{}{
// 		"user_id":			UserID,
// 		"user_name":		UserName,
// 	})

// 	if err != nil {
// 		log.Printf("failed to add member %s to groups %s: %v", UserID, GroupID, err)
// 		return fmt.Errorf("failed to add member %s to groups %s: %v", UserID, GroupID, err)
// 	}

// 	log.Printf("Successfully added member %s to groups %s", UserID, GroupID)
// 	return nil
// }

