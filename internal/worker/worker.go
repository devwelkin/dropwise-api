package worker

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/twomotive/dropwise/internal/config"
	db "github.com/twomotive/dropwise/internal/database/sqlc"
)

// ProcessDueDrops simulates the worker process for fetching and "sending" due drops.
func ProcessDueDrops(apiCfg *config.APIConfig) {
	ctx := context.Background()
	userID := "default-user" // For MVP, using the default user.
	// In a real system, you might iterate through all users or users with specific settings.

	log.Printf("Worker: Checking for due drops for user: %s", userID)

	// Parameters to fetch one due drop for the user
	getParams := db.GetDueDropsForUserParams{
		UserID: sql.NullString{String: userID, Valid: true},
		Limit:  1, // Process one drop at a time per user for this simulation
	}

	dueDrops, err := apiCfg.DB.GetDueDropsForUser(ctx, getParams)
	if err != nil {
		log.Printf("Worker: Error fetching due drops for user %s: %v", userID, err)
		return
	}

	if len(dueDrops) == 0 {
		log.Printf("Worker: No due drops found for user %s at this time.", userID)
		return
	}

	// Process the first due drop found (since Limit is 1)
	dueDrop := dueDrops[0]
	log.Printf("Worker: Found due drop for user %s: ID=%s, Topic='%s', URL='%s'",
		userID, dueDrop.ID.String(), dueDrop.Topic, dueDrop.Url)

	// Simulate sending the drop
	// In a real worker, this is where you'd integrate with an email service or other notification mechanism.
	log.Printf("Worker: Simulating sending drop ID %s (Topic: %s) to user %s...", dueDrop.ID.String(), dueDrop.Topic, userID)
	// Simulate some processing time or actual sending logic
	time.Sleep(1 * time.Second) // Placeholder for actual send operation
	log.Printf("Worker: Drop ID %s (Topic: %s) 'sent' successfully (simulation).", dueDrop.ID.String(), dueDrop.Topic)

	// Mark the drop as sent in the database
	markParams := db.MarkDropAsSentParams{
		ID:           dueDrop.ID,
		LastSentDate: sql.NullTime{Time: time.Now(), Valid: true},
	}

	updatedDrop, err := apiCfg.DB.MarkDropAsSent(ctx, markParams)
	if err != nil {
		log.Printf("Worker: Error marking drop ID %s as sent for user %s: %v", dueDrop.ID.String(), userID, err)
		return
	}

	log.Printf("Worker: Successfully marked drop ID %s as sent. New status: %s, Send count: %d, Last sent: %v",
		updatedDrop.ID.String(), updatedDrop.Status, updatedDrop.SendCount, updatedDrop.LastSentDate.Time)
}
