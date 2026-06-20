package resources

import (
	"testing"

	groupsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/groups/v1"
	networksv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/networks/v1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPrivateNetworkEnumConversions(t *testing.T) {
	protocol, err := privateResourceProtocolFromString("HTTPS")
	if err != nil {
		t.Fatalf("unexpected protocol error: %v", err)
	}
	if protocol != networksv1.PrivateResourceProtocol_PRIVATE_RESOURCE_PROTOCOL_HTTPS {
		t.Fatalf("unexpected protocol: %v", protocol)
	}
	if got := privateResourceProtocolToString(networksv1.PrivateResourceProtocol_PRIVATE_RESOURCE_PROTOCOL_TCP); got != "tcp" {
		t.Fatalf("unexpected protocol string: %s", got)
	}

	principalType, err := privateResourceAccessPrincipalTypeFromString("group")
	if err != nil {
		t.Fatalf("unexpected principal type error: %v", err)
	}
	if principalType != networksv1.PrivateResourceAccessPrincipalType_PRIVATE_RESOURCE_ACCESS_PRINCIPAL_TYPE_GROUP {
		t.Fatalf("unexpected principal type: %v", principalType)
	}
	if got := privateResourceAccessPrincipalTypeToString(networksv1.PrivateResourceAccessPrincipalType_PRIVATE_RESOURCE_ACCESS_PRINCIPAL_TYPE_USER); got != "user" {
		t.Fatalf("unexpected principal type string: %s", got)
	}

	source, err := groupSourceFromString("")
	if err != nil {
		t.Fatalf("unexpected group source error: %v", err)
	}
	if source != groupsv1.GroupSource_GROUP_SOURCE_PLATFORM {
		t.Fatalf("unexpected group source: %v", source)
	}
	if got := groupSourceToString(groupsv1.GroupSource_GROUP_SOURCE_SCIM); got != "scim" {
		t.Fatalf("unexpected group source string: %s", got)
	}

	memberType, err := groupMemberTypeFromString("agent")
	if err != nil {
		t.Fatalf("unexpected member type error: %v", err)
	}
	if memberType != groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_AGENT {
		t.Fatalf("unexpected member type: %v", memberType)
	}
	if got := groupMemberTypeToString(groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_APP); got != "app" {
		t.Fatalf("unexpected member type string: %s", got)
	}
}

func TestInt32ListFromPlanRejectsUnknownElement(t *testing.T) {
	_, err := int32ListFromPlan(types.ListValueMust(types.Int32Type, []attr.Value{types.Int32Unknown()}))
	if err == nil {
		t.Fatalf("expected error for unknown port element")
	}
}

func TestPrivateResourceState(t *testing.T) {
	privateResource := &networksv1.PrivateResource{
		Meta:              &networksv1.EntityMeta{Id: "resource-id"},
		OrganizationId:    "org-id",
		NetworkId:         "network-id",
		Name:              "database",
		Protocol:          networksv1.PrivateResourceProtocol_PRIVATE_RESOURCE_PROTOCOL_TCP,
		TargetHost:        "db.internal",
		TargetPorts:       []int32{5432},
		InterceptHost:     "db.agyn.internal",
		InterceptPorts:    []int32{5432},
		ProvisioningState: networksv1.ProvisioningState_PROVISIONING_STATE_ACTIVE,
	}
	state := privateResourceState(privateResource)
	if state.ID.ValueString() != "resource-id" || state.Protocol.ValueString() != "tcp" || state.ProvisioningState.ValueString() != "active" {
		t.Fatalf("unexpected state: %#v", state)
	}
}
