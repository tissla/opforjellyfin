## bild

# 🏴‍☠️ OpforJellyfin

CLI-tool to automate download and organisation of [One Pace](https://onepace.net) episodes for **Jellyfin**!

> ✨ Torrent downloads
> ✨ Placement after Jellyfin standards
> ✨ Matched to metadata shamelessly stolen from [SpykerNZ/one-pace-for-plex](https://github.com/SpykerNZ/one-pace-for-plex)   

---

## 📸 Example

## bild 2


## 🚀 Installation

1. **Install Go** (version ≥ 1.23)  
2. Clone repo:
   ```bash
   git clone https://github.com/tissla/opforjellyfin.git
   cd opforjellyfin
3. Build binary:
   ```bash
   go build -o opfor


## 🔧 Usage (READ FIRST!)

1. Set your download directory before doing anything else. All your metadata will be stored here, and downloads will be matched to their proper folders.
  ```bash
  ./opfor setDir "/media/anime/One Piece"
  ```
2. Find all available episodes with 'list', or use the -t flag to specify a title, or -r flag to specify a key-range.
  ```bash
  ./opfor list
  ./opfor list -t Wano
  ./opfor list -r 15-20
  ```
3. Download a torrent by using the downloadkey, displayed in front of the title. You can download one or multiple at the same time.
  ```bash
  ./opfor download 15 16 17
  ```

## 📦 Metadata

I hope to continue update the metadata with pictures and new episodes in https://github.com/tissla/one-pace-jellyfin.
the 'sync' command allows the user to stay up to date with new additions to the metadata-repo.

## 🤝 Contributions

All pull requests are welcome. All criticisms are welcome. I'm here to build and to learn and to get better.


## ❤️  Acknowledgements

- SpykerNZ for his metadata
- Anacrolix awesome torrent lib
- Charm team for cool stuff that I should use more
