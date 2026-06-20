package agentapi

import (
	"context"
	"fmt"

	groupsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/groups/v1"
)

func (c *Client) CreateGroup(ctx context.Context, req *groupsv1.CreateGroupRequest) (*groupsv1.Group, error) {
	return withConflictRetry(ctx, "create group", func() (*groupsv1.Group, error) {
		resp, err := c.groupsGateway.CreateGroup(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("create group: %w", err)
		}
		return resp.Group, nil
	})
}

func (c *Client) GetGroup(ctx context.Context, id string) (*groupsv1.Group, error) {
	resp, err := c.groupsGateway.GetGroup(ctx, &groupsv1.GetGroupRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}
	return resp.Group, nil
}

func (c *Client) UpdateGroup(ctx context.Context, req *groupsv1.UpdateGroupRequest) (*groupsv1.Group, error) {
	resp, err := c.groupsGateway.UpdateGroup(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("update group: %w", err)
	}
	return resp.Group, nil
}

func (c *Client) DeleteGroup(ctx context.Context, id string) error {
	return withConflictRetryNoResult(ctx, "delete group", func() error {
		_, err := c.groupsGateway.DeleteGroup(ctx, &groupsv1.DeleteGroupRequest{Id: id})
		if err != nil {
			return fmt.Errorf("delete group: %w", err)
		}
		return nil
	})
}

func (c *Client) AddGroupMember(ctx context.Context, req *groupsv1.AddMemberRequest) (*groupsv1.GroupMembership, error) {
	return withConflictRetry(ctx, "add group member", func() (*groupsv1.GroupMembership, error) {
		resp, err := c.groupsGateway.AddMember(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("add group member: %w", err)
		}
		return resp.Membership, nil
	})
}

func (c *Client) GetGroupMembershipByGroupAndMember(ctx context.Context, groupID string, memberType groupsv1.GroupMemberType, memberID string) (*groupsv1.GroupMembership, error) {
	resp, err := c.groupsGateway.ListMembers(ctx, &groupsv1.ListMembersRequest{GroupId: groupID, MemberType: memberType.Enum()})
	if err != nil {
		return nil, fmt.Errorf("list group members: %w", err)
	}
	for _, membership := range resp.GetMemberships() {
		if membership.GetMemberId() == memberID {
			return membership, nil
		}
	}
	return nil, ErrNotFound
}

func (c *Client) RemoveGroupMember(ctx context.Context, groupID string, memberID string) error {
	return withConflictRetryNoResult(ctx, "remove group member", func() error {
		_, err := c.groupsGateway.RemoveMember(ctx, &groupsv1.RemoveMemberRequest{GroupId: groupID, MemberId: memberID})
		if err != nil {
			return fmt.Errorf("remove group member: %w", err)
		}
		return nil
	})
}
