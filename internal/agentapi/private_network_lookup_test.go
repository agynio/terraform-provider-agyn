package agentapi

import (
	"context"
	"testing"

	groupsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/groups/v1"
	networksv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/networks/v1"
)

func TestGetGroupMembershipByGroupAndMemberPaginates(t *testing.T) {
	ctx := context.Background()
	calls := make([]*groupsv1.ListMembersRequest, 0, 2)
	client := &Client{groupsGateway: fakeGroupsGateway{listMembers: func(_ context.Context, req *groupsv1.ListMembersRequest) (*groupsv1.ListMembersResponse, error) {
		calls = append(calls, req)
		switch req.GetPageToken() {
		case "":
			return &groupsv1.ListMembersResponse{Memberships: []*groupsv1.GroupMembership{{GroupId: "group-id", MemberType: groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_USER, MemberId: "other-user"}}, NextPageToken: "page-2"}, nil
		case "page-2":
			return &groupsv1.ListMembersResponse{Memberships: []*groupsv1.GroupMembership{{GroupId: "group-id", MemberType: groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_USER, MemberId: "target-user"}}}, nil
		default:
			t.Fatalf("unexpected page token %q", req.GetPageToken())
			return nil, nil
		}
	}}}

	membership, err := client.GetGroupMembershipByGroupAndMember(ctx, "group-id", groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_USER, "target-user")
	if err != nil {
		t.Fatalf("get group membership: %v", err)
	}
	if membership.GetMemberId() != "target-user" {
		t.Fatalf("unexpected membership: %#v", membership)
	}
	assertGroupMembershipLookupCalls(t, calls)
}

func TestGetPrivateResourceAccessByResourceAndPrincipalPaginates(t *testing.T) {
	ctx := context.Background()
	principalType := networksv1.PrivateResourceAccessPrincipalType_PRIVATE_RESOURCE_ACCESS_PRINCIPAL_TYPE_GROUP
	calls := make([]*networksv1.ListPrivateResourceAccessRequest, 0, 2)
	client := &Client{networksGateway: fakeNetworksGateway{listPrivateResourceAccess: func(_ context.Context, req *networksv1.ListPrivateResourceAccessRequest) (*networksv1.ListPrivateResourceAccessResponse, error) {
		calls = append(calls, req)
		switch req.GetPageToken() {
		case "":
			return &networksv1.ListPrivateResourceAccessResponse{PrivateResourceAccess: []*networksv1.PrivateResourceAccess{{PrivateResourceId: "resource-id", PrincipalType: principalType, PrincipalId: "other-group"}}, NextPageToken: "page-2"}, nil
		case "page-2":
			return &networksv1.ListPrivateResourceAccessResponse{PrivateResourceAccess: []*networksv1.PrivateResourceAccess{{PrivateResourceId: "resource-id", PrincipalType: principalType, PrincipalId: "target-group"}}}, nil
		default:
			t.Fatalf("unexpected page token %q", req.GetPageToken())
			return nil, nil
		}
	}}}

	access, err := client.GetPrivateResourceAccessByResourceAndPrincipal(ctx, "resource-id", principalType, "target-group")
	if err != nil {
		t.Fatalf("get private resource access: %v", err)
	}
	if access.GetPrincipalId() != "target-group" {
		t.Fatalf("unexpected access grant: %#v", access)
	}
	assertPrivateResourceAccessLookupCalls(t, calls, principalType)
}

func assertGroupMembershipLookupCalls(t *testing.T, calls []*groupsv1.ListMembersRequest) {
	t.Helper()
	if len(calls) != 2 {
		t.Fatalf("expected 2 list calls, got %d", len(calls))
	}
	if calls[0].GetGroupId() != "group-id" || calls[0].GetMemberType() != groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_USER || calls[0].GetPageSize() != lookupPageSize || calls[0].GetPageToken() != "" {
		t.Fatalf("unexpected first call: %#v", calls[0])
	}
	if calls[1].GetGroupId() != "group-id" || calls[1].GetMemberType() != groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_USER || calls[1].GetPageSize() != lookupPageSize || calls[1].GetPageToken() != "page-2" {
		t.Fatalf("unexpected second call: %#v", calls[1])
	}
}

func assertPrivateResourceAccessLookupCalls(t *testing.T, calls []*networksv1.ListPrivateResourceAccessRequest, principalType networksv1.PrivateResourceAccessPrincipalType) {
	t.Helper()
	if len(calls) != 2 {
		t.Fatalf("expected 2 list calls, got %d", len(calls))
	}
	if calls[0].GetPrivateResourceId() != "resource-id" || calls[0].GetPrincipalType() != principalType || calls[0].GetPrincipalId() != "target-group" || calls[0].GetPageSize() != lookupPageSize || calls[0].GetPageToken() != "" {
		t.Fatalf("unexpected first call: %#v", calls[0])
	}
	if calls[1].GetPrivateResourceId() != "resource-id" || calls[1].GetPrincipalType() != principalType || calls[1].GetPrincipalId() != "target-group" || calls[1].GetPageSize() != lookupPageSize || calls[1].GetPageToken() != "page-2" {
		t.Fatalf("unexpected second call: %#v", calls[1])
	}
}

type fakeGroupsGateway struct {
	listMembers func(context.Context, *groupsv1.ListMembersRequest) (*groupsv1.ListMembersResponse, error)
}

func (f fakeGroupsGateway) CreateGroup(context.Context, *groupsv1.CreateGroupRequest) (*groupsv1.CreateGroupResponse, error) {
	panic("not implemented")
}
func (f fakeGroupsGateway) GetGroup(context.Context, *groupsv1.GetGroupRequest) (*groupsv1.GetGroupResponse, error) {
	panic("not implemented")
}
func (f fakeGroupsGateway) ListGroups(context.Context, *groupsv1.ListGroupsRequest) (*groupsv1.ListGroupsResponse, error) {
	panic("not implemented")
}
func (f fakeGroupsGateway) UpdateGroup(context.Context, *groupsv1.UpdateGroupRequest) (*groupsv1.UpdateGroupResponse, error) {
	panic("not implemented")
}
func (f fakeGroupsGateway) DeleteGroup(context.Context, *groupsv1.DeleteGroupRequest) (*groupsv1.DeleteGroupResponse, error) {
	panic("not implemented")
}
func (f fakeGroupsGateway) AddMember(context.Context, *groupsv1.AddMemberRequest) (*groupsv1.AddMemberResponse, error) {
	panic("not implemented")
}
func (f fakeGroupsGateway) RemoveMember(context.Context, *groupsv1.RemoveMemberRequest) (*groupsv1.RemoveMemberResponse, error) {
	panic("not implemented")
}
func (f fakeGroupsGateway) ListMembers(ctx context.Context, req *groupsv1.ListMembersRequest) (*groupsv1.ListMembersResponse, error) {
	return f.listMembers(ctx, req)
}
func (f fakeGroupsGateway) ListMemberGroups(context.Context, *groupsv1.ListMemberGroupsRequest) (*groupsv1.ListMemberGroupsResponse, error) {
	panic("not implemented")
}

type fakeNetworksGateway struct {
	listPrivateResourceAccess func(context.Context, *networksv1.ListPrivateResourceAccessRequest) (*networksv1.ListPrivateResourceAccessResponse, error)
}

func (f fakeNetworksGateway) CreateNetwork(context.Context, *networksv1.CreateNetworkRequest) (*networksv1.CreateNetworkResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) GetNetwork(context.Context, *networksv1.GetNetworkRequest) (*networksv1.GetNetworkResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) ListNetworks(context.Context, *networksv1.ListNetworksRequest) (*networksv1.ListNetworksResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) UpdateNetwork(context.Context, *networksv1.UpdateNetworkRequest) (*networksv1.UpdateNetworkResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) DeleteNetwork(context.Context, *networksv1.DeleteNetworkRequest) (*networksv1.DeleteNetworkResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) CreateTunnelCredential(context.Context, *networksv1.CreateTunnelCredentialRequest) (*networksv1.CreateTunnelCredentialResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) GetTunnelCredential(context.Context, *networksv1.GetTunnelCredentialRequest) (*networksv1.GetTunnelCredentialResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) ListTunnelCredentials(context.Context, *networksv1.ListTunnelCredentialsRequest) (*networksv1.ListTunnelCredentialsResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) DeleteTunnelCredential(context.Context, *networksv1.DeleteTunnelCredentialRequest) (*networksv1.DeleteTunnelCredentialResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) CreatePrivateResource(context.Context, *networksv1.CreatePrivateResourceRequest) (*networksv1.CreatePrivateResourceResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) GetPrivateResource(context.Context, *networksv1.GetPrivateResourceRequest) (*networksv1.GetPrivateResourceResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) ListPrivateResources(context.Context, *networksv1.ListPrivateResourcesRequest) (*networksv1.ListPrivateResourcesResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) UpdatePrivateResource(context.Context, *networksv1.UpdatePrivateResourceRequest) (*networksv1.UpdatePrivateResourceResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) DeletePrivateResource(context.Context, *networksv1.DeletePrivateResourceRequest) (*networksv1.DeletePrivateResourceResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) CreatePrivateResourceAccess(context.Context, *networksv1.CreatePrivateResourceAccessRequest) (*networksv1.CreatePrivateResourceAccessResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) DeletePrivateResourceAccess(context.Context, *networksv1.DeletePrivateResourceAccessRequest) (*networksv1.DeletePrivateResourceAccessResponse, error) {
	panic("not implemented")
}
func (f fakeNetworksGateway) ListPrivateResourceAccess(ctx context.Context, req *networksv1.ListPrivateResourceAccessRequest) (*networksv1.ListPrivateResourceAccessResponse, error) {
	return f.listPrivateResourceAccess(ctx, req)
}
