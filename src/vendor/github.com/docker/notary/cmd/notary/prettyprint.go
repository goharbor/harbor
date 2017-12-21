package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/docker/notary/client"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
)

const (
	fourItemRow = "%s\t%s\t%s\t%s\n"
	fiveItemRow = "%s\t%s\t%s\t%s\t%s\n"
)

func initTabWriter(columns []string, writer io.Writer) *tabwriter.Writer {
	tw := tabwriter.NewWriter(writer, 4, 4, 4, ' ', 0)
	fmt.Fprintln(tw, strings.Join(columns, "\t"))
	breakLine := make([]string, 0, len(columns))
	for _, h := range columns {
		breakLine = append(
			breakLine,
			strings.Repeat("-", len(h)),
		)
	}
	fmt.Fprintln(tw, strings.Join(breakLine, "\t"))
	return tw
}

// --- pretty printing certs ---

func truncateWithEllipsis(str string, maxWidth int, leftTruncate bool) string {
	if len(str) <= maxWidth {
		return str
	}
	if leftTruncate {
		return fmt.Sprintf("...%s", str[len(str)-(maxWidth-3):])
	}
	return fmt.Sprintf("%s...", str[:maxWidth-3])
}

const (
	maxGUNWidth = 25
	maxLocWidth = 40
)

type keyInfo struct {
	gun      data.GUN // assumption that this is "" if role is root
	role     data.RoleName
	keyID    string
	location string
}

// We want to sort by gun, then by role, then by keyID, then by location
// In the case of a root role, then there is no GUN, and a root role comes
// first.
type keyInfoSorter []keyInfo

func (k keyInfoSorter) Len() int      { return len(k) }
func (k keyInfoSorter) Swap(i, j int) { k[i], k[j] = k[j], k[i] }
func (k keyInfoSorter) Less(i, j int) bool {
	// special-case role
	if k[i].role != k[j].role {
		if k[i].role == data.CanonicalRootRole {
			return true
		}
		if k[j].role == data.CanonicalRootRole {
			return false
		}
		// otherwise, neither of them are root, they're just different, so
		// go with the traditional sort order.
	}

	// sort order is GUN, role, keyID, location.
	orderedI := []string{k[i].gun.String(), k[i].role.String(), k[i].keyID, k[i].location}
	orderedJ := []string{k[j].gun.String(), k[j].role.String(), k[j].keyID, k[j].location}

	for x := 0; x < 4; x++ {
		switch {
		case orderedI[x] < orderedJ[x]:
			return true
		case orderedI[x] > orderedJ[x]:
			return false
		}
		// continue on and evalulate the next item
	}
	// this shouldn't happen - that means two values are exactly equal
	return false
}

// Given a list of KeyStores in order of listing preference, pretty-prints the
// root keys and then the signing keys.
func prettyPrintKeys(keyStores []trustmanager.KeyStore, writer io.Writer) {
	var info []keyInfo

	for _, store := range keyStores {
		for keyID, keyIDInfo := range store.ListKeys() {
			info = append(info, keyInfo{
				role:     keyIDInfo.Role,
				location: store.Name(),
				gun:      keyIDInfo.Gun,
				keyID:    keyID,
			})
		}
	}

	if len(info) == 0 {
		writer.Write([]byte("No signing keys found.\n"))
		return
	}

	sort.Stable(keyInfoSorter(info))

	tw := initTabWriter([]string{"ROLE", "GUN", "KEY ID", "LOCATION"}, writer)

	for _, oneKeyInfo := range info {
		fmt.Fprintf(
			tw,
			fourItemRow,
			oneKeyInfo.role,
			truncateWithEllipsis(oneKeyInfo.gun.String(), maxGUNWidth, true),
			oneKeyInfo.keyID,
			truncateWithEllipsis(oneKeyInfo.location, maxLocWidth, true),
		)
	}
	tw.Flush()
}

// --- pretty printing targets ---

type targetsSorter []*client.TargetWithRole

func (t targetsSorter) Len() int      { return len(t) }
func (t targetsSorter) Swap(i, j int) { t[i], t[j] = t[j], t[i] }
func (t targetsSorter) Less(i, j int) bool {
	return t[i].Name < t[j].Name
}

// --- pretty printing roles ---

type roleSorter []data.Role

func (r roleSorter) Len() int      { return len(r) }
func (r roleSorter) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r roleSorter) Less(i, j int) bool {
	return r[i].Name < r[j].Name
}

// Pretty-prints the sorted list of TargetWithRoles.
func prettyPrintTargets(ts []*client.TargetWithRole, writer io.Writer) {
	if len(ts) == 0 {
		writer.Write([]byte("\nNo targets present in this repository.\n\n"))
		return
	}

	sort.Stable(targetsSorter(ts))

	tw := initTabWriter([]string{"NAME", "DIGEST", "SIZE (BYTES)", "ROLE"}, writer)

	for _, t := range ts {
		fmt.Fprintf(
			tw,
			fourItemRow,
			t.Name,
			hex.EncodeToString(t.Hashes["sha256"]),
			fmt.Sprintf("%d", t.Length),
			t.Role,
		)
	}
	tw.Flush()
}

// Pretty-prints the list of provided Roles
func prettyPrintRoles(rs []data.Role, writer io.Writer, roleType string) {
	if len(rs) == 0 {
		writer.Write([]byte(fmt.Sprintf("\nNo %s present in this repository.\n\n", roleType)))
		return
	}

	// this sorter works for Role types
	sort.Stable(roleSorter(rs))

	tw := initTabWriter([]string{"ROLE", "PATHS", "KEY IDS", "THRESHOLD"}, writer)

	for _, r := range rs {
		var path, kid string
		pp := prettyPaths(r.Paths)
		if len(pp) > 0 {
			path = pp[0]
		}
		if len(r.KeyIDs) > 0 {
			kid = r.KeyIDs[0]
		}
		fmt.Fprintf(
			tw,
			fourItemRow,
			r.Name,
			path,
			kid,
			fmt.Sprintf("%v", r.Threshold),
		)
		printExtraRoleRows(tw, pp, r.KeyIDs)
	}
	tw.Flush()
}

func printExtraRoleRows(tw *tabwriter.Writer, paths, keyIDs []string) {
	lPaths := len(paths)
	lKeyIDs := len(keyIDs)
	longer := len(keyIDs)
	if len(paths) > len(keyIDs) {
		longer = len(paths)
	}
	for i := 1; i < longer; i++ {
		var path, kid string
		if lPaths > i {
			path = paths[i]
		}
		if lKeyIDs > i {
			kid = keyIDs[i]
		}
		fmt.Fprintf(
			tw,
			fourItemRow,
			"",
			path,
			kid,
			"",
		)
	}
}

// Pretty-formats a list of delegation paths, and ensures the empty string is printed as "" in the console
func prettyPaths(paths []string) []string {
	// sort paths first
	sort.Strings(paths)
	pp := make([]string, 0, len(paths))
	for _, path := range paths {
		// manually escape "" and designate that it is all paths with an extra print <all paths>
		if path == "" {
			path = "\"\" <all paths>"
		}
		pp = append(pp, path)
	}
	return pp
}
