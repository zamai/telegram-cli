package main

import (
	"context"
	"testing"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestListHistory(t *testing.T) {
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		if _, ok := req.(*tg.MessagesGetHistoryRequest); ok {
			return &tg.MessagesMessages{
				Messages: []tg.MessageClass{
					// API order is newest-first.
					&tg.Message{ID: 2, PeerID: &tg.PeerUser{UserID: 5}, Message: "second", Date: 20, Out: true},
					&tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: 5}, Message: "first", Date: 10},
				},
				Users: []tg.UserClass{&tg.User{ID: 5, Username: "alice"}},
			}, nil
		}
		return nil, errors.Errorf("unexpected request %T", req)
	})

	res, err := listHistory(context.Background(), api, &tg.InputPeerSelf{}, 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Messages) != 2 {
		t.Fatalf("got %d messages, want 2", len(res.Messages))
	}
	// Output should be chronological (oldest-first).
	if res.Messages[0].ID != 1 || res.Messages[1].ID != 2 {
		t.Errorf("order = %d,%d want 1,2", res.Messages[0].ID, res.Messages[1].ID)
	}
	if res.Messages[0].Text != "first" || !res.Messages[1].Out {
		t.Errorf("unexpected content: %+v", res.Messages)
	}
}

func TestListHistoryLimit(t *testing.T) {
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		if _, ok := req.(*tg.MessagesGetHistoryRequest); ok {
			return &tg.MessagesMessages{
				Messages: []tg.MessageClass{
					&tg.Message{ID: 3, PeerID: &tg.PeerUser{UserID: 5}, Date: 3},
					&tg.Message{ID: 2, PeerID: &tg.PeerUser{UserID: 5}, Date: 2},
					&tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: 5}, Date: 1},
				},
				Users: []tg.UserClass{&tg.User{ID: 5}},
			}, nil
		}
		return nil, errors.Errorf("unexpected request %T", req)
	})

	res, err := listHistory(context.Background(), api, &tg.InputPeerSelf{}, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Messages) != 2 {
		t.Fatalf("limit not respected: got %d", len(res.Messages))
	}
}

func TestMessageTimelineFromMessages(t *testing.T) {
	timeline := newMessageTimeline()
	msg := &tg.Message{
		ID:      7,
		PeerID:  &tg.PeerUser{UserID: 5},
		Message: "hello",
		Date:    70,
	}
	msg.SetFromID(&tg.PeerUser{UserID: 5})
	res := timeline.FromMessages(
		[]tg.MessageClass{
			msg,
			&tg.MessageEmpty{},
		},
		entitiesOf([]tg.UserClass{&tg.User{ID: 5, Username: "alice", FirstName: "Alice"}}, nil),
	)

	if res.Peer.Username != "alice" {
		t.Fatalf("peer = %+v", res.Peer)
	}
	if len(res.Messages) != 1 {
		t.Fatalf("messages = %d, want 1", len(res.Messages))
	}
	if res.Messages[0].From == nil || res.Messages[0].From.Username != "alice" {
		t.Fatalf("from = %+v", res.Messages[0].From)
	}
	if res.Messages[0].Text != "hello" {
		t.Fatalf("text = %q", res.Messages[0].Text)
	}
}
