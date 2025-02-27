package collection_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/OpenSlides/openslides-autoupdate-service/internal/restrict/collection"
	"github.com/OpenSlides/openslides-autoupdate-service/internal/restrict/perm"
	"github.com/OpenSlides/openslides-autoupdate-service/pkg/datastore/dsfetch"
	"github.com/OpenSlides/openslides-autoupdate-service/pkg/datastore/dskey"
	"github.com/OpenSlides/openslides-autoupdate-service/pkg/datastore/dsmock"
	"github.com/OpenSlides/openslides-autoupdate-service/pkg/set"
)

type testData struct {
	name          string
	data          map[dskey.Key][]byte
	expect        []int
	expectOne     bool
	requestUserID int
	elementIDs    []int
}

func testCase(name string, t *testing.T, f collection.FieldRestricter, expect bool, yaml string, op ...testCaseOption) {
	t.Helper()
	td := testData{
		name:          name,
		expect:        nil,
		expectOne:     expect,
		data:          dsmock.YAMLData(yaml),
		requestUserID: 1,
		elementIDs:    []int{1},
	}

	for _, o := range op {
		o(&td)
	}

	if td.requestUserID != 0 {
		userIDKey, err := dskey.FromString(fmt.Sprintf("user/%d/id", td.requestUserID))
		if err != nil {
			t.Fatalf("invalid key %v", fmt.Sprintf("user/%d/id", td.requestUserID))
		}

		td.data[userIDKey] = []byte(strconv.Itoa(td.requestUserID))
	}

	td.test(t, f)
}

func testCaseMulti(name string, t *testing.T, f collection.FieldRestricter, ids, expect []int, yaml string, op ...testCaseOption) {
	t.Helper()
	td := testData{
		name:          name,
		expect:        expect,
		data:          dsmock.YAMLData(yaml),
		requestUserID: 1,
		elementIDs:    ids,
	}

	for _, o := range op {
		o(&td)
	}

	userIDKey := dskey.MustKey(fmt.Sprintf("user/%d/id", td.requestUserID))

	td.data[userIDKey] = []byte(strconv.Itoa(td.requestUserID))

	td.test(t, f)
}

func (tt testData) test(t *testing.T, f collection.FieldRestricter) {
	t.Helper()

	t.Run(tt.name, func(t *testing.T) {
		t.Helper()
		ds := dsfetch.New(dsmock.Stub(tt.data))
		perms := perm.NewMeetingPermission(ds, tt.requestUserID)

		allowedIDs, err := f(context.Background(), ds, perms, tt.elementIDs...)
		if err != nil {
			t.Fatalf("restriction mode returned unexpected error: %v", err)
		}

		if tt.expect == nil {
			// test for one value
			got := len(allowedIDs) == 1

			if got != tt.expectOne {
				t.Errorf("restriction mode returned %t, expected %t", got, tt.expectOne)
			}
			return
		}

		if !set.Equal(set.New(allowedIDs...), set.New(tt.expect...)) {
			t.Errorf("restriction mode returned %v, expected %v", allowedIDs, tt.expect)
		}
	})
}

type testCaseOption func(*testData)

// withPerms uses the group X337 to add permissions to the request user in the given
// meeting. X is the meetingID.
//
// Make sure to call withRequestUser before withPerms.
func withPerms(meetingID int, perms ...perm.TPermission) testCaseOption {
	return func(td *testData) {
		permString := "["
		for _, p := range perms {
			permString += fmt.Sprintf("%q,", p)
		}
		permString = permString[:len(permString)-1] + "]"
		groupID := 1000 + 337

		groupsKey := dskey.MustKey(fmt.Sprintf("user/1/group_$%d_ids", meetingID))
		groupIDKey := dskey.MustKey(fmt.Sprintf("group/%d/id", groupID))
		groupPermissionKey := dskey.MustKey(fmt.Sprintf("group/%d/permissions", groupID))
		meetingIDKey := dskey.MustKey(fmt.Sprintf("meeting/%d/id", meetingID))

		td.data[groupsKey] = jsonAppend(td.data[groupsKey], groupID)
		td.data[groupIDKey] = []byte(strconv.Itoa(groupID))
		td.data[groupPermissionKey] = []byte(permString)
		td.data[meetingIDKey] = []byte(strconv.Itoa(meetingID))
	}
}

func withRequestUser(userID int) testCaseOption {
	return func(td *testData) {
		td.requestUserID = userID
	}
}

func withElementID(id int) testCaseOption {
	return func(td *testData) {
		td.elementIDs = []int{id}
	}
}

func jsonAppend(value []byte, element ...int) []byte {
	var list []int
	if value != nil {
		if err := json.Unmarshal([]byte(value), &list); err != nil {
			panic(err)
		}
	}
	list = append(list, element...)
	newValue, err := json.Marshal(list)
	if err != nil {
		panic(err)
	}
	return newValue

}
