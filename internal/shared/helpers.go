package shared

import "fmt"

// RangesOverlap returns true if two ranges overlap at all
func RangesOverlap(a1, a2, b1, b2 int) bool {
	return a1 <= b2 && b1 <= a2
}

// NormalizeRange strips leading zeros and normalizes dashes: "001-007" -> "1-7"
func NormalizeRange(r string) string {
	r = NormalizeDash(r)
	a, b := ParseRange(r)
	if a < 0 || b < 0 {
		return r
	}
	return fmt.Sprintf("%d-%d", a, b)
}

// RangeContains returns true if outer fully contains inner
func RangeContains(outerStart, outerEnd, innerStart, innerEnd int) bool {
	return outerStart <= innerStart && innerEnd <= outerEnd
}

// OverlapSize returns the number of overlapping chapters between two ranges, or 0 if none
func OverlapSize(a1, a2, b1, b2 int) int {
	start := a1
	if b1 > start {
		start = b1
	}
	end := a2
	if b2 < end {
		end = b2
	}
	if start > end {
		return 0
	}
	return end - start + 1
}

// RangeSize returns the span of a range
func RangeSize(start, end int) int {
	return end - start + 1
}

// FindMatchingEpisodes finds metadata episodes that match a torrent chapter range.
// Returns all episodes where the overlap covers at least half of the smaller range.
func FindMatchingEpisodes(chapterRange string, index *MetadataIndex) []MatchedEpisode {
	if chapterRange == "" || index == nil {
		return nil
	}

	norm := NormalizeRange(chapterRange)
	tStart, tEnd := ParseRange(norm)
	if tStart < 0 || tEnd < 0 {
		return nil
	}
	tSize := RangeSize(tStart, tEnd)

	var matches []MatchedEpisode

	for seasonKey, season := range index.Seasons {
		// Exact season range match
		sNorm := NormalizeRange(season.Range)
		if sNorm == norm {
			matches = append(matches, MatchedEpisode{
				SeasonKey: seasonKey,
				MatchType: MatchExact,
			})
			return matches
		}

		for epRange, epData := range season.EpisodeRange {
			epNorm := NormalizeRange(epRange)

			// Exact episode match
			if epNorm == norm {
				matches = append(matches, MatchedEpisode{
					SeasonKey:    seasonKey,
					EpisodeRange: epRange,
					EpisodeTitle: epData.Title,
					MatchType:    MatchExact,
				})
				return matches
			}

			// Overlap check
			eStart, eEnd := ParseRange(epNorm)
			if eStart < 0 || eEnd < 0 {
				continue
			}

			overlap := OverlapSize(tStart, tEnd, eStart, eEnd)
			if overlap == 0 {
				continue
			}

			eSize := RangeSize(eStart, eEnd)
			smallerSize := tSize
			if eSize < smallerSize {
				smallerSize = eSize
			}

			// Require overlap covers >50% of the smaller range
			if overlap*2 >= smallerSize {
				matchType := MatchOverlap
				if RangeContains(tStart, tEnd, eStart, eEnd) {
					matchType = MatchContains
				} else if RangeContains(eStart, eEnd, tStart, tEnd) {
					matchType = MatchContainedBy
				}
				matches = append(matches, MatchedEpisode{
					SeasonKey:    seasonKey,
					EpisodeRange: epRange,
					EpisodeTitle: epData.Title,
					MatchType:    matchType,
				})
			}
		}
	}

	return matches
}

// MatchType describes how a torrent range matched a metadata range
type MatchType int

const (
	MatchExact       MatchType = iota // ranges are identical
	MatchContains                     // torrent contains the metadata episode
	MatchContainedBy                  // metadata episode contains the torrent
	MatchOverlap                      // significant partial overlap
)

// MatchedEpisode represents a metadata episode that matched a torrent range
type MatchedEpisode struct {
	SeasonKey    string
	EpisodeRange string
	EpisodeTitle string
	MatchType    MatchType
}
