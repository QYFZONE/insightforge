package sqlite

import (
	"insightforge/internal/domain/session"
	dbmodel "insightforge/internal/infra/store/sqlite/model"
)

// toSessionModel 把领域 Session 转成数据库模型。
func toSessionModel(item session.Session) dbmodel.Session {
	return dbmodel.Session{
		ID:        item.ID,
		Topic:     item.Topic,
		Status:    string(item.Status),
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

// toSession 把数据库模型转回领域 Session。
func toSession(record dbmodel.Session) session.Session {
	return session.Session{
		ID:        record.ID,
		Topic:     record.Topic,
		Status:    session.Status(record.Status),
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}
}

// toEventModel 把领域事件和已经序列化的 payload 转成数据库模型。
func toEventModel(event session.Event, payload string) dbmodel.Event {
	return dbmodel.Event{
		ID:        event.ID,
		SessionID: event.SessionID,
		Type:      event.Type,
		Message:   event.Message,
		Payload:   payload,
		CreatedAt: event.CreatedAt,
	}
}

// toEvent 把数据库事件转回领域事件，并反序列化 payload。
func toEvent(record dbmodel.Event) (session.Event, error) {
	payload, err := decodePayload(record.Payload)
	if err != nil {
		return session.Event{}, err
	}
	return session.Event{
		ID:        record.ID,
		SessionID: record.SessionID,
		Type:      record.Type,
		Message:   record.Message,
		Payload:   payload,
		CreatedAt: record.CreatedAt,
	}, nil
}
