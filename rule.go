package reverseproxy

import (
	"net/http"
	"net/url"
	pathlib "path"
	"strings"
)

// Rule stores information about how to proxy a given URL to another URL.
type Rule struct {
	SourceHost string `json:source_host`
	SourcePath string `json:source_path`
	DestHost   string `json:dest_host`
	DestPath   string `json:dest_path`
	DestScheme string `json:dest_scheme`

	// CaseSensitiveHost determines if SourceHost should be matched in a
	// case-sensitive way.
	CaseSensitiveHost bool `json:case_sensitive_host`

	// CaseSensitivePath determines if SourcePath should be matched in a
	// case-sensitive way.
	CaseSensitivePath bool `json:case_sensitive_path`

	// CleanRequestPath indicates whether or not incoming paths should be
	// normalized before being matched. For example, the path "foo/../bar" would
	// be normalized to "/bar" if this flag were set.
	CleanRequestPath bool `json:clean_request_path`
}

// MatchesRequest determines whether or not a given request matches the source
// parameters of a given rule.
func (r Rule) MatchesRequest(req *http.Request) bool {
	if !r.CaseSensitiveHost {
		if strings.ToLower(r.SourceHost) != strings.ToLower(req.Host) {
			return false
		}
	} else if r.SourceHost != req.Host {
		return false
	}

	// If the source path is not absolute, anything should match. If the
	// destination is relative but the source is absolute, it should not match.
	reqPath := r.reqPath(req)
	if !pathlib.IsAbs(r.SourcePath) {
		return true
	} else if !pathlib.IsAbs(reqPath) {
		return false
	}
	return PathContains(r.SourcePath, reqPath, r.CaseSensitivePath)
}

// DestinationURL provides the url.URL that a given request should be proxied
// to.
func (r Rule) DestinationURL(req *http.Request) url.URL {
	newURL := *req.URL
	newURL.Scheme = r.DestScheme
	newURL.Host = r.DestHost

	// Compute the new path if needed
	if pathlib.IsAbs(r.SourcePath) {
		rel := RelativePath(r.SourcePath, r.reqPath(req), r.CaseSensitivePath)
		if len(r.DestPath) == 0 {
			newURL.Path = "/" + rel
		} else {
			newURL.Path = pathlib.Join(r.DestPath, rel)
		}
	}

	return newURL
}

func (r Rule) reqPath(req *http.Request) string {
	if r.CleanRequestPath {
		return pathlib.Clean(req.URL.Path)
	} else {
		return req.URL.Path
	}
}
