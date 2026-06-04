package metadata

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"opforjellyfin/internal/shared"
)

type ArcMatch struct {
	Season        string
	Name          string
	Range         string
	EpisodeRanges []string
}

func FindArc(index *shared.MetadataIndex, query string) (*ArcMatch, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("arc name cannot be empty")
	}

	arcs := sortedArcs(index)
	if len(arcs) == 0 {
		return nil, fmt.Errorf("no arcs found in metadata index")
	}

	normalizedQuery := normalizeArcQuery(query)

	for _, arc := range arcs {
		if normalizeArcQuery(arc.Name) == normalizedQuery || normalizeArcQuery(arc.Season) == normalizedQuery {
			return &arc, nil
		}
	}

	var matches []ArcMatch
	for _, arc := range arcs {
		if strings.Contains(normalizeArcQuery(arc.Name), normalizedQuery) {
			matches = append(matches, arc)
		}
	}

	if len(matches) == 1 {
		return &matches[0], nil
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("arc name %q matched multiple arcs: %s", query, formatArcNames(matches))
	}

	var closeMatches []ArcMatch
	for _, arc := range arcs {
		if levenshtein(normalizeArcQuery(arc.Name), normalizedQuery) <= 2 {
			closeMatches = append(closeMatches, arc)
		}
	}
	if len(closeMatches) == 1 {
		return &closeMatches[0], nil
	}

	return nil, fmt.Errorf("no arc found matching %q", query)
}

func sortedArcs(index *shared.MetadataIndex) []ArcMatch {
	var arcs []ArcMatch
	if index == nil {
		return arcs
	}

	for seasonKey, season := range index.Seasons {
		name := season.Name
		if name == "" {
			name = seasonKey
		}

		var ranges []string
		for epRange := range season.EpisodeRange {
			ranges = append(ranges, shared.NormalizeDash(epRange))
		}
		sort.Slice(ranges, func(i, j int) bool {
			iStart, _ := shared.ParseRange(ranges[i])
			jStart, _ := shared.ParseRange(ranges[j])
			return iStart < jStart
		})

		arcs = append(arcs, ArcMatch{
			Season:        seasonKey,
			Name:          name,
			Range:         shared.NormalizeDash(season.Range),
			EpisodeRanges: ranges,
		})
	}

	sort.Slice(arcs, func(i, j int) bool {
		return seasonSortValue(arcs[i].Season) < seasonSortValue(arcs[j].Season)
	})

	return arcs
}

func normalizeArcQuery(value string) string {
	value = strings.ToLower(value)
	var builder strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func formatArcNames(arcs []ArcMatch) string {
	names := make([]string, 0, len(arcs))
	for _, arc := range arcs {
		names = append(names, fmt.Sprintf("%s (%s)", arc.Name, arc.Season))
	}
	return strings.Join(names, ", ")
}

func seasonSortValue(season string) int {
	if season == "Specials" {
		return 0
	}

	fields := strings.Fields(season)
	if len(fields) != 2 {
		return 9999
	}

	value, err := strconv.Atoi(fields[1])
	if err != nil {
		return 9999
	}
	return value
}

func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		current := make([]int, len(b)+1)
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			current[j] = minInt(
				current[j-1]+1,
				prev[j]+1,
				prev[j-1]+cost,
			)
		}
		prev = current
	}

	return prev[len(b)]
}

func minInt(values ...int) int {
	min := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
	}
	return min
}
