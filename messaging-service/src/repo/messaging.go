package repo

import (
	"context"
	"messaging-service/src/types/records"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *Repo) SaveRoom(room *records.Room) error {
	ctx := context.Background()
	now := time.Now()
	room.CreatedAt = now
	room.UpdatedAt = now
	room.DeletedAt = nil

	_, err := r.rooms.InsertOne(ctx, room)
	if err != nil {
		return err
	}
	for _, m := range room.Members {
		if m == nil {
			continue
		}
		m.CreatedAt = now
		m.UpdatedAt = now
		m.DeletedAt = nil
		m.RoomUUID = room.UUID
		if _, err := r.members.InsertOne(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) LeaveRoom(userUUID string, roomUUID string) error {
	ctx := context.Background()
	now := time.Now()
	_, err := r.members.UpdateOne(ctx,
		bson.M{"user_uuid": userUUID, "room_uuid": roomUUID, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now, "updated_at": now}},
	)
	return err
}

func (r *Repo) GetMembersByRoomUUID(roomUUID string) ([]*records.Member, error) {
	ctx := context.Background()
	cur, err := r.members.Find(ctx, bson.M{"room_uuid": roomUUID, "deleted_at": nil})
	if err != nil {
		return nil, err
	}
	var out []*records.Member
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repo) GetRoomByRoomUUID(roomUUID string) (*records.Room, error) {
	ctx := context.Background()
	var result records.Room
	err := r.rooms.FindOne(ctx, bson.M{"uuid": roomUUID, "deleted_at": nil}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	members, err := r.GetMembersByRoomUUID(roomUUID)
	if err != nil {
		return nil, err
	}
	result.Members = members

	cur, err := r.messages.Find(ctx, bson.M{"room_uuid": roomUUID, "deleted_at": nil},
		options.Find().SetSort(bson.D{{Key: "created_at_nano", Value: -1}, {Key: "_id", Value: -1}}))
	if err != nil {
		return nil, err
	}
	var msgs []*records.Message
	if err := cur.All(ctx, &msgs); err != nil {
		return nil, err
	}
	result.Messages = msgs
	return &result, nil
}

func (r *Repo) SaveMessage(msg *records.Message) error {
	ctx := context.Background()
	now := time.Now()
	msg.CreatedAt = now
	msg.UpdatedAt = now
	msg.DeletedAt = nil
	_, err := r.messages.InsertOne(ctx, msg)
	return err
}

func (r *Repo) GetMessagesByRoomUUID(roomUUID string, offset int) ([]*records.Message, error) {
	ctx := context.Background()
	cur, err := r.messages.Find(ctx,
		bson.M{"room_uuid": roomUUID, "deleted_at": nil},
		options.Find().
			SetSort(bson.D{{Key: "created_at_nano", Value: -1}, {Key: "_id", Value: -1}}).
			SetSkip(int64(offset)).
			SetLimit(PAGINATION_MESSAGES),
	)
	if err != nil {
		return nil, err
	}
	var results []*records.Message
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *Repo) GetMessagesByRoomUUIDs(roomUUIDs string, offset int) ([]*records.Message, error) {
	parts := strings.Split(roomUUIDs, ",")
	uuids := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			uuids = append(uuids, p)
		}
	}
	if len(uuids) == 0 {
		return nil, nil
	}
	ctx := context.Background()
	cur, err := r.messages.Find(ctx,
		bson.M{"room_uuid": bson.M{"$in": uuids}, "deleted_at": nil},
		options.Find().SetSkip(int64(offset)),
	)
	if err != nil {
		return nil, err
	}
	var results []*records.Message
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *Repo) GetRoomsByUserUUIDForSubscribing(userUUID string) ([]*records.Room, error) {
	ctx := context.Background()
	cur, err := r.members.Find(ctx, bson.M{"user_uuid": userUUID, "deleted_at": nil})
	if err != nil {
		return nil, err
	}
	var mems []*records.Member
	if err := cur.All(ctx, &mems); err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	var rooms []*records.Room
	for _, m := range mems {
		if _, ok := seen[m.RoomUUID]; ok {
			continue
		}
		seen[m.RoomUUID] = struct{}{}
		room, err := r.getRoomDocByUUID(ctx, m.RoomUUID)
		if err != nil || room == nil {
			continue
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func (r *Repo) getRoomDocByUUID(ctx context.Context, roomUUID string) (*records.Room, error) {
	var room records.Room
	err := r.rooms.FindOne(ctx, bson.M{"uuid": roomUUID, "deleted_at": nil}).Decode(&room)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &room, nil
}

type roomSortEntry struct {
	room   *records.Room
	sortNs float64
}

func (r *Repo) GetRoomsByUserUUID(userUUID string, offset int) ([]*records.Room, error) {
	ctx := context.Background()
	cur, err := r.members.Find(ctx, bson.M{"user_uuid": userUUID, "deleted_at": nil})
	if err != nil {
		return nil, err
	}
	var mems []*records.Member
	if err := cur.All(ctx, &mems); err != nil {
		return nil, err
	}
	roomUUIDs := make(map[string]struct{})
	for _, m := range mems {
		roomUUIDs[m.RoomUUID] = struct{}{}
	}

	var entries []roomSortEntry
	for ru := range roomUUIDs {
		room, err := r.getRoomDocByUUID(ctx, ru)
		if err != nil || room == nil {
			continue
		}
		sortNs := room.CreatedAtNano
		var latest records.Message
		err = r.messages.FindOne(ctx,
			bson.M{"room_uuid": ru, "deleted_at": nil},
			options.FindOne().SetSort(bson.D{{Key: "created_at_nano", Value: -1}, {Key: "_id", Value: -1}}),
		).Decode(&latest)
		if err == nil {
			sortNs = latest.CreatedAtNano
		}
		entries = append(entries, roomSortEntry{room: room, sortNs: sortNs})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].sortNs > entries[j].sortNs
	})

	if offset >= len(entries) {
		return []*records.Room{}, nil
	}
	end := offset + PAGINATION_ROOMS
	if end > len(entries) {
		end = len(entries)
	}
	page := entries[offset:end]

	results := make([]*records.Room, 0, len(page))
	for _, e := range page {
		ru := e.room.UUID
		members, err := r.GetMembersByRoomUUID(ru)
		if err != nil {
			return nil, err
		}
		e.room.Members = members

		var latestMsg records.Message
		err = r.messages.FindOne(ctx,
			bson.M{"room_uuid": ru, "deleted_at": nil},
			options.FindOne().SetSort(bson.D{{Key: "created_at_nano", Value: -1}, {Key: "_id", Value: -1}}),
		).Decode(&latestMsg)
		if err == nil {
			lm := latestMsg
			e.room.Messages = []*records.Message{&lm}
		} else {
			e.room.Messages = nil
		}
		results = append(results, e.room)
	}

	sort.Slice(results, func(i, j int) bool {
		iNano := results[i].CreatedAtNano
		jNano := results[j].CreatedAtNano
		if len(results[i].Messages) > 0 {
			iNano = results[i].Messages[0].CreatedAtNano
		}
		if len(results[j].Messages) > 0 {
			jNano = results[j].Messages[0].CreatedAtNano
		}
		return iNano > jNano
	})

	return results, nil
}

func (r *Repo) DeleteRoom(roomUUID string) error {
	ctx := context.Background()
	now := time.Now()

	_, err := r.messages.UpdateMany(ctx,
		bson.M{"room_uuid": roomUUID, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now, "updated_at": now}},
	)
	if err != nil {
		return err
	}
	_, err = r.members.UpdateMany(ctx,
		bson.M{"room_uuid": roomUUID, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now, "updated_at": now}},
	)
	if err != nil {
		return err
	}
	_, err = r.rooms.UpdateOne(ctx,
		bson.M{"uuid": roomUUID, "deleted_at": nil},
		bson.M{"$set": bson.M{"deleted_at": now, "updated_at": now}},
	)
	return err
}
