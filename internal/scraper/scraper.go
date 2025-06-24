// scraper/scraper.go
package scraper

import (
	"fmt"
	"net/http"
	"opforjellyfin/internal/metadata"
	"opforjellyfin/internal/shared"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// FetchTorrents loads torrents from nyaa.si or configured tracker
func FetchTorrents() ([]shared.TorrentEntry, error) {
	//prep
	cfg := shared.LoadConfig()
	baseURL := strings.TrimRight(cfg.TorrentAPIURL, "/")
	var rawEntries []shared.TorrentEntry

	page := 1
	for {

		//nyaa specific
		searchURL := fmt.Sprintf("%s/?f=0&c=1_2&q=one+pace&p=%d", baseURL, page)
		resp, err := http.Get(searchURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil, err
		}

		rows := doc.Find("table tbody tr")
		if rows.Length() == 0 {
			break //donesies
		}

		// nyaa-structure.
		//TODO: add switch case to support other table structs
		rows.Each(func(i int, s *goquery.Selection) {
			title := s.Find("td:nth-child(2) a").Last().Text()
			seedersStr := s.Find("td:nth-child(6)").Text()
			torrentLink, _ := s.Find("td:nth-child(3) a[href^='/download/']").Attr("href")

			torrentID := extractIDFromLink(torrentLink)

			chapterRange := shared.ExtractChapterRangeFromTitle(title)
			rawIndex := extractRawIndex(chapterRange)

			seeders, _ := strconv.Atoi(strings.TrimSpace(seedersStr))
			quality := parseQuality(title)
			torrentName := extractTorrentName(title)

			if torrentLink != "" && strings.Contains(strings.ToLower(title), "one pace") {
				rawEntries = append(rawEntries, shared.TorrentEntry{
					Title:         title,
					Quality:       quality,
					TorrentName:   torrentName,
					Seeders:       seeders,
					RawIndex:      rawIndex,
					TorrentLink:   torrentLink,
					TorrentID:     torrentID,
					ChapterRange:  chapterRange,
					IsSpecial:     chapterRange == "",
					MetaDataAvail: metadata.HaveMetadata(chapterRange),
					HaveIt:        metadata.HaveVideoStatus(chapterRange),
				})
			}
		})

		page++
	}

	// sort by rawIndex ascending, then by seeders descending
	// maybe move this out to helpers
	sort.SliceStable(rawEntries, func(i, j int) bool {
		if rawEntries[i].RawIndex == rawEntries[j].RawIndex {
			return rawEntries[i].Seeders > rawEntries[j].Seeders
		}
		return rawEntries[i].RawIndex < rawEntries[j].RawIndex
	})

	rangeToKey := map[string]int{}
	key := 1
	specialKey := 9999

	// assign downloadKey after rawIndex
	for i := range rawEntries {
		cr := rawEntries[i].ChapterRange
		if cr == "" {
			rawEntries[i].DownloadKey = specialKey
			specialKey--
			continue // specials
		}
		if _, exists := rangeToKey[cr]; !exists {
			rangeToKey[cr] = key
			key++
		}
		rawEntries[i].DownloadKey = rangeToKey[cr]
	}

	return rawEntries, nil
}

// parseQuality returns video quality based on title string
func parseQuality(title string) string {
	title = strings.ToLower(title)

	switch {
	case strings.Contains(title, "1080p"):
		return "1080p"
	case strings.Contains(title, "720p"):
		return "720p"
	case strings.Contains(title, "480p"):
		return "480p"
	default:
		re := regexp.MustCompile(`\b\d{3,}p\b`) // try to find other quality yankily
		match := re.FindString(title)
		if match != "" {
			return match
		}
		return "n/a"
	}
}

// extracts raw index
func extractRawIndex(rangeStr string) int {
	if rangeStr == "" {
		return 9999 // specials
	}
	parts := strings.Split(rangeStr, "-")
	if len(parts) > 0 {
		if n, err := strconv.Atoi(parts[0]); err == nil {
			return n
		}
	}
	return 9999
}

// extracts torrent name for display
func extractTorrentName(title string) string {
	parts := regexp.MustCompile(`\[[^\]]+\]`).Split(title, -1)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			return part
		}
	}
	return "Unknown"
}

// extractIDFromLink parses numeric ID from a /download/xxxxxxx.torrent link
func extractIDFromLink(link string) int {
	re := regexp.MustCompile(`/download/(\d+)\.torrent`)
	matches := re.FindStringSubmatch(link)
	if len(matches) >= 2 {
		id, _ := strconv.Atoi(matches[1])
		return id
	}
	return 0
}
