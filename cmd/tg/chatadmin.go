package main

import (
	"context"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

// chatAdmin owns Telegram's split between basic groups and channels for chat
// administration commands.
type chatAdmin struct {
	api *tg.Client
}

func newChatAdmin(api *tg.Client) chatAdmin {
	return chatAdmin{api: api}
}

func (c chatAdmin) SetTitle(ctx context.Context, p peers.Peer, title string) error {
	var err error
	switch v := p.(type) {
	case peers.Channel:
		_, err = c.api.ChannelsEditTitle(ctx, &tg.ChannelsEditTitleRequest{
			Channel: v.InputChannel(),
			Title:   title,
		})
	case peers.Chat:
		_, err = c.api.MessagesEditChatTitle(ctx, &tg.MessagesEditChatTitleRequest{
			ChatID: v.ID(),
			Title:  title,
		})
	default:
		return errors.New("peer is not a group or channel")
	}
	if err != nil {
		return errors.Wrap(err, "edit title")
	}
	return nil
}

func (c chatAdmin) Leave(ctx context.Context, p peers.Peer) error {
	var err error
	switch v := p.(type) {
	case peers.Channel:
		_, err = c.api.ChannelsLeaveChannel(ctx, v.InputChannel())
	case peers.Chat:
		_, err = c.api.MessagesDeleteChatUser(ctx, &tg.MessagesDeleteChatUserRequest{
			ChatID: v.ID(),
			UserID: &tg.InputUserSelf{},
		})
	default:
		return errors.New("peer is not a group or channel")
	}
	if err != nil {
		return errors.Wrap(err, "leave")
	}
	return nil
}

func (c chatAdmin) SetPhoto(ctx context.Context, p peers.Peer, photo tg.InputChatPhotoClass) error {
	var err error
	switch v := p.(type) {
	case peers.Channel:
		_, err = c.api.ChannelsEditPhoto(ctx, &tg.ChannelsEditPhotoRequest{
			Channel: v.InputChannel(),
			Photo:   photo,
		})
	case peers.Chat:
		_, err = c.api.MessagesEditChatPhoto(ctx, &tg.MessagesEditChatPhotoRequest{
			ChatID: v.ID(),
			Photo:  photo,
		})
	default:
		return errors.New("peer is not a group or channel")
	}
	if err != nil {
		return errors.Wrap(err, "edit photo")
	}
	return nil
}

func (c chatAdmin) Invite(ctx context.Context, p peers.Peer, users []tg.InputUserClass) error {
	var err error
	switch v := p.(type) {
	case peers.Channel:
		_, err = c.api.ChannelsInviteToChannel(ctx, &tg.ChannelsInviteToChannelRequest{
			Channel: v.InputChannel(),
			Users:   users,
		})
	case peers.Chat:
		for _, u := range users {
			if _, err = c.api.MessagesAddChatUser(ctx, &tg.MessagesAddChatUserRequest{
				ChatID: v.ID(),
				UserID: u,
			}); err != nil {
				break
			}
		}
	default:
		return errors.New("peer is not a group or channel")
	}
	if err != nil {
		return errors.Wrap(err, "invite")
	}
	return nil
}

func (c chatAdmin) CurrentInviteLink(ctx context.Context, p peers.Peer) (string, error) {
	var (
		full *tg.MessagesChatFull
		err  error
	)
	switch v := p.(type) {
	case peers.Channel:
		full, err = c.api.ChannelsGetFullChannel(ctx, v.InputChannel())
	case peers.Chat:
		full, err = c.api.MessagesGetFullChat(ctx, v.ID())
	default:
		return "", errors.New("peer is not a group or channel")
	}
	if err != nil {
		return "", errors.Wrap(err, "get full chat")
	}
	return inviteLinkFromFull(full)
}
