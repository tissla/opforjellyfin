package shared

import "time"

// TODO: cleanup unused properties

// config file
type Config struct {
	TargetDir     string `json:"target_dir"`
	GitHubRepo    string `json:"github_base_url"`
	TorrentAPIURL string `json:"torrent_api_url"`
}

// MetadataIndex represents structured metadata for seasons and episodes.
type MetadataIndex struct {
	Seasons map[string]SeasonIndex `json:"seasons"`
}

// SeasonIndex represents episodes and their manga chapter ranges.
type SeasonIndex struct {
	Range    string            `json:"range"`
	Episodes map[string]string `json:"episodes"`
}

// download struct
type TorrentDownload struct {
	Title           string
	TorrentID       int
	Started         time.Time
	OutDir          string
	Progress        int64    // used by ui
	TotalSize       int64    // used by ui
	Messages        []string // used
	Error           string   // stores specific torrent errors
	ChapterRange    string
	ProgressMessage string //used for placement messages after download is done
	Done            bool   // set to true when torrent is downloaded
	Placed          bool   // set to true when files are placed, before clearing active downloads
}

// entry for dl
type TorrentEntry struct {
	Title         string
	Quality       string
	DownloadKey   int
	TorrentName   string
	Seeders       int
	RawIndex      int // RawIndex is based on ChapterRange, used for placement
	TorrentLink   string
	TorrentID     int
	ChapterRange  string
	MetaDataAvail bool
	IsSpecial     bool
	HaveIt        int
}
