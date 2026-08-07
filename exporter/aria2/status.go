package aria2

import (
	"path/filepath"

	"github.com/deorth-kku/aria2rpc-go"
)

// bittorrent, gid, name
func TaskLabels(stat aria2rpc.Status) []string {
	if stat.BitTorrent != nil && len(stat.BitTorrent.Info.Name) != 0 {
		return []string{"true", stat.GID, stat.BitTorrent.Info.Name}
	}
	return []string{"false", stat.GID, filepath.Base(stat.Files[0].Path)}
}
