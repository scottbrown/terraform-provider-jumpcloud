// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

type PasswordPolicy struct {
	ID                                     string `json:"id,omitempty"`
	Name                                   string `json:"name"`
	Description                            string `json:"description,omitempty"`
	Default                                bool   `json:"default,omitempty"`
	MinLength                              int    `json:"minLength,omitempty"`
	EnableMinLength                        bool   `json:"enableMinLength,omitempty"`
	NeedsUppercase                         bool   `json:"needsUppercase,omitempty"`
	NeedsLowercase                         bool   `json:"needsLowercase,omitempty"`
	NeedsNumeric                           bool   `json:"needsNumeric,omitempty"`
	NeedsSymbolic                          bool   `json:"needsSymbolic,omitempty"`
	EnableMaxLoginAttempts                 bool   `json:"enableMaxLoginAttempts,omitempty"`
	MaxLoginAttempts                       int    `json:"maxLoginAttempts,omitempty"`
	EnableLockoutTimeInSeconds             bool   `json:"enableLockoutTimeInSeconds,omitempty"`
	LockoutTimeInSeconds                   int    `json:"lockoutTimeInSeconds,omitempty"`
	EnablePasswordExpirationInDays         bool   `json:"enablePasswordExpirationInDays,omitempty"`
	PasswordExpirationInDays               int    `json:"passwordExpirationInDays,omitempty"`
	EnableDaysBeforeExpirationToForceReset bool   `json:"enableDaysBeforeExpirationToForceReset,omitempty"`
	DaysBeforeExpirationToForceReset       int    `json:"daysBeforeExpirationToForceReset,omitempty"`
	EnableDaysAfterExpirationToSelfRecover bool   `json:"enableDaysAfterExpirationToSelfRecover,omitempty"`
	DaysAfterExpirationToSelfRecover       int    `json:"daysAfterExpirationToSelfRecover,omitempty"`
	EnableMinChangePeriodInDays            bool   `json:"enableMinChangePeriodInDays,omitempty"`
	MinChangePeriodInDays                  int    `json:"minChangePeriodInDays,omitempty"`
	EnableMaxHistory                       bool   `json:"enableMaxHistory,omitempty"`
	MaxHistory                             int    `json:"maxHistory,omitempty"`
	EnableResetLockoutCounter              bool   `json:"enableResetLockoutCounter,omitempty"`
	ResetLockoutCounterMinutes             int    `json:"resetLockoutCounterMinutes,omitempty"`
	DisallowCommonlyUsedPasswords          bool   `json:"disallowCommonlyUsedPasswords,omitempty"`
	DisallowSequentialOrRepetitiveChars    bool   `json:"disallowSequentialOrRepetitiveChars,omitempty"`
	AllowUsernameSubstring                 bool   `json:"allowUsernameSubstring,omitempty"`
	AllowUnenrolledMFAPasswordReset        bool   `json:"allowUnenrolledMFAPasswordReset,omitempty"`
	EnableRecoveryEmail                    bool   `json:"enableRecoveryEmail,omitempty"`
	DisplayComplexityOnResetScreen         bool   `json:"displayComplexityOnResetScreen,omitempty"`
	Precedence                             int    `json:"precedence,omitempty"`
}

type passwordPolicyGetResponse struct {
	ObjectID string         `json:"objectId"`
	Policy   PasswordPolicy `json:"policy"`
}

type passwordPolicyListResponse struct {
	Results []struct {
		ObjectID    string `json:"objectId"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Precedence  int    `json:"precedence"`
		Default     bool   `json:"default"`
	} `json:"results"`
}

func (c *Client) CreatePasswordPolicy(ctx context.Context, policy *PasswordPolicy) (*PasswordPolicy, error) {
	var resp passwordPolicyGetResponse
	err := c.do(ctx, "POST", "/passwordpolicies", policy, &resp)
	if err != nil {
		return nil, fmt.Errorf("creating password policy: %w", err)
	}
	result := resp.Policy
	result.ID = resp.ObjectID
	return &result, nil
}

func (c *Client) GetPasswordPolicy(ctx context.Context, id string) (*PasswordPolicy, error) {
	var resp passwordPolicyGetResponse
	err := c.do(ctx, "GET", "/passwordpolicies/"+id, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("getting password policy: %w", err)
	}
	result := resp.Policy
	result.ID = resp.ObjectID
	return &result, nil
}

func (c *Client) UpdatePasswordPolicy(ctx context.Context, id string, policy *PasswordPolicy) (*PasswordPolicy, error) {
	var resp passwordPolicyGetResponse
	err := c.do(ctx, "PUT", "/passwordpolicies/"+id, policy, &resp)
	if err != nil {
		return nil, fmt.Errorf("updating password policy: %w", err)
	}
	result := resp.Policy
	result.ID = resp.ObjectID
	return &result, nil
}

func (c *Client) DeletePasswordPolicy(ctx context.Context, id string) error {
	err := c.do(ctx, "DELETE", "/passwordpolicies/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting password policy: %w", err)
	}
	return nil
}

func (c *Client) ListPasswordPolicies(ctx context.Context, params *PaginationParams) ([]PasswordPolicy, error) {
	path := c.buildPaginatedPath("/passwordpolicies", params)
	var resp passwordPolicyListResponse
	err := c.do(ctx, "GET", path, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("listing password policies: %w", err)
	}
	policies := make([]PasswordPolicy, len(resp.Results))
	for i, r := range resp.Results {
		policies[i] = PasswordPolicy{
			ID:         r.ObjectID,
			Name:       r.Name,
			Precedence: r.Precedence,
			Default:    r.Default,
		}
	}
	return policies, nil
}

func (c *Client) GetPasswordPolicyByName(ctx context.Context, name string) (*PasswordPolicy, error) {
	policies, err := c.ListPasswordPolicies(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, p := range policies {
		if p.Name == name {
			return c.GetPasswordPolicy(ctx, p.ID)
		}
	}
	return nil, fmt.Errorf("password policy with name %q not found", name)
}
