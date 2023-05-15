package mongoclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/x-xyz/goapi/base/ptr"
)

func TestMakeBsonM(t *testing.T) {
	type PatchableUser struct {
		Name    *string `bson:"name,omitempty"`
		Age     *int    `bson:"woman_secret,omitempty"`
		Address string  `bson:"address"`
		Bio     string  `bson:"bio"`
	}

	patchable := &PatchableUser{}
	patchable.Name = ptr.String("")
	patchable.Age = ptr.Int(10)
	patchable.Bio = "hey!yo!"

	updater, err := MakeBsonM(patchable)

	assert.NoError(t, err)
	assert.Equal(
		t,
		bson.M{
			"name":         "",
			"woman_secret": 10,
			// field address is empty, so ignore
			"bio": "hey!yo!",
		},
		updater,
	)
}
