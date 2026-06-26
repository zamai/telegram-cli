package main

import (
	"sort"

	"github.com/go-faster/errors"

	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

// messageTimeline owns projection from Telegram message/entity shapes into the
// stable CLI message results used by history, search, context, schedule, and watch.
type messageTimeline struct{}

func newMessageTimeline() messageTimeline {
	return messageTimeline{}
}

func (messageTimeline) FromMessages(msgs []tg.MessageClass, ent peer.Entities) historyResult {
	var out historyResult
	for _, mc := range msgs {
		msg, ok := mc.(*tg.Message)
		if !ok {
			continue
		}
		if out.Peer.Type == "" {
			out.Peer = describePeer(msg.PeerID, ent)
		}
		out.Messages = append(out.Messages, buildMessageItem(msg, ent))
	}
	return out
}

func (t messageTimeline) FromResponse(res tg.MessagesMessagesClass) (historyResult, error) {
	msgs, ent, err := messagesFrom(res)
	if err != nil {
		return historyResult{}, err
	}
	return t.FromMessages(msgs, ent), nil
}

func (messageTimeline) FromUpdate(msg *tg.Message, e tg.Entities) watchEvent {
	ent := peer.EntitiesFromUpdate(e)
	return watchEvent{Peer: describePeer(msg.PeerID, ent), Message: buildMessageItem(msg, ent)}
}

func (messageTimeline) Chronological(res historyResult) historyResult {
	sort.Slice(res.Messages, func(i, j int) bool {
		return res.Messages[i].ID < res.Messages[j].ID
	})
	return res
}

func messagesFrom(res tg.MessagesMessagesClass) ([]tg.MessageClass, peer.Entities, error) {
	mod, ok := res.AsModified()
	if !ok {
		return nil, peer.Entities{}, errors.Errorf("unexpected messages type %T", res)
	}
	return mod.GetMessages(), entitiesOf(mod.GetUsers(), mod.GetChats()), nil
}
