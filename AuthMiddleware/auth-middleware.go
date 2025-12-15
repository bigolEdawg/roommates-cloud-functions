package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
)

// ==================== GLOBAL VARIABLES ====================
var (
	authClient *auth.Client
	firestoreClient *firestore.Client
	ctx        = context.Background()
)

// ==================== FIREBASE INIT ====================
func initFirebase() {
	// For development with emulator
	os.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "localhost:9099")
	
	// Initialize Firebase without credentials (for emulator)
	app, err := firebase.NewApp(ctx, &firebase.Config{
		ProjectID: "roommates-app", // Your project ID
	}, option.WithoutAuthentication())
	
	if err != nil {
		log.Fatalf("Error initializing Firebase: %v\n", err)
	}

	// Get Auth client
	authClient, err = app.Auth(ctx)
	if err != nil {
		log.Fatalf("Error getting Auth client: %v\n", err)
	}

	// Get Firestore client
	firestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error getting Firestore client: %v\n", err)
	}

	log.Println("🔥 Firebase initialized (using emulator)")
}

// ==================== MIDDLEWARE ====================
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "No token"}`, http.StatusUnauthorized)
			return
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "Invalid token format"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Verify token
		decodedToken, err := authClient.VerifyIDToken(ctx, token)
		if err != nil {
			http.Error(w, `{"error": "Invalid token: `+err.Error()+`"}`, http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), "userID", decodedToken.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// ==================== HANDLERS ====================
// GET /api/user/groups - Get user's groups
func getUserGroups(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	// Query Firestore: users/{userID}/groups
	iter := firestoreClient.Collection("users").Doc(userID).Collection("groups").Documents(ctx)
	defer iter.Stop()

	var groups []map[string]interface{}
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}
		data := doc.Data()
		data["id"] = doc.Ref.ID
		groups = append(groups, data)
	}

	json.NewEncoder(w).Encode(groups)
}

// POST /api/user/groups - Create new group
func createGroup(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	// Parse request body
	var body struct {
		Name string `json:"name"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error": "Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Create group in Firestore
	groupRef := firestoreClient.Collection("users").Doc(userID).Collection("groups").NewDoc()
	
	_, err := groupRef.Set(ctx, map[string]interface{}{
		"name":      body.Name,
		"ownerId":   userID,
		"createdAt": time.Now(),
		"members":   []string{userID}, // User is first member
	})

	if err != nil {
		http.Error(w, `{"error": "Failed to create group"}`, http.StatusInternalServerError)
		return
	}

	// Return created group
	response := map[string]interface{}{
		"id":        groupRef.ID,
		"name":      body.Name,
		"ownerId":   userID,
		"createdAt": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ==================== MAIN ====================
func main() {
	// Initialize Firebase
	initFirebase()
	defer firestoreClient.Close()

	// Create router
	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Protected routes (require auth)
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/user/groups", authMiddleware(getUserGroups)).Methods("GET")
	api.HandleFunc("/user/groups", authMiddleware(createGroup)).Methods("POST")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}