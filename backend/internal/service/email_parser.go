package service

import (
	"regexp"
	"strings"
	"time"

	"jobView-backend/internal/model"
)

var (
	linkRegex       = regexp.MustCompile(`https?://[^\s]+`)
	meetingIDRegex  = regexp.MustCompile(`(?i)(会议号|会议ID|Meeting\s*ID)[:：]?\s*([A-Za-z0-9-]{5,})`)
	dateTimeRegexes = []*regexp.Regexp{
		reregexpMust(`(?i)(\d{4}-\d{1,2}-\d{1,2})\s*[日号]?\s*(\d{1,2}:\d{2})`),
		reregexpMust(`(?i)(\d{1,2}/\d{1,2}/\d{4})\s*(\d{1,2}:\d{2})`),
		reregexpMust(`(?i)(\d{1,2}月\d{1,2}日)\s*(\d{1,2}:\d{2})`),
	}
)

type emailParseResult struct {
	Classification string
	Confidence     float64
	Payload        model.MailEventPayload
}

func reregexpMust(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}

// parseEmailContent 根据主题和正文识别类型、提取关键字段
func parseEmailContent(subject, body string) emailParseResult {
	subjectLower := strings.ToLower(subject)
	bodyLower := strings.ToLower(body)

	examScore := keywordScore(subjectLower, bodyLower, []string{
		"笔试", "考试", "在线测评", "测试", "assessment", "online test", "code test",
	})
	interviewScore := keywordScore(subjectLower, bodyLower, []string{
		"面试", "interview", "面谈", "interviewer", "interview schedule", "面试邀请",
	})

	classification := model.MailClassificationUnknown
	confidence := 0.0

	if examScore == 0 && interviewScore == 0 {
		classification = model.MailClassificationUnknown
	} else if examScore >= interviewScore {
		classification = model.MailClassificationExam
		confidence = scoreToConfidence(examScore)
	} else {
		classification = model.MailClassificationInterview
		confidence = scoreToConfidence(interviewScore)
	}

	links := linkRegex.FindAllString(body, -1)
	var examLink *string
	var meetingLink *string
	if len(links) > 0 {
		if classification == model.MailClassificationExam {
			examLink = &links[0]
		} else if classification == model.MailClassificationInterview {
			meetingLink = &links[0]
		}
	}

	var meetingID *string
	if match := meetingIDRegex.FindStringSubmatch(body); len(match) > 2 {
		value := strings.TrimSpace(match[2])
		meetingID = &value
	}

	var detectedTime *time.Time
	for _, re := range dateTimeRegexes {
		if match := re.FindStringSubmatch(body); len(match) >= 3 {
			if ts := combineDateTime(match[1], match[2]); ts != nil {
				detectedTime = ts
				break
			}
		}
	}

	payload := model.MailEventPayload{
		CompanyCandidates:  extractCandidates(subject),
		PositionCandidates: extractCandidates(body),
		DetectedTime:       detectedTime,
		ExamLink:           examLink,
		MeetingLink:        meetingLink,
		MeetingID:          meetingID,
		RawLinks:           links,
	}

	return emailParseResult{
		Classification: classification,
		Confidence:     confidence,
		Payload:        payload,
	}
}

func keywordScore(subject, body string, keywords []string) int {
	score := 0
	for _, keyword := range keywords {
		kw := strings.ToLower(keyword)
		if strings.Contains(subject, kw) {
			score += 3
		}
		count := strings.Count(body, kw)
		if count > 0 {
			score += 1 + count/2
		}
	}
	return score
}

func scoreToConfidence(score int) float64 {
	if score <= 0 {
		return 0
	}
	if score >= 8 {
		return 0.95
	}
	return float64(score) / 10.0
}

func combineDateTime(datePart, timePart string) *time.Time {
	datePart = strings.TrimSpace(datePart)
	timePart = strings.TrimSpace(timePart)
	formats := []string{
		"2006-1-2 15:04",
		"2006/1/2 15:04",
		"1/2/2006 15:04",
	}
	plDate := convertChineseDate(datePart)
	for _, dateStr := range []string{datePart, plDate} {
		if dateStr == "" {
			continue
		}
		for _, format := range formats {
			if ts, err := time.ParseInLocation(format, dateStr+" "+timePart, time.Local); err == nil {
				return &ts
			}
		}
	}
	return nil
}

func convertChineseDate(input string) string {
	input = strings.TrimSpace(input)
	if strings.Contains(input, "月") && strings.Contains(input, "日") {
		input = strings.ReplaceAll(input, "年", "-")
		input = strings.ReplaceAll(input, "月", "-")
		input = strings.ReplaceAll(input, "日", "")
		return input
	}
	return input
}

func extractCandidates(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	segments := strings.FieldsFunc(text, func(r rune) bool {
		switch r {
		case ',', '，', '/', '|', ';', '；', '。', '！', '!', '\n', '\r':
			return true
		}
		return false
	})
	var candidates []string
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if len([]rune(seg)) >= 2 && len(seg) <= 64 {
			candidates = append(candidates, seg)
		}
		if len(candidates) >= 5 {
			break
		}
	}
	return candidates
}
