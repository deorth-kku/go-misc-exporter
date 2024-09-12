package aria2

import (
	"path/filepath"

	"github.com/siku2/arigo"
)

// bittorrent, gid, name
func TaskLabels(stat arigo.Status) []string {
	if len(stat.BitTorrent.Info.Name) != 0 {
		return []string{"true", stat.GID, stat.BitTorrent.Info.Name}
	}
	return []string{"false", stat.GID, filepath.Base(stat.Files[0].Path)}
}
