package api

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"modders-feedback/server/internal/domain"
	"modders-feedback/server/internal/httpx"
)

const rssLimit = 50

var (
	markdownImage = regexp.MustCompile(`!\[[^\[\]]*]\([^)\s]+(?:\s+"[^"]*")?\)`)
	htmlMedia     = regexp.MustCompile(`(?is)</?(?:img|video|audio|source|object|embed|iframe)\b[^>]*>`)
	htmlFileLink  = regexp.MustCompile(`(?is)<a\b[^>]*href=["'][^"']*(?:/api/attachments/|\.(?:png|jpe?g|gif|webp|pdf|mp4|webm|mov|mkv|avi|txt|log|md|csv))(?:[?#][^"']*)?["'][^>]*>.*?</a>`)
	markdownFile  = regexp.MustCompile(`(?i)\[[^\[\]]+]\([^)\s]*(?:/api/attachments/|\.(?:png|jpe?g|gif|webp|pdf|mp4|webm|mov|mkv|avi|txt|log|md|csv))(?:[?#][^)\s]*)?(?:\s+"[^"]*")?\)`)
	bareMediaURL  = regexp.MustCompile(`(?i)(?:https?://|/)[^\s<>()]*?(?:/api/attachments/|\.(?:png|jpe?g|gif|webp|pdf|mp4|webm|mov|mkv|avi))(?:[?#][^\s<>()]*)?`)
)

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language"`
	LastBuildDate string    `xml:"lastBuildDate,omitempty"`
	Items         []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
}

func (h *Handler) rss(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListRecentFeedback(rssLimit)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "读取反馈失败", "Could not load feedback"))
		return
	}
	base := publicBase(r)
	english := httpx.Locale(r) == "en"
	feed := rssFeed{
		Version: "2.0",
		Channel: rssChannel{
			Title:       rssText(english, "模组反馈", "Mod feedback"),
			Link:        base + "/",
			Description: rssText(english, "新提交的模组反馈", "Newly submitted mod feedback"),
			Language:    rssText(english, "zh-CN", "en"),
			Items:       make([]rssItem, 0, len(items)),
		},
	}
	if len(items) > 0 {
		feed.Channel.LastBuildDate = items[0].CreatedAt.UTC().Format(time.RFC1123Z)
	}
	for _, item := range items {
		if item.PublicID == "" || item.ModSlug == "" {
			continue
		}
		link := base + domain.FeedbackPath(item.ModSlug, item.Category, item.PublicID)
		feed.Channel.Items = append(feed.Channel.Items, rssItem{
			Title:       rssTitle(item, english),
			Link:        link,
			GUID:        fmt.Sprintf("feedback:%d", item.ID),
			PubDate:     item.CreatedAt.UTC().Format(time.RFC1123Z),
			Description: rssBody(item.Body),
		})
	}
	payload, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, httpx.Text(r, "生成订阅失败", "Could not build the feed"))
		return
	}
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(xml.Header))
	_, _ = w.Write(payload)
}

func rssTitle(item domain.Feedback, english bool) string {
	name := item.ModName
	if name == "" {
		name = item.ModSlug
	}
	return name + "-" + categoryLabel(item.Category, english) + "-" + item.Title
}

func categoryLabel(category string, english bool) string {
	switch category {
	case domain.CategoryBug:
		return rssText(english, "错误报告", "Bug report")
	case domain.CategoryFeature:
		return rssText(english, "功能建议", "Feature request")
	default:
		return rssText(english, "一般提问", "Question")
	}
}

func rssText(english bool, zh, en string) string {
	if english {
		return en
	}
	return zh
}

func rssBody(body string) string {
	text := markdownImage.ReplaceAllString(body, "")
	text = htmlMedia.ReplaceAllString(text, "")
	text = htmlFileLink.ReplaceAllString(text, "")
	text = markdownFile.ReplaceAllString(text, "")
	text = bareMediaURL.ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}

func publicBase(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := r.Host
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwarded != "" {
		host = strings.Split(forwarded, ",")[0]
	}
	return scheme + "://" + host
}
