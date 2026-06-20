package resources

import (
	"fmt"
	"strings"

	groupsv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/groups/v1"
	networksv1 "github.com/agynio/terraform-provider-agyn/gen/agynio/api/networks/v1"
)

func provisioningStateToString(state networksv1.ProvisioningState) string {
	switch state {
	case networksv1.ProvisioningState_PROVISIONING_STATE_ACTIVE:
		return "active"
	case networksv1.ProvisioningState_PROVISIONING_STATE_FAILED:
		return "failed"
	case networksv1.ProvisioningState_PROVISIONING_STATE_REMOVING:
		return "removing"
	case networksv1.ProvisioningState_PROVISIONING_STATE_UNSPECIFIED:
		return ""
	default:
		panic("unsupported provisioning state " + state.String())
	}
}

func tunnelEnrollmentStateToString(state networksv1.TunnelEnrollmentState) string {
	switch state {
	case networksv1.TunnelEnrollmentState_TUNNEL_ENROLLMENT_STATE_PENDING:
		return "pending"
	case networksv1.TunnelEnrollmentState_TUNNEL_ENROLLMENT_STATE_ENROLLED:
		return "enrolled"
	case networksv1.TunnelEnrollmentState_TUNNEL_ENROLLMENT_STATE_UNSPECIFIED:
		return ""
	default:
		panic("unsupported tunnel enrollment state " + state.String())
	}
}

func tunnelConnectivityToString(connectivity networksv1.TunnelConnectivity) string {
	switch connectivity {
	case networksv1.TunnelConnectivity_TUNNEL_CONNECTIVITY_ONLINE:
		return "online"
	case networksv1.TunnelConnectivity_TUNNEL_CONNECTIVITY_OFFLINE:
		return "offline"
	case networksv1.TunnelConnectivity_TUNNEL_CONNECTIVITY_UNSPECIFIED:
		return ""
	default:
		panic("unsupported tunnel connectivity " + connectivity.String())
	}
}

func privateResourceProtocolFromString(value string) (networksv1.PrivateResourceProtocol, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "tcp":
		return networksv1.PrivateResourceProtocol_PRIVATE_RESOURCE_PROTOCOL_TCP, nil
	case "http":
		return networksv1.PrivateResourceProtocol_PRIVATE_RESOURCE_PROTOCOL_HTTP, nil
	case "https":
		return networksv1.PrivateResourceProtocol_PRIVATE_RESOURCE_PROTOCOL_HTTPS, nil
	default:
		return networksv1.PrivateResourceProtocol_PRIVATE_RESOURCE_PROTOCOL_UNSPECIFIED, fmt.Errorf("protocol must be tcp, http, or https")
	}
}

func privateResourceProtocolToString(protocol networksv1.PrivateResourceProtocol) string {
	switch protocol {
	case networksv1.PrivateResourceProtocol_PRIVATE_RESOURCE_PROTOCOL_TCP:
		return "tcp"
	case networksv1.PrivateResourceProtocol_PRIVATE_RESOURCE_PROTOCOL_HTTP:
		return "http"
	case networksv1.PrivateResourceProtocol_PRIVATE_RESOURCE_PROTOCOL_HTTPS:
		return "https"
	case networksv1.PrivateResourceProtocol_PRIVATE_RESOURCE_PROTOCOL_UNSPECIFIED:
		return ""
	default:
		panic("unsupported private resource protocol " + protocol.String())
	}
}

func privateResourceAccessPrincipalTypeFromString(value string) (networksv1.PrivateResourceAccessPrincipalType, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "agent":
		return networksv1.PrivateResourceAccessPrincipalType_PRIVATE_RESOURCE_ACCESS_PRINCIPAL_TYPE_AGENT, nil
	case "user":
		return networksv1.PrivateResourceAccessPrincipalType_PRIVATE_RESOURCE_ACCESS_PRINCIPAL_TYPE_USER, nil
	case "app":
		return networksv1.PrivateResourceAccessPrincipalType_PRIVATE_RESOURCE_ACCESS_PRINCIPAL_TYPE_APP, nil
	case "group":
		return networksv1.PrivateResourceAccessPrincipalType_PRIVATE_RESOURCE_ACCESS_PRINCIPAL_TYPE_GROUP, nil
	default:
		return networksv1.PrivateResourceAccessPrincipalType_PRIVATE_RESOURCE_ACCESS_PRINCIPAL_TYPE_UNSPECIFIED, fmt.Errorf("principal_type must be agent, user, app, or group")
	}
}

func privateResourceAccessPrincipalTypeToString(principalType networksv1.PrivateResourceAccessPrincipalType) string {
	switch principalType {
	case networksv1.PrivateResourceAccessPrincipalType_PRIVATE_RESOURCE_ACCESS_PRINCIPAL_TYPE_AGENT:
		return "agent"
	case networksv1.PrivateResourceAccessPrincipalType_PRIVATE_RESOURCE_ACCESS_PRINCIPAL_TYPE_USER:
		return "user"
	case networksv1.PrivateResourceAccessPrincipalType_PRIVATE_RESOURCE_ACCESS_PRINCIPAL_TYPE_APP:
		return "app"
	case networksv1.PrivateResourceAccessPrincipalType_PRIVATE_RESOURCE_ACCESS_PRINCIPAL_TYPE_GROUP:
		return "group"
	case networksv1.PrivateResourceAccessPrincipalType_PRIVATE_RESOURCE_ACCESS_PRINCIPAL_TYPE_UNSPECIFIED:
		return ""
	default:
		panic("unsupported private resource access principal type " + principalType.String())
	}
}

func groupSourceFromString(value string) (groupsv1.GroupSource, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "platform":
		return groupsv1.GroupSource_GROUP_SOURCE_PLATFORM, nil
	case "scim":
		return groupsv1.GroupSource_GROUP_SOURCE_SCIM, nil
	default:
		return groupsv1.GroupSource_GROUP_SOURCE_UNSPECIFIED, fmt.Errorf("source must be platform or scim")
	}
}

func groupSourceToString(source groupsv1.GroupSource) string {
	switch source {
	case groupsv1.GroupSource_GROUP_SOURCE_PLATFORM:
		return "platform"
	case groupsv1.GroupSource_GROUP_SOURCE_SCIM:
		return "scim"
	case groupsv1.GroupSource_GROUP_SOURCE_UNSPECIFIED:
		return ""
	default:
		panic("unsupported group source " + source.String())
	}
}

func groupMemberTypeFromString(value string) (groupsv1.GroupMemberType, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "user":
		return groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_USER, nil
	case "agent":
		return groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_AGENT, nil
	case "app":
		return groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_APP, nil
	default:
		return groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_UNSPECIFIED, fmt.Errorf("member_type must be user, agent, or app")
	}
}

func groupMemberTypeToString(memberType groupsv1.GroupMemberType) string {
	switch memberType {
	case groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_USER:
		return "user"
	case groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_AGENT:
		return "agent"
	case groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_APP:
		return "app"
	case groupsv1.GroupMemberType_GROUP_MEMBER_TYPE_UNSPECIFIED:
		return ""
	default:
		panic("unsupported group member type " + memberType.String())
	}
}
