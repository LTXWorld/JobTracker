package model

// 简历分区类型常量
const (
    SectionBase    = "base"
    SectionIntent  = "intent"
    SectionEdu     = "edu"
    SectionExp     = "exp"
    SectionProject = "project"
    SectionSkill   = "skill"
    SectionCert    = "cert"
    SectionHonor   = "honor"
    SectionSummary = "summary"
    SectionLinks   = "links"
)

// 完善度权重常量
const (
    CompletenessBaseFull    = 40
    CompletenessBasePartial = 20
    CompletenessIntentFull  = 30
    CompletenessIntentPart  = 15
    CompletenessEdu         = 10
    CompletenessExp         = 15
    CompletenessEachOther   = 5
    CompletenessDoneThresh  = 80
)

var validSectionTypes = map[string]struct{}{
    SectionBase:    {},
    SectionIntent:  {},
    SectionEdu:     {},
    SectionExp:     {},
    SectionProject: {},
    SectionSkill:   {},
    SectionCert:    {},
    SectionHonor:   {},
    SectionSummary: {},
    SectionLinks:   {},
}

// IsValidSectionType returns true if the section type is known
func IsValidSectionType(t string) bool {
    _, ok := validSectionTypes[t]
    return ok
}
