package core

import (
	"strings"
)

type MirrorArgs struct {
	Link         string
	CustomName   string
	CustomRemote string
	IsMagnet     bool
}

// ParseMirrorArgs mem-parsing argumen command mirip seperti WZML (wzv3)
func ParseMirrorArgs(rawText string) MirrorArgs {
	res := MirrorArgs{}

	// Bersihkan prefix command seperti "/mirror"
	words := strings.Fields(rawText)
	if len(words) > 0 && strings.HasPrefix(words[0], "/") {
		words = words[1:]
	}

	joined := strings.Join(words, " ")

	// Cek format pipa: /mirror <link> | <nama_baru>
	if strings.Contains(joined, "|") {
		parts := strings.SplitN(joined, "|", 2)
		res.Link = strings.TrimSpace(parts[0])
		res.CustomName = strings.TrimSpace(parts[1])
		if strings.HasPrefix(res.Link, "magnet:?") {
			res.IsMagnet = true
		}
		return res
	}

	// Parsing flags: -n (nama), -rc (remote rclone)
	var linkParts []string
	for i := 0; i < len(words); i++ {
		w := words[i]
		if w == "-n" && i+1 < len(words) {
			res.CustomName = words[i+1]
			i++
		} else if w == "-rc" && i+1 < len(words) {
			res.CustomRemote = words[i+1]
			i++
		} else {
			linkParts = append(linkParts, w)
		}
	}

	res.Link = strings.TrimSpace(strings.Join(linkParts, " "))
	if strings.HasPrefix(res.Link, "magnet:?") {
		res.IsMagnet = true
	}

	return res
}
