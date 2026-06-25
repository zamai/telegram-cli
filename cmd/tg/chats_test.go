package main

import (
	"context"
	"testing"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func dialogsResult(dialogs []tg.DialogClass, messages []tg.MessageClass, users []tg.UserClass) func(bin.Encoder) (bin.Encoder, error) {
	return func(req bin.Encoder) (bin.Encoder, error) {
		if _, ok := req.(*tg.MessagesGetDialogsRequest); ok {
			return &tg.MessagesDialogs{Dialogs: dialogs, Messages: messages, Users: users}, nil
		}
		return nil, errors.Errorf("unexpected request %T", req)
	}
}

func dialogPage(n int) ([]tg.DialogClass, []tg.MessageClass, []tg.UserClass) {
	dialogs := make([]tg.DialogClass, 0, n)
	messages := make([]tg.MessageClass, 0, n)
	users := make([]tg.UserClass, 0, n)
	for i := 0; i < n; i++ {
		id := int64(i + 1)
		msgID := i + 1
		peer := &tg.PeerUser{UserID: id}
		dialogs = append(dialogs, &tg.Dialog{Peer: peer, TopMessage: msgID})
		messages = append(messages, &tg.Message{ID: msgID, PeerID: peer, Date: msgID})
		users = append(users, &tg.User{ID: id})
	}
	return dialogs, messages, users
}

func TestListChats(t *testing.T) {
	api := newFuncAPI(t, dialogsResult(
		[]tg.DialogClass{
			&tg.Dialog{
				Peer:        &tg.PeerUser{UserID: 42},
				TopMessage:  10,
				UnreadCount: 3,
				Pinned:      true,
			},
		},
		[]tg.MessageClass{
			&tg.Message{ID: 10, PeerID: &tg.PeerUser{UserID: 42}, Message: "hello there", Date: 123},
		},
		[]tg.UserClass{&tg.User{ID: 42, Username: "durov", FirstName: "Pavel"}},
	))

	list, err := listChats(context.Background(), api, nil, 100, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Chats) != 1 {
		t.Fatalf("got %d chats, want 1", len(list.Chats))
	}
	c := list.Chats[0]
	if c.Peer.ID != 42 || c.Peer.Username != "durov" {
		t.Errorf("peer = %+v", c.Peer)
	}
	if c.Unread != 3 || !c.Pinned {
		t.Errorf("unread/pinned = %d/%v", c.Unread, c.Pinned)
	}
	if c.LastMessage != "hello there" {
		t.Errorf("last = %q", c.LastMessage)
	}
}

func TestListChatsLimit(t *testing.T) {
	api := newFuncAPI(t, dialogsResult(
		[]tg.DialogClass{
			&tg.Dialog{Peer: &tg.PeerUser{UserID: 1}, TopMessage: 1},
			&tg.Dialog{Peer: &tg.PeerUser{UserID: 2}, TopMessage: 2},
		},
		[]tg.MessageClass{
			&tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: 1}, Date: 1},
			&tg.Message{ID: 2, PeerID: &tg.PeerUser{UserID: 2}, Date: 2},
		},
		[]tg.UserClass{&tg.User{ID: 1}, &tg.User{ID: 2}},
	))

	list, err := listChats(context.Background(), api, nil, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Chats) != 1 {
		t.Fatalf("limit not respected: got %d", len(list.Chats))
	}
}

func TestListChatsUsesLimitAsBatchSize(t *testing.T) {
	var requestLimits []int
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		r, ok := req.(*tg.MessagesGetDialogsRequest)
		if !ok {
			return nil, errors.Errorf("unexpected request %T", req)
		}
		requestLimits = append(requestLimits, r.Limit)
		dialogs, messages, users := dialogPage(r.Limit)
		return &tg.MessagesDialogs{Dialogs: dialogs, Messages: messages, Users: users}, nil
	})

	list, err := listChats(context.Background(), api, nil, 20, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Chats) != 20 {
		t.Fatalf("got %d chats, want 20", len(list.Chats))
	}
	if len(requestLimits) != 1 || requestLimits[0] != 20 {
		t.Fatalf("request limits = %v, want [20]", requestLimits)
	}
}

func TestListChatsCapsBatchSize(t *testing.T) {
	var requestLimits []int
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		r, ok := req.(*tg.MessagesGetDialogsRequest)
		if !ok {
			return nil, errors.Errorf("unexpected request %T", req)
		}
		requestLimits = append(requestLimits, r.Limit)
		dialogs, messages, users := dialogPage(r.Limit)
		return &tg.MessagesDialogsSlice{
			Dialogs:  dialogs,
			Messages: messages,
			Users:    users,
			Count:    250,
		}, nil
	})

	list, err := listChats(context.Background(), api, nil, 250, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Chats) != 250 {
		t.Fatalf("got %d chats, want 250", len(list.Chats))
	}
	if len(requestLimits) != 3 {
		t.Fatalf("request limits = %v, want 3 capped requests", requestLimits)
	}
	for _, got := range requestLimits {
		if got != maxDialogsBatchSize {
			t.Fatalf("request limits = %v, want capped at %d", requestLimits, maxDialogsBatchSize)
		}
	}
}

func TestListChatsZeroLimitDoesNotQuery(t *testing.T) {
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		return nil, errors.Errorf("unexpected request %T", req)
	})

	list, err := listChats(context.Background(), api, nil, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Chats) != 0 {
		t.Fatalf("got %d chats, want 0", len(list.Chats))
	}
}

func TestChatsListDefaultLimit(t *testing.T) {
	cmd := (&app{}).newChatsListCmd()
	flag := cmd.Flags().Lookup("limit")
	if flag == nil {
		t.Fatal("limit flag not found")
	}
	if flag.DefValue != "20" {
		t.Fatalf("default limit = %q, want 20", flag.DefValue)
	}
}
