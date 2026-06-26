package main

import (
	"context"
	"testing"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

func TestInviteHash(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"https://t.me/+AbCdEf", "AbCdEf"},
		{"https://t.me/joinchat/XyZ", "XyZ"},
		{"t.me/+Hash123", "Hash123"},
		{"+Hash123", "Hash123"},
		{"PlainHash", "PlainHash"},
	} {
		if got := inviteHash(tc.in); got != tc.want {
			t.Errorf("inviteHash(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestChatAdminSetTitleBasicChat(t *testing.T) {
	var got *tg.MessagesEditChatTitleRequest
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		if r, ok := req.(*tg.MessagesEditChatTitleRequest); ok {
			got = r
			return &tg.Updates{}, nil
		}
		return nil, errors.Errorf("unexpected request %T", req)
	})
	m := peers.Options{}.Build(api)

	err := newChatAdmin(api).SetTitle(context.Background(), m.Chat(&tg.Chat{ID: 10}), "new name")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.ChatID != 10 || got.Title != "new name" {
		t.Fatalf("unexpected request: %+v", got)
	}
}

func TestChatAdminSetTitleChannel(t *testing.T) {
	var got *tg.ChannelsEditTitleRequest
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		if r, ok := req.(*tg.ChannelsEditTitleRequest); ok {
			got = r
			return &tg.Updates{}, nil
		}
		return nil, errors.Errorf("unexpected request %T", req)
	})
	m := peers.Options{}.Build(api)

	err := newChatAdmin(api).SetTitle(context.Background(), m.Channel(&tg.Channel{ID: 10, AccessHash: 20}), "new name")
	if err != nil {
		t.Fatal(err)
	}
	ch, ok := got.GetChannel().(*tg.InputChannel)
	if got == nil || !ok || ch.ChannelID != 10 || got.Title != "new name" {
		t.Fatalf("unexpected request: %+v", got)
	}
}

func TestChatAdminLeaveBasicChat(t *testing.T) {
	var got *tg.MessagesDeleteChatUserRequest
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		if r, ok := req.(*tg.MessagesDeleteChatUserRequest); ok {
			got = r
			return &tg.Updates{}, nil
		}
		return nil, errors.Errorf("unexpected request %T", req)
	})
	m := peers.Options{}.Build(api)

	err := newChatAdmin(api).Leave(context.Background(), m.Chat(&tg.Chat{ID: 10}))
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.ChatID != 10 {
		t.Fatalf("unexpected request: %+v", got)
	}
	if _, ok := got.UserID.(*tg.InputUserSelf); !ok {
		t.Fatalf("user = %T, want InputUserSelf", got.UserID)
	}
}

func TestChatAdminLeaveChannel(t *testing.T) {
	var got tg.InputChannelClass
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		if r, ok := req.(*tg.ChannelsLeaveChannelRequest); ok {
			got = r.Channel
			return &tg.Updates{}, nil
		}
		return nil, errors.Errorf("unexpected request %T", req)
	})
	m := peers.Options{}.Build(api)

	err := newChatAdmin(api).Leave(context.Background(), m.Channel(&tg.Channel{ID: 10, AccessHash: 20}))
	if err != nil {
		t.Fatal(err)
	}
	ch, ok := got.(*tg.InputChannel)
	if !ok || ch.ChannelID != 10 {
		t.Fatalf("unexpected channel: %+v", got)
	}
}

func TestChatAdminSetPhotoBasicChat(t *testing.T) {
	var got *tg.MessagesEditChatPhotoRequest
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		if r, ok := req.(*tg.MessagesEditChatPhotoRequest); ok {
			got = r
			return &tg.Updates{}, nil
		}
		return nil, errors.Errorf("unexpected request %T", req)
	})
	m := peers.Options{}.Build(api)
	photo := &tg.InputChatUploadedPhoto{}

	err := newChatAdmin(api).SetPhoto(context.Background(), m.Chat(&tg.Chat{ID: 10}), photo)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.ChatID != 10 || got.Photo != photo {
		t.Fatalf("unexpected request: %+v", got)
	}
}

func TestChatAdminSetPhotoChannel(t *testing.T) {
	var got *tg.ChannelsEditPhotoRequest
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		if r, ok := req.(*tg.ChannelsEditPhotoRequest); ok {
			got = r
			return &tg.Updates{}, nil
		}
		return nil, errors.Errorf("unexpected request %T", req)
	})
	m := peers.Options{}.Build(api)
	photo := &tg.InputChatUploadedPhoto{}

	err := newChatAdmin(api).SetPhoto(context.Background(), m.Channel(&tg.Channel{ID: 10, AccessHash: 20}), photo)
	if err != nil {
		t.Fatal(err)
	}
	ch, ok := got.GetChannel().(*tg.InputChannel)
	if got == nil || !ok || ch.ChannelID != 10 || got.Photo != photo {
		t.Fatalf("unexpected request: %+v", got)
	}
}

func TestChatAdminInviteBasicChat(t *testing.T) {
	var got *tg.MessagesAddChatUserRequest
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		if r, ok := req.(*tg.MessagesAddChatUserRequest); ok {
			got = r
			return &tg.MessagesInvitedUsers{Updates: &tg.Updates{}}, nil
		}
		return nil, errors.Errorf("unexpected request %T", req)
	})
	m := peers.Options{}.Build(api)
	user := &tg.InputUser{UserID: 99}

	err := newChatAdmin(api).Invite(context.Background(), m.Chat(&tg.Chat{ID: 10}), []tg.InputUserClass{user})
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.ChatID != 10 || got.UserID != user {
		t.Fatalf("unexpected request: %+v", got)
	}
}

func TestChatAdminInviteChannel(t *testing.T) {
	var got *tg.ChannelsInviteToChannelRequest
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		if r, ok := req.(*tg.ChannelsInviteToChannelRequest); ok {
			got = r
			return &tg.MessagesInvitedUsers{Updates: &tg.Updates{}}, nil
		}
		return nil, errors.Errorf("unexpected request %T", req)
	})
	m := peers.Options{}.Build(api)
	user := &tg.InputUser{UserID: 99}

	err := newChatAdmin(api).Invite(context.Background(), m.Channel(&tg.Channel{ID: 10, AccessHash: 20}), []tg.InputUserClass{user})
	if err != nil {
		t.Fatal(err)
	}
	ch, ok := got.GetChannel().(*tg.InputChannel)
	if got == nil || !ok || ch.ChannelID != 10 || len(got.Users) != 1 || got.Users[0] != user {
		t.Fatalf("unexpected request: %+v", got)
	}
}
