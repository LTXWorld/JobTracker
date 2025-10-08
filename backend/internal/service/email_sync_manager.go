package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"strconv"
	"strings"
	"time"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/robfig/cron/v3"

	"jobView-backend/internal/config"
	"jobView-backend/internal/model"
	"jobView-backend/internal/repository"
	"jobView-backend/internal/utils"
)

// EmailSyncManager 定时拉取用户邮箱并解析关键事件
type EmailSyncManager struct {
	mailCfg       config.MailConfig
	mailRepo      repository.EmailIntegrationRepository
	eventRepo     repository.MailEventRepository
	encryptionKey string
	cron          *cron.Cron
	logger        *log.Logger
	processor     *EmailEventProcessor
}

func NewEmailSyncManager(mailCfg config.MailConfig, mailRepo repository.EmailIntegrationRepository, eventRepo repository.MailEventRepository, processor *EmailEventProcessor, encryptionKey string, logger *log.Logger) *EmailSyncManager {
	if logger == nil {
		logger = log.Default()
	}
	return &EmailSyncManager{
		mailCfg:       mailCfg,
		mailRepo:      mailRepo,
		eventRepo:     eventRepo,
		processor:     processor,
		encryptionKey: encryptionKey,
		logger:        logger,
	}
}

// Start 启动定时任务
func (m *EmailSyncManager) Start(ctx context.Context) error {
	if !m.mailCfg.PollingEnabled {
		m.logger.Println("[EmailSync] Polling disabled via configuration")
		return nil
	}
	if m.cron != nil {
		return fmt.Errorf("email sync manager already started")
	}
	cronScheduler := cron.New(cron.WithSeconds())
	if _, err := cronScheduler.AddFunc(toCronSpec(m.mailCfg.WorkHoursCron), func() {
		m.safeSync(ctx, "work_hours")
	}); err != nil {
		return fmt.Errorf("failed to schedule work hours sync: %w", err)
	}
	if _, err := cronScheduler.AddFunc(toCronSpec(m.mailCfg.OffHoursCron), func() {
		m.safeSync(ctx, "off_hours")
	}); err != nil {
		return fmt.Errorf("failed to schedule off hours sync: %w", err)
	}
	cronScheduler.Start()
	m.cron = cronScheduler
	m.logger.Println("[EmailSync] scheduler started")
	return nil
}

// Stop 停止定时任务
func (m *EmailSyncManager) Stop() {
	if m.cron != nil {
		m.cron.Stop()
		m.cron = nil
		m.logger.Println("[EmailSync] scheduler stopped")
	}
}

func (m *EmailSyncManager) safeSync(ctx context.Context, reason string) {
	defer func() {
		if r := recover(); r != nil {
			m.logger.Printf("[EmailSync] panic recovered (%s): %v", reason, r)
		}
	}()
	if err := m.SyncOnce(ctx); err != nil {
		m.logger.Printf("[EmailSync] sync failed (%s): %v", reason, err)
	}
}

// SyncOnce 立即执行一次同步，便于测试和手动触发
func (m *EmailSyncManager) SyncOnce(ctx context.Context) error {
	mailboxes, err := m.mailRepo.ListActiveMailboxes()
	if err != nil {
		return fmt.Errorf("list mailboxes failed: %w", err)
	}
	if len(mailboxes) == 0 {
		return nil
	}
	for _, mailbox := range mailboxes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := m.syncMailbox(ctx, mailbox); err != nil {
			m.logger.Printf("[EmailSync] mailbox sync failed for user=%d email=%s: %v", mailbox.UserID, mailbox.EmailAddress, err)
			errMsg := err.Error()
			_ = m.mailRepo.UpdateSyncState(mailbox.UserID, mailbox.LastMessageUID, time.Now(), "error", &errMsg)
		}
	}
	return nil
}

func (m *EmailSyncManager) syncMailbox(ctx context.Context, mailbox *model.UserMailbox) error {
	if mailbox == nil {
		return nil
	}
	secret, err := utils.DecryptString(m.encryptionKey, mailbox.EncryptedSecret)
	if err != nil {
		return fmt.Errorf("decrypt authorization failed: %w", err)
	}

	var c *client.Client
	if mailbox.UseSSL {
		c, err = client.DialTLS(fmt.Sprintf("%s:%d", mailbox.Host, mailbox.Port), nil)
	} else {
		c, err = client.Dial(fmt.Sprintf("%s:%d", mailbox.Host, mailbox.Port))
	}
	if err != nil {
		return fmt.Errorf("connect to mailbox failed: %w", err)
	}
	defer c.Logout()

	if err := c.Login(mailbox.EmailAddress, secret); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	status, err := c.Select("INBOX", false)
	if err != nil {
		return fmt.Errorf("select inbox failed: %w", err)
	}

	startUID := determineStartUID(mailbox.LastMessageUID, status)
	if startUID == 0 {
		return nil
	}
	seqSet := new(imap.SeqSet)
	seqSet.AddRange(startUID, ^uint32(0))

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, imap.FetchInternalDate, section.FetchItem()}
	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.UidFetch(seqSet, items, messages)
	}()

	var fetchedMaxUID uint32
	for msg := range messages {
		if msg == nil {
			continue
		}
		if msg.Uid > fetchedMaxUID {
			fetchedMaxUID = msg.Uid
		}
		if err := m.processMessage(mailbox, msg, section); err != nil {
			m.logger.Printf("[EmailSync] process message failed user=%d email=%s uid=%d: %v", mailbox.UserID, mailbox.EmailAddress, msg.Uid, err)
		}
	}

	if err := <-done; err != nil {
		return fmt.Errorf("fetch messages failed: %w", err)
	}

	var lastUIDStr *string
	if fetchedMaxUID > 0 {
		uidStr := strconv.FormatUint(uint64(fetchedMaxUID), 10)
		lastUIDStr = &uidStr
	} else {
		lastUIDStr = mailbox.LastMessageUID
	}
	if err := m.mailRepo.UpdateSyncState(mailbox.UserID, lastUIDStr, time.Now(), "active", nil); err != nil {
		m.logger.Printf("[EmailSync] update sync state failed user=%d: %v", mailbox.UserID, err)
	}
	return nil
}

func (m *EmailSyncManager) processMessage(mailbox *model.UserMailbox, msg *imap.Message, section *imap.BodySectionName) error {
	body := msg.GetBody(section)
	if body == nil {
		return fmt.Errorf("message body is empty")
	}
	reader, err := mail.CreateReader(body)
	if err != nil {
		return fmt.Errorf("create mail reader failed: %w", err)
	}

	var plainBuilder strings.Builder
	for {
		part, err := reader.NextPart()
		if err != nil {
			break
		}
		switch h := part.Header.(type) {
		case *mail.InlineHeader:
			mediaType, _, _ := h.ContentType()
			if mediaType == "text/plain" {
				buf := new(bytes.Buffer)
				if _, err := buf.ReadFrom(part.Body); err == nil {
					plainBuilder.WriteString(buf.String())
				}
			} else if mediaType == "text/html" && plainBuilder.Len() == 0 {
				buf := new(bytes.Buffer)
				if _, err := buf.ReadFrom(part.Body); err == nil {
					txt := stripHTML(buf.String())
					plainBuilder.WriteString(txt)
				}
			}
		}
	}

	bodyText := strings.TrimSpace(plainBuilder.String())
	if bodyText == "" {
		bodyText = "(正文为空或为附件内容)"
	}

	subject := reader.Header.Get("Subject")
	subjectDecoded, _ := decodeMIMEHeader(subject)
	sender := reader.Header.Get("From")
	senderDecoded, _ := decodeMIMEHeader(sender)

	result := parseEmailContent(subjectDecoded, bodyText)
	payloadBytes, _ := json.Marshal(result.Payload)

	snippet := snippet(bodyText, 180)

	mailEvent := &model.MailEvent{
		UserID:         mailbox.UserID,
		MailboxID:      mailbox.ID,
		MessageID:      optionalString(reader.Header.Get("Message-Id")),
		MessageUID:     optionalString(strconv.FormatUint(uint64(msg.Uid), 10)),
		Subject:        subjectDecoded,
		Sender:         senderDecoded,
		ReceivedAt:     msg.InternalDate,
		Snippet:        &snippet,
		Classification: result.Classification,
		Confidence:     result.Confidence,
		Payload:        payloadBytes,
		Status:         deriveInitialStatus(result),
	}

	created, err := m.eventRepo.Create(mailEvent)
	if err != nil {
		return err
	}

	if m.processor != nil {
		if err := m.processor.Process(created); err != nil {
			m.logger.Printf("[EmailSync] process automation failed user=%d event=%d: %v", mailbox.UserID, created.ID, err)
		}
	}
	return nil
}

func deriveInitialStatus(result emailParseResult) string {
	if result.Classification == model.MailClassificationUnknown || result.Confidence < 0.4 {
		return model.MailEventStatusNeedsReview
	}
	return model.MailEventStatusPending
}

func determineStartUID(lastUID *string, status *imap.MailboxStatus) uint32 {
	if status == nil {
		return 1
	}
	if lastUID != nil && *lastUID != "" {
		if uid64, err := strconv.ParseUint(*lastUID, 10, 32); err == nil {
			return uint32(uid64) + 1
		}
	}
	if status.UidNext > 0 {
		if status.UidNext > 20 {
			return status.UidNext - 20
		}
		return 1
	}
	return 1
}

func optionalString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func snippet(text string, length int) string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= length {
		return string(runes)
	}
	return string(runes[:length]) + "…"
}

func stripHTML(html string) string {
	replacer := strings.NewReplacer("<br>", "\n", "<br/>", "\n", "<br />", "\n")
	plain := replacer.Replace(html)
	plain = strings.ReplaceAll(plain, "</p>", "\n")
	plain = regexpStripTags(plain)
	return plain
}

func regexpStripTags(input string) string {
	var b strings.Builder
	inTag := false
	for _, r := range input {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		case '\n':
			if !inTag {
				b.WriteRune('\n')
			}
		default:
			if !inTag {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

var wordDecoder = &mime.WordDecoder{}

func decodeMIMEHeader(input string) (string, error) {
	if input == "" {
		return "", nil
	}
	decoded, err := wordDecoder.DecodeHeader(input)
	if err != nil {
		return input, err
	}
	return decoded, nil
}

func toCronSpec(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return "0 0 * * * *" // 默认每小时一次
	}
	parts := strings.Fields(input)
	if len(parts) == 5 {
		return "0 " + input
	}
	return input
}
