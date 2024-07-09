package vsphere

import (
	"context"
	"net/url"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
)

// GetResourceTags returns a map of resource tags
func (v *CloudDriver) GetResourceTags(ctx context.Context, resourceType string) (map[string]tags.AttachedTags, error) {
	tags, err := v.getResourceTags(ctx, v.Client.Client, resourceType)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (v *CloudDriver) getTagsAndCategory(ctx context.Context, client *vim25.Client, resourceType, tagCategory string) (map[string]tags.AttachedTags, string, error) {
	categoryID, e := v.getCategoryID(ctx, client, tagCategory)
	if e != nil {
		return nil, "", e
	}

	if categoryID == "" {
		return nil, "", errors.Errorf("No tag with category type %s is created", tagCategory)
	}

	tags, e := v.getResourceTags(ctx, client, resourceType)
	if e != nil {
		return nil, "", e
	}
	if len(tags) == 0 {
		return nil, "", errors.Errorf("No tag is attached to resource %s", resourceType)
	}

	return tags, categoryID, e
}

func (v *CloudDriver) getTagManager(ctx context.Context, client *vim25.Client) (*tags.Manager, error) {
	c := rest.NewClient(client)
	err := c.Login(ctx, url.UserPassword(v.VCenterUsername, v.VCenterPassword))
	if err != nil {
		return nil, err
	}

	return tags.NewManager(c), nil
}

func (v *CloudDriver) getResourceTags(ctx context.Context, client *vim25.Client, resourceType string) (map[string]tags.AttachedTags, error) {
	t, err := v.getTagManager(ctx, client)
	if err != nil {
		return nil, err
	}
	m, err := view.NewManager(client).CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{resourceType}, true)
	if err != nil {
		return nil, err
	}

	resource, err := m.Find(ctx, []string{resourceType}, property.Match{})
	if err != nil {
		return nil, err
	}

	refs := make([]mo.Reference, len(resource))
	for i := range resource {
		refs[i] = resource[i]
	}
	attachedTags, err := t.GetAttachedTagsOnObjects(ctx, refs)
	if err != nil {
		return nil, err
	}

	tags := make(map[string]tags.AttachedTags)
	for _, t := range attachedTags {
		tags[t.ObjectID.Reference().Value] = t
	}
	return tags, nil
}

func (v *CloudDriver) getCategoryID(ctx context.Context, client *vim25.Client, name string) (string, error) {
	t, err := v.getTagManager(ctx, client)
	if err != nil {
		return "", err
	}
	categories, err := t.GetCategories(ctx)
	if err != nil {
		return "", err
	}
	for _, category := range categories {
		if category.Name == name {
			return category.ID, nil
		}
	}
	return "", nil
}

func (v *CloudDriver) ifTagHasCategory(tags []tags.Tag, categoryID string) bool {
	for _, tag := range tags {
		if tag.CategoryID == categoryID {
			return true
		}
	}
	return false
}
