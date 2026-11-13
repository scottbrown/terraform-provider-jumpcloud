// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

type Policy struct {
	ID         string          `json:"id,omitempty"`
	Name       string          `json:"name"`
	TemplateID string          `json:"-"`
	State      string          `json:"state,omitempty"`
	Notes      string          `json:"notes,omitempty"`
	Values     json.RawMessage `json:"values,omitempty"`
}

func (p *Policy) UnmarshalJSON(data []byte) error {
	type Alias Policy
	aux := &struct {
		*Alias
		Template json.RawMessage `json:"template,omitempty"`
	}{Alias: (*Alias)(p)}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	if len(aux.Template) > 0 {
		// API returns either a string ID or an object with an "id" field
		var templateStr string
		if err := json.Unmarshal(aux.Template, &templateStr); err == nil {
			p.TemplateID = templateStr
		} else {
			var templateObj struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(aux.Template, &templateObj); err == nil {
				p.TemplateID = templateObj.ID
			}
		}
	}
	return nil
}

func (p Policy) MarshalJSON() ([]byte, error) {
	type Alias Policy
	aux := &struct {
		Alias
		Template string `json:"template,omitempty"`
	}{
		Alias:    (Alias)(p),
		Template: p.TemplateID,
	}
	return json.Marshal(aux)
}

type PolicyTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	OSFamily    string `json:"osFamily,omitempty"`
}

type PolicyAssociation struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (c *Client) CreatePolicy(ctx context.Context, policy *Policy) (*Policy, error) {
	var result Policy
	err := c.do(ctx, "POST", "/policies", policy, &result)
	if err != nil {
		return nil, fmt.Errorf("creating policy: %w", err)
	}
	return &result, nil
}

func (c *Client) GetPolicy(ctx context.Context, id string) (*Policy, error) {
	var result Policy
	err := c.do(ctx, "GET", "/policies/"+id, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting policy: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdatePolicy(ctx context.Context, id string, policy *Policy) (*Policy, error) {
	var result Policy
	err := c.do(ctx, "PUT", "/policies/"+id, policy, &result)
	if err != nil {
		return nil, fmt.Errorf("updating policy: %w", err)
	}
	return &result, nil
}

func (c *Client) DeletePolicy(ctx context.Context, id string) error {
	err := c.do(ctx, "DELETE", "/policies/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting policy: %w", err)
	}
	return nil
}

func (c *Client) ListPolicies(ctx context.Context, params *PaginationParams) ([]Policy, error) {
	path := c.buildPaginatedPath("/policies", params)
	var result []Policy
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing policies: %w", err)
	}
	return result, nil
}

func (c *Client) GetPolicyAssociations(ctx context.Context, policyID, targetType string) ([]PolicyAssociation, error) {
	path := fmt.Sprintf("/policies/%s/associations?targets=%s", policyID, targetType)
	var result []PolicyAssociation
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting policy associations: %w", err)
	}
	return result, nil
}

func (c *Client) AddPolicyAssociation(ctx context.Context, policyID string, assoc *PolicyAssociation) error {
	path := fmt.Sprintf("/policies/%s/associations", policyID)
	body := map[string]any{
		"op":   "add",
		"type": assoc.Type,
		"id":   assoc.ID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("adding policy association: %w", err)
	}
	return nil
}

func (c *Client) RemovePolicyAssociation(ctx context.Context, policyID string, assoc *PolicyAssociation) error {
	path := fmt.Sprintf("/policies/%s/associations", policyID)
	body := map[string]any{
		"op":   "remove",
		"type": assoc.Type,
		"id":   assoc.ID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("removing policy association: %w", err)
	}
	return nil
}

func (c *Client) GetPolicyTemplate(ctx context.Context, id string) (*PolicyTemplate, error) {
	var result PolicyTemplate
	err := c.do(ctx, "GET", "/policytemplates/"+id, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting policy template: %w", err)
	}
	return &result, nil
}

func (c *Client) ListPolicyTemplates(ctx context.Context, params *PaginationParams) ([]PolicyTemplate, error) {
	path := c.buildPaginatedPath("/policytemplates", params)
	var result []PolicyTemplate
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing policy templates: %w", err)
	}
	return result, nil
}
