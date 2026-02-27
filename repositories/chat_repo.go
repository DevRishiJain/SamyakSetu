package repositories

import (
	"context"
	"time"

	"github.com/samyaksetu/backend/database"
	"github.com/samyaksetu/backend/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatRepository handles all database operations for chat messages.
type ChatRepository struct {
	db *database.MongoDB
}

// NewChatRepository creates a new ChatRepository instance.
func NewChatRepository(db *database.MongoDB) *ChatRepository {
	return &ChatRepository{db: db}
}

// SaveMessage inserts a single chat message into the database.
func (r *ChatRepository) SaveMessage(msg *models.ChatMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	msg.CreatedAt = time.Now()
	result, err := r.db.Collection("chat_messages").InsertOne(ctx, msg)
	if err != nil {
		return err
	}

	msg.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}
