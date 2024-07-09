package vsphere

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/govmomi/vapi/tags"
)

func TestIfTagHasCategory(t *testing.T) {
	tests := []struct {
		name         string
		tags         []tags.Tag
		categoryId   string
		expectedBool bool
	}{
		{
			name: "Category Found",
			tags: []tags.Tag{
				{CategoryID: "category1"},
				{CategoryID: "category2"},
			},
			categoryId:   "category2",
			expectedBool: true,
		},
		{
			name: "Category Not Found",
			tags: []tags.Tag{
				{CategoryID: "category1"},
				{CategoryID: "category2"},
			},
			categoryId:   "category3",
			expectedBool: false,
		},
		{
			name:         "No Tags Provided",
			tags:         []tags.Tag{},
			categoryId:   "anyCategory",
			expectedBool: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &CloudDriver{}
			result := v.ifTagHasCategory(tt.tags, tt.categoryId)
			assert.Equal(t, tt.expectedBool, result)
		})
	}
}
