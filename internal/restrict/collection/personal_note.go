package collection

import (
	"context"
	"fmt"

	"github.com/OpenSlides/openslides-autoupdate-service/internal/restrict/perm"
	"github.com/OpenSlides/openslides-autoupdate-service/pkg/datastore/dsfetch"
)

// PersonalNote handels restriction for the personal_node collection.
//
// The user can see a personal node, if personal_note/user_id is the same as the id of the requested user.
//
// The superadmin can not see personal_notes from other users.
//
// Mode A: The user can see the personal note.
type PersonalNote struct{}

// MeetingID returns the meetingID for the object.
func (p PersonalNote) MeetingID(ctx context.Context, ds *dsfetch.Fetch, id int) (int, bool, error) {
	meetingID, err := ds.PersonalNote_MeetingID(id).Value(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("getting meetingID: %w", err)
	}

	return meetingID, true, nil
}

// Modes returns the field restriction for each mode.
func (p PersonalNote) Modes(mode string) FieldRestricter {
	switch mode {
	case "A":
		return p.see
	}
	return nil
}

func (p PersonalNote) see(ctx context.Context, ds *dsfetch.Fetch, mperms *perm.MeetingPermission, personalNoteIDs ...int) ([]int, error) {
	return eachRelationField(ctx, ds.PersonalNote_UserID, personalNoteIDs, func(userID int, ids []int) ([]int, error) {
		if mperms.UserID() == userID {
			return ids, nil
		}
		return nil, nil
	})
}

// SuperAdmin restricts the super admin.
func (p PersonalNote) SuperAdmin(mode string) FieldRestricter {
	return p.Modes(mode)
}
