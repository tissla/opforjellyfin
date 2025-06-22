package shared

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// strict version
func ExtractChapterKeyFromTitle(title string) (string, error) {
	re := regexp.MustCompile(`(?i)\[One Pace\]\[([^\]]+)\]`)
	matches := re.FindStringSubmatch(title)
	if len(matches) < 2 {
		return "", fmt.Errorf("could not extract chapter info from title: %s", title)
	}

	chapterInfo := matches[1]

	parts := strings.Split(chapterInfo, ",")
	first := strings.TrimSpace(parts[0])

	if matched, _ := regexp.MatchString(`^\d+$`, first); matched {
		return fmt.Sprintf("%s-%s", first, first), nil
	}

	if matched, _ := regexp.MatchString(`^\d+-\d+$`, first); matched {
		return first, nil
	}

	return "", fmt.Errorf("could not parse chapter format: %s", first)
}

// extracts the two ints separated by "-"
func ParseRange(r string) (int, int) {
	parts := strings.Split(r, "-")
	if len(parts) != 2 {
		return -1, -1
	}
	a, _ := strconv.Atoi(parts[0])
	b, _ := strconv.Atoi(parts[1])
	return a, b
}

// extract string between two xmls tags. e.g "<season>3</season>"" -> "3"
func ExtractXMLTag(data []byte, tag string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?i)<%s>(.*?)</%s>`, tag, tag))
	matches := re.FindSubmatch(data)
	if len(matches) >= 2 {
		return strings.TrimSpace(string(matches[1]))
	}
	return ""
}

// gets chapter range from .nfo file. e.g "Manga Chapter(s): 8-11" -> "8-11" or "Manga Chapter(s): 1" -> "1"
func ExtractChapterRangeFromNFO(content string) string {
	re := regexp.MustCompile(`(?i)Manga\s*Chapter\(s\)?:\s*(\d+)(?:[\s,-]*(\d+))?`)
	match := re.FindStringSubmatch(content)
	if len(match) >= 2 {
		start := match[1]
		end := match[2]
		if end == "" {
			end = start
		}
		return fmt.Sprintf("%s-%s", start, end)
	}
	return ""
}

// used to get season from folder-name. "Season 02" -> "02"
func ExtractSeasonNumber(seasonKey string) string {
	parts := strings.Fields(seasonKey)
	if len(parts) == 2 {
		return parts[1]
	}
	return "00"
}

// used to get episode from episodekey in title. "S02E03" -> "03"
func ExtractEpisodeNumberFromKey(episodeKey string) string {
	re := regexp.MustCompile(`E(\d+)$`)
	matches := re.FindStringSubmatch(episodeKey)
	if len(matches) == 2 {
		return matches[1]
	}
	return "00"
}

// used to get seasonnumber from episodekey in title. e.g: S05E04 -> "05"
func ExtractSeasonNumberFromKey(episodeKey string) string {
	re := regexp.MustCompile(`S(\d+)E\d+`)
	matches := re.FindStringSubmatch(episodeKey)
	if len(matches) == 2 {
		return matches[1]
	}
	return "00"
}

// sometimes the dash is wrong
func NormalizeDash(s string) string {
	// Replace en-dash and em-dash with hyphen-minus
	return strings.NewReplacer("–", "-", "—", "-").Replace(s)
}
