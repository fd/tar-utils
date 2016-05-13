package version

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const payloadData = "e927cc4876808ab86054e3489a04efd20bc9cf9f3fe2356e56b1274aa8ff4fc0dfa8f97203153fb75a3e6f274c84094b5b20a0306943e121ce818b5af8333c9ebaf084abf27f78effaf7ea1c36ed89bdf8ff8a369da01388206d987a52ed22cb29fa600d61da0772c5822499337bc8ad8655ebe185bfff5c4eaba1d4de5a577863ed661607379003b94374dd85b0c35e24dcfc3dabb0147607582c7402a782be5fc0a19b7a92fb0c91599ed5dfe25cf180bc675cef87ccd1f79ba86c72768ab7862831bfaac0da54bca6166c78dbc558e2e324e5f85ee22156901b0e82c2ac9d2acf29ed11ae86852a57c3c53ef75d292d0c5d21cf1deb7e5fb1bd641fb46a97718f7983260b9415eb0b240731de0359bf1e3764954cfb94277bdf972b13eaa6a38c0e3bb0be58fa850857b774f325e336dcf2550644ecaa1edcdea3b44e7632c5ae7d723d4e8692c04e1d1a9fa64cdce23082a9032f34027a994ac6a13ddc7b9d3204350908fe1567bff31ece702051446e42e8ea7dfed88f88ad42de1b0effb19ccd1da94462d17411edc6fb510175912ea455bc7387e2040cbf0cd79ef4baae27375fae38f5351f5cf4ebd540d7d560eb7cbe8d6aa2e040faa0a2c00f8a32759f5a1bafee6ba690192b64bee612dfe6e142d3ff53854adb91e2da8f86f0a58685d12b832e083baf61ad80f6a353224e16bc7693585e2894147b8286985032"

var payload = Info{PreRelease: []string{"dev"}}

func init() {
	if strings.HasPrefix(payloadData, "e927cc4") {
		// payload is unset
		return
	}

	var info Info

	err := json.Unmarshal([]byte(payloadData), &info)
	if err != nil {
		return
	}

	payload = info
}

// Get version info
func Get() *Info {
	return payload.clone()
}

// Info holds all version info
type Info struct {
	// Major.Minor.Patch-Pre+Extra
	Major      int
	Minor      int
	Patch      int
	PreRelease []string
	Metadata   []string
	Released   time.Time
	ReleasedBy string
	Commit     string
}

func (i *Info) clone() *Info {
	c := &Info{}
	*c = *i
	c.PreRelease = nil
	c.Metadata = nil
	c.PreRelease = append(c.PreRelease, i.PreRelease...)
	c.Metadata = append(c.Metadata, i.Metadata...)
	return c
}

func (i *Info) String() string {
	return i.Semver()
}

// Semver returns the version info in semver 2 format
func (i *Info) Semver() string {
	s := fmt.Sprintf("%d.%d.%d", i.Major, i.Minor, i.Patch)
	if len(i.PreRelease) > 0 {
		s += "-" + strings.Join(i.PreRelease, ".")
	}
	if len(i.Metadata) > 0 {
		s += "+" + strings.Join(i.Metadata, ".")
	}
	return s
}

// ParseSemver parses a semver 2 string
func ParseSemver(s string) (*Info, error) {
	re := regexp.MustCompile(`^([0-9]+)(?:\.([0-9]+)(?:\.([0-9]+))?)?(?:[-]([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:[+]([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)
	m := re.FindStringSubmatch(s)
	if m == nil {
		return nil, errors.New("invalid semver")
	}

	var (
		major, _ = strconv.Atoi(m[1])
		minor, _ = strconv.Atoi(m[2])
		patch, _ = strconv.Atoi(m[3])
		pre      = strings.Split(m[4], ".")
		meta     = strings.Split(m[5], ".")
	)

	if m[4] == "" {
		pre = nil
	}
	if m[5] == "" {
		meta = nil
	}

	return &Info{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: pre,
		Metadata:   meta,
	}, nil
}

type BumpType uint8

const (
	// What to bump
	ReleaseCandidate BumpType = 1 << iota
	Patch
	Minor
	Major
	Final
)

func Bump(i *Info, typ BumpType) *Info {
	i = i.clone()

	if typ&Final > 0 {
		i.PreRelease = nil
	}
	if typ&Patch > 0 {
		i.PreRelease = nil
		i.Patch++
	}
	if typ&Minor > 0 {
		i.PreRelease = nil
		i.Patch = 0
		i.Minor++
	}
	if typ&Major > 0 {
		i.PreRelease = nil
		i.Patch = 0
		i.Minor = 0
		i.Major++
	}
	if typ&ReleaseCandidate > 0 {
		var found bool
		for idx, v := range i.PreRelease {
			if strings.HasPrefix(v, "rc") {
				d, err := strconv.Atoi(v[2:])
				if err != nil {
					continue
				}
				found = true
				d++
				i.PreRelease[idx] = "rc" + strconv.Itoa(d)
			}
		}
		if !found {
			i.PreRelease = []string{"rc0"}
		}
	}

	return i
}
