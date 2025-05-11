package worker

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/twomotive/dropwise/internal/config"
	db "github.com/twomotive/dropwise/internal/database/sqlc"
	"github.com/twomotive/dropwise/internal/server/httputils"
)

// processDropsLogic contains the core logic for fetching and "sending" due drops.
// It returns the number of drops processed and any error encountered.
func ProcessDropsLogic(ctx context.Context, apiCfg *config.APIConfig) (processedCount int, err error) {
	userID := "default-user" // For MVP, using the default user.
	log.Printf("WorkerLogic: Checking for due drops for user: %s", userID)

	getParams := db.GetDueDropsForUserParams{
		UserID: sql.NullString{String: userID, Valid: true},
		Limit:  1, // Process one drop at a time per user
	}

	dueDrops, err := apiCfg.DB.GetDueDropsForUser(ctx, getParams)
	if err != nil {
		log.Printf("WorkerLogic: Error fetching due drops for user %s: %v", userID, err)
		return 0, err
	}

	if len(dueDrops) == 0 {
		log.Printf("WorkerLogic: No due drops found for user %s at this time.", userID)
		return 0, nil
	}

	// Process the first due drop found
	dueDrop := dueDrops[0]
	log.Printf("WorkerLogic: Found due drop for user %s: ID=%s, Topic='%s', URL='%s'",
		userID, dueDrop.ID.String(), dueDrop.Topic, dueDrop.Url)

	log.Printf("WorkerLogic: Simulating sending drop ID %s (Topic: %s) to user %s...", dueDrop.ID.String(), dueDrop.Topic, userID)
	time.Sleep(1 * time.Second) // Placeholder for actual send operation
	log.Printf("WorkerLogic: Drop ID %s (Topic: %s) 'sent' successfully (simulation).", dueDrop.ID.String(), dueDrop.Topic)

	markParams := db.MarkDropAsSentParams{
		ID:           dueDrop.ID,
		LastSentDate: sql.NullTime{Time: time.Now(), Valid: true},
	}

	updatedDrop, err := apiCfg.DB.MarkDropAsSent(ctx, markParams)
	if err != nil {
		log.Printf("WorkerLogic: Error marking drop ID %s as sent for user %s: %v", dueDrop.ID.String(), userID, err)
		return 0, err
	}

	log.Printf("WorkerLogic: Successfully marked drop ID %s as sent. New status: %s, Send count: %d, Last sent: %v",
		updatedDrop.ID.String(), updatedDrop.Status, updatedDrop.SendCount, updatedDrop.LastSentDate.Time)

	return 1, nil // Processed one drop
}

// ProcessDueDropsHTTP is an HTTP handler that triggers the drop processing logic.
// This function is suitable for use as a Google Cloud Function entry point.
func ProcessDueDropsHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet { // Cloud Scheduler might use GET or POST
		httputils.RespondWithError(w, http.StatusMethodNotAllowed, "Only GET or POST method is allowed")
		return
	}

	log.Println("WorkerHTTP: Received request to process due drops.")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("WorkerHTTP: Error loading configuration: %v", err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Configuration error")
		return
	}

	processedCount, err := ProcessDropsLogic(r.Context(), cfg)
	if err != nil {
		log.Printf("WorkerHTTP: Error processing drops: %v", err)
		httputils.RespondWithError(w, http.StatusInternalServerError, "Error processing drops: "+err.Error())
		return
	}

	responseMessage := map[string]interface{}{
		"message":         "Drop processing finished.",
		"processed_count": processedCount,
	}
	log.Printf("WorkerHTTP: Finished processing. Drops processed: %d", processedCount)
	httputils.RespondWithJSON(w, http.StatusOK, responseMessage)
}
