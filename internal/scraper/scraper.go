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

// TODO: sort file, add more structs, add scrape-map

// gets the torrents using current config, throws error if no valid config found
func FetchTorrents(cfg shared.Config) ([]shared.TorrentEntry, error) {
	//prep

	// use scraper config from main config
	srcConfig := cfg.Source

	// ensure we have a valid scraper config
	if srcConfig.Name == "" || srcConfig.BaseURL == "" {
		return nil, fmt.Errorf("no scraper configuration found. Please run 'opfor setDir <path>' first")
	}

	baseURL := srcConfig.BaseURL

	var rawEntries []shared.TorrentEntry

	page := 1
	for {
		searchURL := fmt.Sprintf(baseURL+srcConfig.SearchPathTemplate, srcConfig.SearchQuery, page)

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

		rows := doc.Find(srcConfig.RowSelector)
		if rows.Length() == 0 {
			break // finito
		}

		rows.Each(func(i int, s *goquery.Selection) {
			entry, ok := parseRow(s, &srcConfig, baseURL)
			if ok {
				rawEntries = append(rawEntries, entry)
			}
		})

		page++
	}

	// Sort and assign download keys
	return processEntries(rawEntries), nil
}

// parseRow extracts torrent data from a table row using the scraper config
func parseRow(s *goquery.Selection, config *shared.ScraperConfig, baseURL string) (shared.TorrentEntry, bool) {
	// Extract fields using configured selectors
	title := s.Find(config.Fields.Title).Text()
	seedersStr := s.Find(config.Fields.Seeders).Text()
	torrentLink, _ := s.Find(config.Fields.TorrentLink).Attr("href")
	date := s.Find(config.Fields.UploadDate).Text()

	// Validate based on config
	if config.Validation.RequiredInTitle != "" {
		if !strings.Contains(strings.ToLower(title), strings.ToLower(config.Validation.RequiredInTitle)) {
			return shared.TorrentEntry{}, false
		}
	}

	if torrentLink == "" {
		return shared.TorrentEntry{}, false
	}

	// Extract torrent ID using regex from config
	torrentID := 0
	if config.Fields.TorrentID != "" {
		re := regexp.MustCompile(config.Fields.TorrentID)
		matches := re.FindStringSubmatch(torrentLink)
		if len(matches) >= 2 {
			torrentID, _ = strconv.Atoi(matches[1])
		}
	}

	// Parse the rest of the data
	chapterRange := shared.ExtractChapterRangeFromTitle(title)
	rawIndex := extractRawIndex(chapterRange)
	seeders, _ := strconv.Atoi(strings.TrimSpace(seedersStr))
	quality := parseQuality(title)
	torrentName := extractTorrentName(title)

	// Make torrent link absolute if needed
	if !strings.HasPrefix(torrentLink, "http") {
		torrentLink = baseURL + torrentLink
	}

	metaDataAvail := metadata.HaveMetadata(chapterRange)

	videoStatus := metadata.HaveVideoStatus(chapterRange)

	isExtended := isExtended(title)

	return shared.TorrentEntry{
		Title:         title,
		Quality:       quality,
		TorrentName:   torrentName,
		Seeders:       seeders,
		RawIndex:      rawIndex,
		TorrentLink:   torrentLink,
		TorrentID:     torrentID,
		ChapterRange:  chapterRange,
		IsSpecial:     chapterRange == "",
		MetaDataAvail: metaDataAvail,
		HaveIt:        videoStatus,
		Date:          date,
		IsExtended:    isExtended,
	}, true
}

// processEntries sorts entries, filters out dead torrents and assigns unique download keys.
func processEntries(rawEntries []shared.TorrentEntry) []shared.TorrentEntry {
	// filter out torrents with 0 seeders
	filtered := make([]shared.TorrentEntry, 0, len(rawEntries))
	for _, entry := range rawEntries {
		if entry.Seeders > 0 {
			filtered = append(filtered, entry)
		}
	}

	// sort by rawIndex ascending, then seeders descending
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].RawIndex == filtered[j].RawIndex {
			return filtered[i].Seeders > filtered[j].Seeders
		}
		return filtered[i].RawIndex < filtered[j].RawIndex
	})

	// assign unique download keys
	key := 1
	specialKey := 9999

	for i := range filtered {
		if filtered[i].IsSpecial || filtered[i].ChapterRange == "" {
			filtered[i].DownloadKey = specialKey
			specialKey--
		} else {
			filtered[i].DownloadKey = key
			key++
		}
	}

	return filtered
}

func isExtended(title string) bool {
	title = strings.ToLower(title)

	return strings.Contains(title, "extended")
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
