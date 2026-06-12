package ratelimit

import (
	"time"

	authv1 "server-management-service/gen/go/auth/v1"
	reportingv1 "server-management-service/gen/go/reporting/v1"
	server_managementv1 "server-management-service/gen/go/server_management/v1"
)

var DefaultPolicy = Policy{
	Name:   "global",
	Limit:  300,
	Burst:  300,
	Window: time.Minute,
	Phase:  PhasePostAuth,
	Scopes: []KeyScope{
		ScopeUser,
	},
}

var MethodPolicies = map[string]Policy{
	authv1.AuthService_Login_FullMethodName: {
		Name:   "login",
		Limit:  5,
		Burst:  5,
		Window: time.Minute,
		Phase:  PhasePreAuth,
		Scopes: []KeyScope{
			ScopeIP,
			ScopeIdentifier,
		},
	},
	authv1.AuthService_RefreshToken_FullMethodName: {
		Name:   "refresh_token",
		Limit:  30,
		Burst:  10,
		Window: time.Minute,
		Phase:  PhasePreAuth,
		Scopes: []KeyScope{
			ScopeIP,
		},
	},
	server_managementv1.ServerManagementService_ViewServers_FullMethodName: {
		Name:   "server_read",
		Limit:  120,
		Burst:  200,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{ScopeUser},
	},
	server_managementv1.ServerManagementService_CreateServer_FullMethodName: {
		Name:   "server_write",
		Limit:  60,
		Burst:  80,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{ScopeUser},
	},
	server_managementv1.ServerManagementService_UpdateServer_FullMethodName: {
		Name:   "server_write",
		Limit:  60,
		Burst:  80,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{ScopeUser},
	},
	server_managementv1.ServerManagementService_DeleteServer_FullMethodName: {
		Name:   "server_dangerous",
		Limit:  20,
		Burst:  20,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{ScopeUser},
	},
	server_managementv1.ServerManagementService_ImportServers_FullMethodName: {
		Name:   "server_import",
		Limit:  5,
		Burst:  5,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{ScopeUser},
	},
	reportingv1.ReportingService_RequestReport_FullMethodName: {
		Name:   "report_request",
		Limit:  10,
		Burst:  10,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{ScopeUser},
	},
}

func PolicyForMethod(method string) Policy {
	if policy, ok := MethodPolicies[method]; ok {
		return normalizePolicy(policy)
	}

	return DefaultPolicy
}

func normalizePolicy(policy Policy) Policy {
	if policy.Burst <= 0 {
		policy.Burst = policy.Limit
	}

	if policy.Phase == "" {
		policy.Phase = PhasePostAuth
	}

	if len(policy.Scopes) == 0 {
		policy.Scopes = []KeyScope{ScopeUser}
	}

	return policy
}
