package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nouvadev/dropwise/internal/config"
	db "github.com/nouvadev/dropwise/internal/database/sqlc"
	"github.com/nouvadev/dropwise/internal/server/httputils"
)

// / ProcessDropsLogic contains the core logic for fetching and "sending" due drops.
// It now fetches distinct users with due drops and processes one drop per user.
// It returns the total number of drops processed and any critical error encountered during the overall process.
func ProcessDropsLogic(ctx context.Context, apiCfg *config.APIConfig) (totalProcessedCount int, err error) {
	log.Println("WorkerLogic: Starting batch processing for due drops.")
	totalProcessedCount = 0
	overallSuccess := true // Tracks if any non-critical error occurred

	// Step 1: Get all distinct user UUIDs with 'new' drops
	userUUIDs, err := apiCfg.DB.ListUserUUIDsWithDueDrops(ctx)
	if err != nil {
		log.Printf("WorkerLogic: Critical error fetching users with due drops: %v", err)
		return 0, fmt.Errorf("failed to fetch users with due drops: %w", err) // Stop if we can't get the user list
	}

	if len(userUUIDs) == 0 {
		log.Println("WorkerLogic: No users found with due drops at this time.")
		return 0, nil
	}

	log.Printf("WorkerLogic: Found %d distinct user identifier(s) with due drops.", len(userUUIDs))

	// Step 2: Loop through each user UUID
	for _, userUUID := range userUUIDs {
		if !userUUID.Valid {
			log.Println("WorkerLogic: Skipping invalid or empty user UUID from ListUserUUIDsWithDueDrops.")
			continue
		}
		currentUserUUID := userUUID

		log.Printf("WorkerLogic: Checking for due drops for user: %s", currentUserUUID.UUID.String())

		// Step 2a: Get one due drop for the current user
		getParams := db.GetDueDropsByUserUUIDParams{
			UserUuid: currentUserUUID,
			Limit:    1, // Process one drop per user per run
		}

		dueDrops, err := apiCfg.DB.GetDueDropsByUserUUID(ctx, getParams)
		if err != nil {
			log.Printf("WorkerLogic: Error fetching due drops for user %s: %v", currentUserUUID.UUID.String(), err)
			overallSuccess = false
			continue // Move to the next user
		}

		if len(dueDrops) == 0 {
			// This case should ideally not happen if ListUserUUIDsWithDueDrops returned this user,
			// but it's a good safeguard (e.g., if a drop was processed/deleted by another instance).
			log.Printf("WorkerLogic: No due drops found for user %s at this time (unexpected after listing).", currentUserUUID.UUID.String())
			continue // Move to the next user
		}

		// Process the first due drop found
		dueDrop := dueDrops[0]
		log.Printf("WorkerLogic: Found due drop for user %s: ID=%s, Topic='%s', URL='%s'",
			currentUserUUID.UUID.String(), dueDrop.ID.String(), dueDrop.Topic, dueDrop.Url)

		// Step 2b: Simulate sending the drop (placeholder for actual email logic)
		log.Printf("WorkerLogic: Simulating sending drop ID %s (Topic: %s) to user %s...", dueDrop.ID.String(), dueDrop.Topic, currentUserUUID.UUID.String())
		// In a real scenario, you might have a function like:
		// emailSent, err := emailService.SendDropReminder(currentUserID, dueDrop)
		// For now, we simulate success.
		time.Sleep(500 * time.Millisecond) // Reduced sleep time for faster batch processing simulation
		log.Printf("WorkerLogic: Drop ID %s (Topic: %s) 'sent' successfully to user %s (simulation).", dueDrop.ID.String(), dueDrop.Topic, currentUserUUID.UUID.String())

		// Step 2c: Mark the drop as sent
		markParams := db.MarkDropAsSentParams{
			ID:           dueDrop.ID,
			LastSentDate: sql.NullTime{Time: time.Now().UTC(), Valid: true}, // Use UTC for consistency
		}

		updatedDrop, err := apiCfg.DB.MarkDropAsSent(ctx, markParams)
		if err != nil {
			log.Printf("WorkerLogic: Error marking drop ID %s as sent for user %s: %v", dueDrop.ID.String(), currentUserUUID.UUID.String(), err)
			overallSuccess = false
			// Continue to next user, but this drop processing failed after "sending"
			continue
		}

		log.Printf("WorkerLogic: Successfully marked drop ID %s as sent for user %s. New status: %s, Send count: %d, Last sent: %v",
			updatedDrop.ID.String(), currentUserUUID.UUID.String(), updatedDrop.Status, updatedDrop.SendCount, updatedDrop.LastSentDate.Time)
		totalProcessedCount++
	}

	log.Printf("WorkerLogic: Batch processing finished. Total drops processed in this run: %d", totalProcessedCount)
	if !overallSuccess {
		log.Println("WorkerLogic: Some non-critical errors occurred during processing for one or more users/drops. Check logs for details.")
		// The function still returns nil for the error if it completed the loop,
		// as individual errors are logged and handled per user/drop.
		// A more sophisticated error aggregation could be added if needed for the caller.
	}
	return totalProcessedCount, nil
}

// ProcessDueDropsHTTP is an HTTP handler that triggers the drop processing logic.
// This function is suitable for use as a Google Cloud Function entry point.
func ProcessDueDropsHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet { // Cloud Scheduler might use GET or POST
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only GET or POST method is allowed")
		return
	}

	log.Println("WorkerHTTP: Received request to process due drops.")

	// It's crucial to initialize the database connection if it hasn't been already.
	// LoadConfig ensures GetDBQueries is called, which uses sync.Once for initialization.
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("WorkerHTTP: Error loading configuration: %v", err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Configuration error")
		return
	}

	// Ensure the database connection is closed eventually if this function is the sole manager.
	// However, for Cloud Functions, the global connection is typically managed across invocations.
	// If this were a standalone app, defer config.CloseDB() might be here.
	// For Cloud Functions, explicit closing is less critical as the environment manages instance lifecycle.

	processedCount, err := ProcessDropsLogic(r.Context(), cfg)
	if err != nil {
		// This error from ProcessDropsLogic is for critical failures (e.g., can't list users).
		// Individual drop processing errors are logged within ProcessDropsLogic but don't cause it to return an error.
		log.Printf("WorkerHTTP: Critical error during drop processing: %v", err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Critical error processing drops: "+err.Error())
		return
	}

	responseMessage := map[string]interface{}{
		"message":         "Drop processing finished.",
		"processed_count": processedCount,
	}
	log.Printf("WorkerHTTP: Finished processing. Drops processed in this invocation: %d", processedCount)
	httputils.RespondWithJSON(w, http.StatusOK, responseMessage)
}
