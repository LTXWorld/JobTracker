package service

import (
    "encoding/json"
    "testing"
)

// Test helper functions used by completeness calculation
func TestCompletenessHelpers(t *testing.T) {
    // nonEmpty
    m := map[string]any{"name": "张三", "phone": "13800000000", "email": "a@b.com"}
    if !nonEmpty(m, "name", "phone", "email") {
        t.Fatalf("expected nonEmpty to return true for full fields")
    }
    if nonEmpty(m, "name", "missing") {
        t.Fatalf("expected nonEmpty to return false when a key is missing")
    }

    // hasArrayItems
    obj := map[string]any{
        "items": []any{map[string]any{"k": 1}},
    }
    raw, _ := json.Marshal(obj)
    if !hasArrayItems(raw, "items") {
        t.Fatalf("expected hasArrayItems to return true when array has elements")
    }
    emptyObj := map[string]any{"items": []any{}}
    rawEmpty, _ := json.Marshal(emptyObj)
    if hasArrayItems(rawEmpty, "items") {
        t.Fatalf("expected hasArrayItems to return false for empty array")
    }
}

