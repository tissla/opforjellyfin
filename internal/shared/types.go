package shared

import "time"

// TODO: cleanup unused properties

// config file
type Config struct {
	TargetDir     string `json:"target_dir"`
	GitHubRepo    string `json:"github_base_url"`
	TorrentAPIURL string `json:"torrent_api_url"`
}

// Index maps seasons
type MetadataIndex struct {
	Seasons map[string]SeasonIndex `json:"seasons"`
}

// seasons maps episodes
type SeasonIndex struct {
	Range        string                 `json:"range"`
	Name         string                 `json:"name"`
	EpisodeRange map[string]EpisodeData `json:"episodes"`
}

// episodes maps titles and have
type EpisodeData struct {
	Title string `json:"title"`
}

// download struct
type TorrentDownload struct {
	Title             string    // title for display
	FullTitle         string    // full torrent title
	TorrentID         int       // torrentID for tempdir
	ChapterRange      string    // Main
	Started           time.Time // time torrent started (unused?)
	Progress          int64     // used by ui progressbar
	TotalSize         int64     // used by ui progress bar
	PlacementFull     []string  // used to display placed messages after all placements are done
	PlacementProgress string    //used for placement messages after download is done
	Done              bool      // set to true when torrent is downloaded
	Placed            bool      // set to true when files are placed, before clearing active downloads
}

// entry for dl
type TorrentEntry struct {
	Title         string // full title
	Quality       string // parsed quality
	DownloadKey   int    // download key set by rawIndex
	TorrentName   string // for display
	Seeders       int    // number of seeders
	RawIndex      int    // RawIndex is based on ChapterRange, used for placement
	TorrentLink   string // torrent link
	TorrentID     int    // torrent ID, extracted from link
	ChapterRange  string // torrent chapter range
	MetaDataAvail bool
	IsSpecial     bool
	HaveIt        int
}
