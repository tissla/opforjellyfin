package shared

import (
	"regexp"
	"sort"
	"strconv"
)

// used by Metadata.go
func RangesOverlap(a1, a2, b1, b2 int) bool {
	return a1 <= b2 && b1 <= a2
}

// file

// sorts season keys of the format "Season \d+"
// "Specials" season special case, equivalent to "Season 0"
func sortKeysByIntSuffix(seasonNames []string) []string {
	re := regexp.MustCompile(`(\d+)$`)
	sort.Slice(seasonNames, func(i, j int) bool {
		// extract season number, convert to int, and compare
		a := re.FindString(seasonNames[i])
		b := re.FindString(seasonNames[j])
		ai, _ := strconv.Atoi(a)
		bi, _ := strconv.Atoi(b)
		return ai < bi
	})
	return seasonNames
}

// SortMetadataSeasons convert a MetadataIndex into an equivalent sorted version
func SortMetadataSeasons(index *MetadataIndex) *SortedMetadataIndex {
	seasonNames := make([]string, 0, len(index.Seasons))
	for k := range index.Seasons {
		seasonNames = append(seasonNames, k)
	}
	seasonNames = sortKeysByIntSuffix(seasonNames)

	result := &SortedMetadataIndex{
		Seasons: make([]SeasonEntry, 0, len(index.Seasons)),
	}

	for _, k := range seasonNames {
		result.Seasons = append(result.Seasons, SeasonEntry{k, index.Seasons[k]})
	}

	return result
}
