package shared

import "time"

// TODO: cleanup unused properties

// config file
type Config struct {
	TargetDir  string        `json:"target_dir"`
	GitHubRepo string        `json:"github_base_url"`
	Source     ScraperConfig `json:"source"`
}

// scrape config
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
	FileSize 		string `json:"file_size"`
}

type ValidationConfig struct {
	RequiredInTitle string `json:"required_in_title"`
}

// Index maps seasons
type MetadataIndex struct {
	Seasons map[string]SeasonIndex `json:"seasons"`
}

type SortedMetadataIndex struct {
	Seasons []SeasonEntry `json:"seasons"`
}

// used for sorting MetadataIndex.Seasons
// Title is equivalent to the key in the map
// SeasonIndex is the associated value
type SeasonEntry struct {
	Title       string      `json:"title"`
	SeasonIndex SeasonIndex `json:"index"`
}

// seasons maps episodes
type SeasonIndex struct {
	Range        string                 `json:"range"`
	Name         string                 `json:"name"`
	EpisodeRange map[string][]EpisodeData `json:"episodes"`
}

// episodes maps titles and have
type EpisodeData struct {
	Title string `json:"title"`
}

type FilePlacement struct {
	Title             string   // title for display
	FullTitle         string   // full file title
	ChapterRange      string   // Main
	PlacementFull     []string // used to display placed messages after all placements are done
	PlacementProgress string   // used for placement messages after once placement is started
	Placed            bool     // set to true when files are placed
}

// download struct
type TorrentDownload struct {
	FilePlacement
	Entry 		TorrentEntry  // fields from TorrentEntry for easy access during placement
	TorrentID int       		// torrentID for tempdir
	Started   time.Time 		// time torrent started (unused?)
	Progress  int64     		// used by ui progressbar
	TotalSize int64     		// used by ui progress bar
	Done      bool      		// set to true when torrent is downloaded
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
	MetaDataAvail bool   // metadata matching chapter range exists
	IsSpecial     bool   // is a special (no chapter range)
	HaveIt        int    // video with same chapter range exists
	Date          time.Time // the upload date of the torrent, used for sorting and display
	FileSize			string // file size as displayed on the torrent site
	IsExtended    bool   // extended version
}
