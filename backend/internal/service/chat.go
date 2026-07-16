package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const MaxMessageLength = 2000

var (
	ErrNotFriends = errors.New("you can only message friends")
	ErrEmptyMessage = errors.New("message cannot be empty")
	ErrMessageTooLong = errors.New("message is too long")
)

type ChatService struct {
	messages repository.MessageRepository
	relationships repository.RelationshipRepository
}

func NewChatService(messages repository.MessageRepository, relationships repository.RelationshipRepository) *ChatService {
	return &ChatService{messages: messages, relationships: relationships}
}

func (s *ChatService) areFriends(ctx context.Context, user1, user2 uuid.UUID) error {
	rel, err := s.relationships.Find(ctx, user1, user2)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFriends
		}
		return err
	}
	if rel.Status != model.RelationshipAccepted {
		return ErrNotFriends
	}
	return nil
}

func (s *ChatService) SendMessage(ctx context.Context, senderID, recipientID uuid.UUID, body string) (*model.Message, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, ErrEmptyMessage
	}
	if len(body) > MaxMessageLength {
		return nil, ErrMessageTooLong
	}
	if err := s.areFriends(ctx, senderID, recipientID); err != nil {
		return nil, err
	}
	msg := &model.Message{
		ID: uuid.New(),
		SenderID: senderID,
		RecipientID: recipientID,
		Body: body,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.messages.Create(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *ChatService) GetConversation(ctx context.Context, user1, user2 uuid.UUID, limit int) ([]model.Message, error) {
	if err := s.areFriends(ctx, user1, user2); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.messages.ListConversation(ctx, user1, user2, limit)
}
 func (s *ChatService) MarkRead(ctx context.Context, user1, user2 uuid.UUID) error {
	return s.messages.MarkRead(ctx, user1, user2)
}

func (s *ChatService) UnreadCounts(ctx context.Context, userID uuid.UUID) ([]model.UnreadCount, error) {
	return s.messages.UnreadCounts(ctx, userID)
}
