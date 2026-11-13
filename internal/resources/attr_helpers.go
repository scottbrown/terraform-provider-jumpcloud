// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// attrString returns the string value of the named attribute from a nested
// object's attribute map. The boolean is false when the attribute is absent,
// null, unknown, or not a string.
func attrString(attrs map[string]attr.Value, name string) (string, bool) {
	v, ok := attrs[name]
	if !ok || v.IsNull() || v.IsUnknown() {
		return "", false
	}
	s, ok := v.(types.String)
	if !ok {
		return "", false
	}
	return s.ValueString(), true
}

// attrBool returns the boolean value of the named attribute from a nested
// object's attribute map. The second boolean is false when the attribute is
// absent, null, unknown, or not a bool.
func attrBool(attrs map[string]attr.Value, name string) (bool, bool) {
	v, ok := attrs[name]
	if !ok || v.IsNull() || v.IsUnknown() {
		return false, false
	}
	b, ok := v.(types.Bool)
	if !ok {
		return false, false
	}
	return b.ValueBool(), true
}

// attrStringList returns the string elements of the named list attribute from
// a nested object's attribute map. The boolean is false when the attribute is
// absent, null, unknown, not a list, or cannot be converted to []string.
func attrStringList(ctx context.Context, attrs map[string]attr.Value, name string) ([]string, bool) {
	v, ok := attrs[name]
	if !ok || v.IsNull() || v.IsUnknown() {
		return nil, false
	}
	l, ok := v.(types.List)
	if !ok {
		return nil, false
	}
	var out []string
	if diags := l.ElementsAs(ctx, &out, false); diags.HasError() {
		return nil, false
	}
	return out, true
}
