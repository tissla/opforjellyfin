package shared

import "time"

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
	Title        string
	TorrentID    int
	Started      time.Time
	OutDir       string
	Progress     int64
	TotalSize    int64
	Messages     []string
	Done         bool
	Error        string
	ChapterRange string
}

type TorrentEntry struct {
	Title         string
	Quality       string
	DownloadKey   int
	SeasonName    string
	Seeders       int
	RawIndex      int
	TorrentLink   string
	TorrentID     int
	ChapterRange  string
	MetaDataAvail bool
	IsSpecial     bool
	HaveIt        bool
}
