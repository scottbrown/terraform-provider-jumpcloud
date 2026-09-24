// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

type AuthenticationPolicy struct {
	ID         string                          `json:"id,omitempty"`
	Name       string                          `json:"name"`
	Type       string                          `json:"type"`
	Disabled   bool                            `json:"disabled,omitempty"`
	Conditions *AuthenticationPolicyConditions `json:"conditions,omitempty"`
	Effect     *AuthenticationPolicyEffect     `json:"effect,omitempty"`
	Targets    *AuthenticationPolicyTargets    `json:"targets,omitempty"`
}

type AuthenticationPolicyConditions struct {
	IPAddressIn    []string `json:"ipAddressIn,omitempty"`
	IPAddressNotIn []string `json:"ipAddressNotIn,omitempty"`
	DeviceTrusted  *bool    `json:"deviceTrusted,omitempty"`
	GeoIPIn        []string `json:"geoIpIn,omitempty"`
	GeoIPNotIn     []string `json:"geoIpNotIn,omitempty"`
}

func (c *AuthenticationPolicyConditions) UnmarshalJSON(data []byte) error {
	// The JumpCloud API returns conditions in a nested format:
	//   {"locationIn": {"countries": [...]}}
	//   {"not": {"ipAddressIn": [...]}}
	//   {"all": [{"ipAddressIn": [...]}, {"deviceManaged": true}]}
	//   {"any": [{"deviceManaged": false}, {"locationIn": {...}}]}
	// We normalize these into our flat struct.

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Try flat fields first (in case the API ever returns them directly)
	type plain AuthenticationPolicyConditions
	var flat plain
	_ = json.Unmarshal(data, &flat)
	*c = AuthenticationPolicyConditions(flat)

	// Parse nested structures
	if loc, ok := raw["locationIn"]; ok {
		c.parseLocationIn(loc, true)
	}
	if loc, ok := raw["locationNotIn"]; ok {
		c.parseLocationIn(loc, false)
	}

	if notRaw, ok := raw["not"]; ok {
		c.parseNot(notRaw)
	}

	if allRaw, ok := raw["all"]; ok {
		c.parseConditionList(allRaw, false)
	}

	if anyRaw, ok := raw["any"]; ok {
		c.parseConditionList(anyRaw, false)
	}

	return nil
}

func (c *AuthenticationPolicyConditions) parseLocationIn(data json.RawMessage, inclusive bool) {
	var loc struct {
		Countries []string `json:"countries"`
	}
	if json.Unmarshal(data, &loc) == nil && len(loc.Countries) > 0 {
		if inclusive {
			c.GeoIPIn = append(c.GeoIPIn, loc.Countries...)
		} else {
			c.GeoIPNotIn = append(c.GeoIPNotIn, loc.Countries...)
		}
	}
}

func (c *AuthenticationPolicyConditions) parseNot(data json.RawMessage) {
	var inner struct {
		IPAddressIn []string `json:"ipAddressIn"`
		LocationIn  *struct {
			Countries []string `json:"countries"`
		} `json:"locationIn"`
	}
	if json.Unmarshal(data, &inner) == nil {
		if len(inner.IPAddressIn) > 0 {
			c.IPAddressNotIn = append(c.IPAddressNotIn, inner.IPAddressIn...)
		}
		if inner.LocationIn != nil && len(inner.LocationIn.Countries) > 0 {
			c.GeoIPNotIn = append(c.GeoIPNotIn, inner.LocationIn.Countries...)
		}
	}
}

func (c *AuthenticationPolicyConditions) parseConditionList(data json.RawMessage, _ bool) {
	var items []json.RawMessage
	if json.Unmarshal(data, &items) != nil {
		return
	}

	for _, item := range items {
		var entry struct {
			IPAddressIn    []string `json:"ipAddressIn"`
			IPAddressNotIn []string `json:"ipAddressNotIn"`
			DeviceManaged  *bool    `json:"deviceManaged"`
			LocationIn     *struct {
				Countries []string `json:"countries"`
			} `json:"locationIn"`
		}
		if json.Unmarshal(item, &entry) != nil {
			continue
		}
		if len(entry.IPAddressIn) > 0 {
			c.IPAddressIn = append(c.IPAddressIn, entry.IPAddressIn...)
		}
		if len(entry.IPAddressNotIn) > 0 {
			c.IPAddressNotIn = append(c.IPAddressNotIn, entry.IPAddressNotIn...)
		}
		if entry.DeviceManaged != nil {
			c.DeviceTrusted = entry.DeviceManaged
		}
		if entry.LocationIn != nil && len(entry.LocationIn.Countries) > 0 {
			c.GeoIPIn = append(c.GeoIPIn, entry.LocationIn.Countries...)
		}
	}
}

type AuthenticationPolicyEffect struct {
	Action      string                                 `json:"action"`
	Obligations *AuthenticationPolicyEffectObligations `json:"obligations,omitempty"`
}

type AuthenticationPolicyEffectObligations struct {
	MFA *AuthenticationPolicyMFA `json:"mfa,omitempty"`
}

type AuthenticationPolicyMFA struct {
	Required       *bool    `json:"required,omitempty"`
	AllowedFactors []string `json:"allowedFactors,omitempty"`
}

type AuthenticationPolicyTargets struct {
	Resources []AuthenticationPolicyTargetResource `json:"resources,omitempty"`
}

type AuthenticationPolicyTargetResource struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func (c *Client) CreateAuthenticationPolicy(ctx context.Context, policy *AuthenticationPolicy) (*AuthenticationPolicy, error) {
	var result AuthenticationPolicy
	err := c.do(ctx, "POST", "/authn/policies", policy, &result)
	if err != nil {
		return nil, fmt.Errorf("creating authentication policy: %w", err)
	}
	return &result, nil
}

func (c *Client) GetAuthenticationPolicy(ctx context.Context, id string) (*AuthenticationPolicy, error) {
	var result AuthenticationPolicy
	err := c.do(ctx, "GET", "/authn/policies/"+id, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting authentication policy: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateAuthenticationPolicy(ctx context.Context, id string, policy *AuthenticationPolicy) (*AuthenticationPolicy, error) {
	var result AuthenticationPolicy
	err := c.doPatch(ctx, "/authn/policies/"+id, policy, &result)
	if err != nil {
		return nil, fmt.Errorf("updating authentication policy: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteAuthenticationPolicy(ctx context.Context, id string) error {
	err := c.do(ctx, "DELETE", "/authn/policies/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting authentication policy: %w", err)
	}
	return nil
}

func (c *Client) ListAuthenticationPolicies(ctx context.Context, params *PaginationParams) ([]AuthenticationPolicy, error) {
	path := c.buildPaginatedPath("/authn/policies", params)
	var result []AuthenticationPolicy
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing authentication policies: %w", err)
	}
	return result, nil
}
