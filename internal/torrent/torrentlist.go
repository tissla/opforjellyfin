// torrent/torrentlist.go
package torrent

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

// FetchOnePaceTorrents loads torrents from nyaa.si or configured tracker
func FetchOnePaceTorrents() ([]shared.TorrentEntry, error) {
	cfg := shared.LoadConfig()
	baseURL := strings.TrimRight(cfg.TorrentAPIURL, "/")
	var rawEntries []shared.TorrentEntry

	page := 1
	for {
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

		rows.Each(func(i int, s *goquery.Selection) {
			title := s.Find("td:nth-child(2) a").Last().Text()
			seedersStr := s.Find("td:nth-child(6)").Text()
			torrentLink, _ := s.Find("td:nth-child(3) a[href^='/download/']").Attr("href")
			torrentID := extractIDFromLink(torrentLink)
			//chapterRange := extractChapterRange(title)
			chapterRange, _ := shared.ExtractChapterKeyFromTitle(title)

			seeders, _ := strconv.Atoi(strings.TrimSpace(seedersStr))
			quality := parseQuality(title)
			rawIndex, seasonName := parseSeasonMeta(title)

			if torrentLink != "" && strings.Contains(strings.ToLower(title), "one pace") {
				rawEntries = append(rawEntries, shared.TorrentEntry{
					Title:         title,
					Quality:       quality,
					SeasonName:    seasonName,
					Seeders:       seeders,
					RawIndex:      rawIndex,
					TorrentLink:   torrentLink,
					TorrentID:     torrentID,
					ChapterRange:  chapterRange,
					IsSpecial:     chapterRange == "",
					MetaDataAvail: metadata.HaveMetadata(chapterRange),
					HaveIt:        metadata.HaveVideoFile(chapterRange),
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

// parseSeasonMeta extracts raw index and season name
func parseSeasonMeta(title string) (int, string) {
	re := regexp.MustCompile(`(?i)\[(\d+)-\d+\]\s+(.+?)\s+\[`)
	matches := re.FindStringSubmatch(title)
	if len(matches) >= 3 {
		rawIndex, _ := strconv.Atoi(matches[1])
		seasonName := strings.TrimSpace(matches[2])
		return rawIndex, seasonName
	}

	// fallback: strip tags and use remaining
	clean := regexp.MustCompile(`\[.*?\]`).ReplaceAllString(title, "")
	clean = strings.TrimSpace(clean)
	if clean == "" {
		clean = "Unknown"
	}
	return 9999, clean
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

// unused old version
func extractChapterRange(title string) string {
	re := regexp.MustCompile(`\[(\d{1,4})-(\d{1,4})\]`)
	if match := re.FindStringSubmatch(title); len(match) == 3 {
		start, _ := strconv.Atoi(match[1])
		end, _ := strconv.Atoi(match[2])
		return fmt.Sprintf("%d-%d", start, end)
	}
	return ""
}
