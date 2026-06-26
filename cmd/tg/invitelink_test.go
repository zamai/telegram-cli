package main

import (
	"context"
	"testing"

	"github.com/go-faster/errors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

func TestInviteLinkFromFull(t *testing.T) {
	t.Run("channel with link", func(t *testing.T) {
		full := &tg.MessagesChatFull{FullChat: &tg.ChannelFull{}}
		full.FullChat.(*tg.ChannelFull).SetExportedInvite(&tg.ChatInviteExported{Link: "https://t.me/+abc"})
		got, err := inviteLinkFromFull(full)
		if err != nil {
			t.Fatal(err)
		}
		if got != "https://t.me/+abc" {
			t.Errorf("link = %q", got)
		}
	})

	t.Run("basic chat with link", func(t *testing.T) {
		full := &tg.MessagesChatFull{FullChat: &tg.ChatFull{}}
		full.FullChat.(*tg.ChatFull).SetExportedInvite(&tg.ChatInviteExported{Link: "https://t.me/+xyz"})
		got, err := inviteLinkFromFull(full)
		if err != nil {
			t.Fatal(err)
		}
		if got != "https://t.me/+xyz" {
			t.Errorf("link = %q", got)
		}
	})

	t.Run("no link", func(t *testing.T) {
		full := &tg.MessagesChatFull{FullChat: &tg.ChannelFull{}}
		got, err := inviteLinkFromFull(full)
		if err != nil {
			t.Fatal(err)
		}
		if got != "" {
			t.Errorf("expected empty link, got %q", got)
		}
	})
}

func TestChatAdminCurrentInviteLinkBasicChat(t *testing.T) {
	full := &tg.MessagesChatFull{FullChat: &tg.ChatFull{
		ID:             10,
		Participants:   &tg.ChatParticipantsForbidden{ChatID: 10},
		NotifySettings: tg.PeerNotifySettings{},
	}}
	full.FullChat.(*tg.ChatFull).SetExportedInvite(&tg.ChatInviteExported{Link: "https://t.me/+basic"})
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		if r, ok := req.(*tg.MessagesGetFullChatRequest); ok && r.ChatID == 10 {
			return full, nil
		}
		return nil, errors.Errorf("unexpected request %T", req)
	})
	m := peers.Options{}.Build(api)

	got, err := newChatAdmin(api).CurrentInviteLink(context.Background(), m.Chat(&tg.Chat{ID: 10}))
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://t.me/+basic" {
		t.Fatalf("link = %q", got)
	}
}

func TestChatAdminCurrentInviteLinkChannel(t *testing.T) {
	full := &tg.MessagesChatFull{FullChat: &tg.ChannelFull{
		ID:             10,
		ChatPhoto:      &tg.PhotoEmpty{},
		NotifySettings: tg.PeerNotifySettings{},
	}}
	full.FullChat.(*tg.ChannelFull).SetExportedInvite(&tg.ChatInviteExported{Link: "https://t.me/+channel"})
	api := newFuncAPI(t, func(req bin.Encoder) (bin.Encoder, error) {
		if r, ok := req.(*tg.ChannelsGetFullChannelRequest); ok {
			ch, ok := r.GetChannel().(*tg.InputChannel)
			if ok && ch.ChannelID == 10 {
				return full, nil
			}
		}
		return nil, errors.Errorf("unexpected request %T", req)
	})
	m := peers.Options{}.Build(api)

	got, err := newChatAdmin(api).CurrentInviteLink(context.Background(), m.Channel(&tg.Channel{ID: 10, AccessHash: 20}))
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://t.me/+channel" {
		t.Fatalf("link = %q", got)
	}
}
