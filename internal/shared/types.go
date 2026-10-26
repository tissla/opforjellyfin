package shared

// TODO: cleanup unused properties

type Config struct {
	TargetDir  string        `json:"target_dir"`
	GitHubRepo string        `json:"github_base_url"`
	Source     ScraperConfig `json:"source"`
	Thumbnails bool          `json:"thumbnails"`
}

type ScraperConfig struct {
	Name               string           `json:"name"`
	BaseURL            string           `json:"base_url"`
	SearchPathTemplate string           `json:"search_path_template"`
	SearchQuery        string           `json:"search_query"`
	RowSelector        string           `json:"row_selector"`
	Fields             ScraperFields    `json:"fields"`
	Validation         ValidationConfig `json:"validation"`
}

type ScraperFields struct {
	Title       string `json:"title"`
	Seeders     string `json:"seeders"`
	TorrentLink string `json:"torrent_link"`
	TorrentID   string `json:"torrent_id"`
	UploadDate  string `json:"upload_date"`
}

type ValidationConfig struct {
	RequiredInTitle string `json:"required_in_title"`
}

type MetadataIndex struct {
	ByFilename map[string]string       `json:"by_filename"`
	Entries    map[string]MetadataFile `json:"entries"`
	Arcs       map[string]MetadataArc  `json:"arcs"`
}

type MetadataFile struct {
	NFOPath   string `json:"nfo_path"`
	EpisodeID string `json:"episode_id"`
	Kind      string `json:"kind"`
	Variant   string `json:"variant"`
}

type MetadataArc struct {
	Folder       string `json:"folder"`
	Name         string `json:"name"`
	SeasonNumber int    `json:"season_number"`
}

// TorrentDownload download struct
type TorrentDownload struct {
	Title             string   // title for display
	FullTitle         string   // full torrent title
	TorrentID         int      // torrentID for tempdir
	ChapterRange      string   // Main
	Progress          int64    // used by ui progressbar
	TotalSize         int64    // used by ui progress bar
	PlacementFull     []string // used to display placed messages after all placements are done
	PlacementProgress string   //used for placement messages after download is done
	Done              bool     // set to true when torrent is downloaded
	Placed            bool     // set to true when files are placed, before clearing active downloads
	VideoPaths        []string // actual files listed by the torrent
}

// TorrentEntry entry for dl
type TorrentEntry struct {
	Title         string // full title
	Quality       string // parsed quality
	DownloadKey   int    // download key set by rawIndex
	TorrentName   string // for display
	Seeders       int    // number of seeders
	RawIndex      int    // sort order in search results
	TorrentLink   string // torrent link
	TorrentID     int    // torrent ID, extracted from link
	ChapterRange  string // torrent chapter range
	MetaDataAvail bool   // metadata exists for the episode or arc title
	IsSpecial     bool   // is a special (no chapter range)
	HaveIt        int    // 0 = none, 1 = some, 2 = all
	Date          string //

	SelectedFiles []string `json:"-"` // nil selects all .mkv and .mp4 files
}
