package wlmanip

import (
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/badvassal/wllib/decode"
	"github.com/badvassal/wllib/decode/action"
	"github.com/badvassal/wllib/defs"
)

// TransEntry represents a single transition.  It is annotated with some extra
// information that allows it to be used for replacing one transition with
// another.
type TransEntry struct {
	FromBlock defs.BlockZIP
	FromLoc   int // Inexact (i.e., never a sub-location).
	Trans     action.Transition
	Selector  int

	FromExactLoc int
	ToExactLoc   int
}

// LocPairMap maintains the full set of transitions among all MSQ blocks.
// [exact-from-loc][exact-to-loc].
type LocPairMap map[int]map[int][]*TransEntry

// Collection is the set of all usable transitions among all MSQ blocks.
// Elements in its slices can be modified, but none should be added or removed.
type Collection struct {
	unfiltered LocPairMap // All transitions.
	filtered   LocPairMap // Filtered by a CollectCfg.
}

// CollectCfg specifies which transitions to keep and which to filter.
type CollectCfg struct {
	KeepWorld          bool
	KeepRelative       bool
	KeepShops          bool
	KeepDerelict       bool
	KeepPrevious       bool
	KeepAutoIntra      bool
	KeepHardcodedIntra bool
	KeepPostSewers     bool
}

// shouldKeepTransition indicates whether a given transition should be kept
// according to a CollectCfg.
func shouldKeepTransition(entry TransEntry, cfg CollectCfg) bool {
	logDiscard := func(reason string) {
		log.Debugf("discarding transition (%s -> %s) %+v: %s",
			LocationString(entry.FromExactLoc),
			LocationString(entry.ToExactLoc),
			entry, reason)
	}

	if !cfg.KeepWorld {
		if entry.FromBlock.GameIdx == 0 &&
			entry.FromBlock.BlockIdx == defs.Block0WorldMap {

			logDiscard("from world map")
			return false
		}
		if entry.Trans.Location == defs.LocationWorldMap {
			logDiscard("to world map")
			return false
		}
	}

	if !cfg.KeepRelative && entry.Trans.Relative {
		logDiscard("relative")
		return false
	}

	if !cfg.KeepShops && entry.Trans.ToClass == action.IDShop {
		logDiscard("shop")
		return false
	}

	if !cfg.KeepDerelict {
		if entry.Trans.IsDerelict() {
			logDiscard("derelict")
			return false
		}
	}

	if !cfg.KeepPrevious && entry.Trans.Location == defs.LocationPrevious {
		logDiscard("previous")
		return false
	}

	if !cfg.KeepPostSewers && LocationPostSewersMap[entry.Trans.Location] {
		logDiscard("post sewers")
		return false
	}

	if !cfg.KeepAutoIntra && entry.Trans.Location == entry.FromLoc {
		desc := SubLocDesc{
			GameIdx:  entry.FromBlock.GameIdx,
			BlockIdx: entry.FromBlock.BlockIdx,
			Selector: entry.Selector,
		}
		loc := selectorToSubLocs(desc)
		if loc.From == -1 && loc.To == -1 {
			logDiscard("auto intra filter")
			return false
		}
	}

	if !cfg.KeepHardcodedIntra &&
		TransitionIsIntra(defs.LocPair{entry.FromLoc, entry.Trans.Location}) {

		logDiscard("hardcoded intra filter")
		return false
	}

	return true
}

// selectorToSubLocs determine the exact from/to location codes for the given
// transition selector.
func selectorToSubLocs(desc SubLocDesc) defs.LocPair {
	pair, ok := SubLocMap[desc]
	if ok {
		log.Debugf("translated %+v to sub location pair %+v", desc, pair)
		return pair
	}

	return defs.LocPair{-1, -1}
}

// collectTransitions gathers the full set of transitions from among all MSQ
// blocks.  The resulting list is unlitered.
func collectTransitions(state decode.DecodeState, cfg CollectCfg) ([]*TransEntry, error) {
	collectGame := func(gameIdx int) ([]*TransEntry, error) {
		var entries []*TransEntry

		for blockIdx, block := range state.Blocks[gameIdx] {
			for selector, t := range block.ActionTables.Transitions {
				if t != nil {
					zip := defs.BlockZIP{
						GameIdx:  gameIdx,
						BlockIdx: blockIdx,
					}
					from, err := defs.BlockZIPToLoc(zip)
					if err != nil {
						return nil, err
					}

					exactLocs := selectorToSubLocs(SubLocDesc{
						GameIdx:  gameIdx,
						BlockIdx: blockIdx,
						Selector: selector,
					})
					if exactLocs.From == -1 {
						exactLocs.From = from
					}
					if exactLocs.To == -1 {
						exactLocs.To = t.Location
					}

					entry := &TransEntry{
						FromBlock:    zip,
						FromLoc:      from,
						Trans:        *t,
						Selector:     selector,
						FromExactLoc: exactLocs.From,
						ToExactLoc:   exactLocs.To,
					}

					entries = append(entries, entry)
				}
			}
		}

		return entries, nil
	}

	es0, err := collectGame(0)
	if err != nil {
		return nil, err
	}

	es1, err := collectGame(1)
	if err != nil {
		return nil, err
	}

	return append(es0, es1...), nil
}

// Collect gathers the transitions from among all MSQ blocks and constructs a
// Collection.
func Collect(state decode.DecodeState, cfg CollectCfg) (*Collection, error) {
	if err := FixupTransitions(&state); err != nil {
		return nil, err
	}

	entries, err := collectTransitions(state, cfg)
	if err != nil {
		return nil, err
	}

	coll := &Collection{
		unfiltered: map[int]map[int][]*TransEntry{},
		filtered:   map[int]map[int][]*TransEntry{},
	}

	add := func(m LocPairMap, e *TransEntry) {
		if m[e.FromExactLoc] == nil {
			m[e.FromExactLoc] = map[int][]*TransEntry{}
		}
		m[e.FromExactLoc][e.ToExactLoc] =
			append(m[e.FromExactLoc][e.ToExactLoc], e)
	}

	for _, e := range entries {
		add(coll.unfiltered, e)

		if shouldKeepTransition(*e, cfg) {
			add(coll.filtered, e)
		}
	}

	for from, m := range coll.filtered {
		for to, _ := range m {
			if coll.filtered[to][from] == nil {
				log.Debugf("discarding entry: %s --> %s: no round trip",
					LocationFullString(from), LocationFullString(to))
				delete(m, to)
				if len(m) == 0 {
					delete(coll.filtered, from)
					break
				}
			}
		}
	}

	return coll, nil
}

func (c *Collection) GetFiltered(lp defs.LocPair) []*TransEntry {
	m := c.filtered[lp.From]
	if m == nil {
		return nil
	}

	return m[lp.To]
}

func (c *Collection) GetUnfiltered(lp defs.LocPair) []*TransEntry {
	m := c.unfiltered[lp.From]
	if m == nil {
		return nil
	}

	return m[lp.To]
}

// Get1WayUp retrieves the set of unfiltered transitions from the given
// location that have the following properties:
// 1. Lead to a lesser depth location, and
// 2. Are one way (no return trip).
func (c *Collection) Get1WayUp(from int) []*TransEntry {
	var entries []*TransEntry

	for to, m := range c.unfiltered[from] {
		// Only consider upward transitions.
		if LocationDepthMap[to] < LocationDepthMap[from] {
			// Only consider one-way transitions.
			if len(c.unfiltered[to][from]) == 0 {
				for _, es := range m {
					entries = append(entries, es)
				}
			}
		}
	}

	return entries
}

// FilteredRoundTrips retrieves the set of round trip transitions from the
// filtered list.  Return trips are not included in the returned slice.  In
// other words, x-->y is returned only if both x-->y and y-->x are present in
// the filtered list.  In this case, ONLY x-->y is returned (not y-->x).
func (c *Collection) FilteredRoundTrips() []defs.LocPair {
	var pairs []defs.LocPair

	// Don't add two pairs that mirror each other.  Reverse routes get added by
	// the caller.
	seen := map[defs.LocPair]struct{}{}

	for from, m := range c.filtered {
		for to, _ := range m {

			fromDepth := LocationDepthMap[from]
			toDepth := LocationDepthMap[to]
			if toDepth >= fromDepth {
				if _, ok := seen[defs.LocPair{to, from}]; ok {
					continue
				}
				seen[defs.LocPair{from, to}] = struct{}{}
				pairs = append(pairs, defs.LocPair{
					From: from,
					To:   to,
				})
			}
		}
	}

	sort.Slice(pairs, func(i int, j int) bool {
		if pairs[i].From < pairs[j].From {
			return true
		}
		if pairs[i].From > pairs[j].From {
			return false
		}
		return pairs[i].To < pairs[j].To
	})

	return pairs
}
