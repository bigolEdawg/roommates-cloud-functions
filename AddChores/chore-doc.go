package chores

type Chore struct {
    Title            string                 `firestore:"title"`
    // Notes            string                 `firestore:"notes,omitempty"`
    // Status           string                 `firestore:"status"` // open, in_progress, done, skipped
    // Priority         int                    `firestore:"priority"`
    // CreatedBy        string                 `firestore:"created_by"`
    // CreatedAt        interface{}            `firestore:"created_at"`  // set: firestore.ServerTimestamp
    // UpdatedAt        interface{}            `firestore:"updated_at"`  // set: firestore.ServerTimestamp
    // Assignees        []string               `firestore:"assignees"`
    // ClaimedBy        *string                `firestore:"claimed_by,omitempty"`
    // EstimatedMinutes int                    `firestore:"estimated_minutes"`

    // Schedule         map[string]interface{} `firestore:"schedule"`          // store rule/one_time fields
    // NextOccurrenceAt time.Time              `firestore:"next_occurrence_at"`
    // LastCompletedAt  *time.Time             `firestore:"last_completed_at,omitempty"`

    // Rotation         map[string]interface{} `firestore:"rotation"`          // mode, queue

    // Reminders        map[string]interface{} `firestore:"reminders"`         // enabled, offsets, channels
    // Snooze           map[string]interface{} `firestore:"snooze"`            // minutes

    // Completed        bool                   `firestore:"completed"`
    // CompletedAt      *time.Time             `firestore:"completed_at,omitempty"`
    // CompletedBy      *string                `firestore:"completed_by,omitempty"`
    // StreakCount      int                    `firestore:"streak_count"`
    // MissedCount      int                    `firestore:"missed_count"`

    // Tags             []string               `firestore:"tags,omitempty"`
    // Attachments      []map[string]string    `firestore:"attachments,omitempty"`
}
