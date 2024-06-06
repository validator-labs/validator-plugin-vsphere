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

func (v *VSphereCloudDriver) GetResourceTags(ctx context.Context, resourceType string) (map[string]tags.AttachedTags, error) {
	tags, err := v.getResourceTags(ctx, v.Client.Client, resourceType)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (v *VSphereCloudDriver) getTagsAndCategory(ctx context.Context, client *vim25.Client, resourceType, tagCategory string) (map[string]tags.AttachedTags, string, error) {
	categoryId, e := v.getCategoryId(ctx, client, tagCategory)
	if e != nil {
		return nil, "", e
	}

	if categoryId == "" {
		return nil, "", errors.Errorf("No tag with category type %s is created", tagCategory)
	}

	tags, e := v.getResourceTags(ctx, client, resourceType)
	if e != nil {
		return nil, "", e
	}
	if len(tags) == 0 {
		return nil, "", errors.Errorf("No tag is attached to resource %s", resourceType)
	}

	return tags, categoryId, e
}

func (v *VSphereCloudDriver) getTagManager(ctx context.Context, client *vim25.Client) (*tags.Manager, error) {
	c := rest.NewClient(client)
	err := c.Login(ctx, url.UserPassword(v.VCenterUsername, v.VCenterPassword))
	if err != nil {
		return nil, err
	}

	return tags.NewManager(c), nil
}

func (v *VSphereCloudDriver) getResourceTags(ctx context.Context, client *vim25.Client, resourceType string) (map[string]tags.AttachedTags, error) {
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

func (v *VSphereCloudDriver) getCategoryId(ctx context.Context, client *vim25.Client, name string) (string, error) {
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

func (v *VSphereCloudDriver) ifTagHasCategory(tags []tags.Tag, categoryId string) bool {
	for _, tag := range tags {
		if tag.CategoryID == categoryId {
			return true
		}
	}
	return false
}
