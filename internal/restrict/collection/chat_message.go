package collection

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-autoupdate-service/internal/restrict/perm"
	"github.com/OpenSlides/openslides-autoupdate-service/pkg/datastore/dsfetch"
)

// ChatMessage handels restrictions for the collection chat_message.
//
// A user can see a chat_message if they have the permission chat.can_manage
// or they are in the respective chat_group/read_group_ids (given by the key
// chat_group_id) or they have written the chat_message (dedicated by the key user_id).
//
// Mode A: A user can see a chat_message.
type ChatMessage struct{}

// MeetingID returns the meetingID for the object.
func (c ChatMessage) MeetingID(ctx context.Context, ds *dsfetch.Fetch, id int) (int, bool, error) {
	chatGroupID, err := ds.ChatMessage_ChatGroupID(id).Value(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("getting chat_group_id: %w", err)
	}

	return ChatGroup{}.MeetingID(ctx, ds, chatGroupID)
}

// Modes give sthe FieldRestricter for a restriction_mode.
func (c ChatMessage) Modes(mode string) FieldRestricter {
	switch mode {
	case "A":
		return c.see
	}
	return nil
}

func (ChatMessage) see(ctx context.Context, ds *dsfetch.Fetch, mperms *perm.MeetingPermission, chatMessageIDs ...int) ([]int, error) {
	return eachRelationField(ctx, ds.ChatMessage_ChatGroupID, chatMessageIDs, func(chatGroupID int, ids []int) ([]int, error) {
		meetingID, _, err := ChatGroup{}.MeetingID(ctx, ds, chatGroupID)
		if err != nil {
			return nil, fmt.Errorf("getting meeting id: %w", err)
		}

		perms, err := mperms.Meeting(ctx, meetingID)
		if err != nil {
			return nil, fmt.Errorf("getting permissions: %w", err)
		}

		if perms.Has(perm.ChatCanManage) {
			return ids, nil
		}

		readGroups, err := ds.ChatGroup_ReadGroupIDs(chatGroupID).Value(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting chat read group ids: %w", err)
		}

		for _, gid := range readGroups {
			if perms.InGroup(gid) {
				return ids, nil
			}
		}

		allowed, err := eachCondition(ids, func(chatMessageID int) (bool, error) {
			author, err := ds.ChatMessage_UserID(chatMessageID).Value(ctx)
			if err != nil {
				return false, fmt.Errorf("reading author of chat message: %w", err)
			}

			return author == mperms.UserID(), nil
		})

		if err != nil {
			return nil, fmt.Errorf("checking auther of chat message: %w", err)
		}

		return allowed, nil
	})
}
