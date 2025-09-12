package service

import (
    "context"
    "encoding/json"
    "strings"
    "time"
)

// recalcCompleteness 以一次查询获取所有分区并计算完善度
// 评分规则：base(40/20) + intent(30/15) + edu(+10) + exp(+15) + project/skill/cert/summary(每项+5)，上限100
func (s *ResumeService) recalcCompleteness(ctx context.Context, resumeID int) error {
    sections, err := s.repo.GetSectionsMap(ctx, resumeID)
    if err != nil { return err }

    comp := 0

    // base
    if raw, ok := sections["base"]; ok && len(raw) > 0 {
        var m map[string]any
        _ = json.Unmarshal(raw, &m)
        if nonEmpty(m, "name", "phone", "email") { comp += 40 } else { comp += 20 }
    }
    // intent
    if raw, ok := sections["intent"]; ok && len(raw) > 0 {
        var m map[string]any
        _ = json.Unmarshal(raw, &m)
        if nonEmpty(m, "position", "city") { comp += 30 } else { comp += 15 }
    }
    // edu
    if raw, ok := sections["edu"]; ok && len(raw) > 0 {
        if hasArrayItems(raw, "items") { comp += 10 }
    }
    // exp
    if raw, ok := sections["exp"]; ok && len(raw) > 0 {
        if hasArrayItems(raw, "items") { comp += 15 }
    }
    // others
    for _, t := range []string{"project", "skill", "cert", "summary"} {
        if raw, ok := sections[t]; ok && len(raw) > 0 { comp += 5 }
    }

    if comp > 100 { comp = 100 }
    return s.repo.UpdateResumeCompleteness(ctx, resumeID, comp, comp >= 80, time.Now())
}

// nonEmpty 判断 map 中的多个 key 是否均为非空字符串
func nonEmpty(m map[string]any, keys ...string) bool {
    for _, k := range keys {
        v, ok := m[k]
        if !ok { return false }
        s, _ := v.(string)
        if strings.TrimSpace(s) == "" { return false }
    }
    return true
}

// hasArrayItems 判断 JSON 对象中某字段为数组且长度>0
func hasArrayItems(raw json.RawMessage, field string) bool {
    var m map[string]any
    if err := json.Unmarshal(raw, &m); err != nil { return false }
    if v, ok := m[field]; ok {
        if arr, ok := v.([]any); ok { return len(arr) > 0 }
    }
    return false
}
