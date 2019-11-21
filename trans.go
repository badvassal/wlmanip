package wlmanip

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/badvassal/wllib/decode"
	"github.com/badvassal/wllib/decode/action"
	"github.com/badvassal/wllib/defs"
)

// A TransOP is used to replace one transition with another.  It consists of four transitions:
//
// 1. A.From --> A.To
// 2. A.From <-- A.To (A return trip)
// 3. B.From --> B.To
// 4. B.From <-- B.To (B return trip)
type TransOp struct {
	A defs.LocPair
	B defs.LocPair
}

type TransXList struct {
	Black []int
	White []int
}

type TransXListPair struct {
	Read  TransXList
	Write TransXList
}

// CopyTrans replaces a destination transition with a source.  A few fields in
// the destintion are preserved to maintain data integrity in the parent MSQ
// block.
func CopyTrans(dst *action.Transition, src action.Transition) {
	dst.Relative = src.Relative
	dst.Prompt = src.Prompt
	dst.LocX = src.LocX
	dst.LocY = src.LocY
	dst.Location = src.Location
}

// delistEntries applies white lists and black lists to a list of entries.
func delistEntries(entries []*TransEntry, isRead bool) []*TransEntry {
	if len(entries) == 0 {
		return nil
	}

	var filtered []*TransEntry
	for _, e := range entries {
		keep := true

		from := e.FromExactLoc
		to := e.ToExactLoc

		pair := LocationXListPairMap[defs.LocPair{from, to}]
		var xlist TransXList
		if isRead {
			xlist = pair.Read
		} else {
			xlist = pair.Write
		}

		entryStr := fmt.Sprintf("%s,%d,", LocationString(from), e.Selector)
		if isRead {
			entryStr += "read"
		} else {
			entryStr += "write"
		}

		if len(xlist.White) > 0 {
			keep = false
			for _, w := range xlist.White {
				if e.Selector == w {
					log.Debugf("whitelisting %s", entryStr)
					keep = true
					break
				}
			}
		}

		if !keep {
			log.Debugf("delisting %s", entryStr)
		} else {
			for _, b := range xlist.Black {
				if e.Selector == b {
					log.Debugf("blacklisting %s", entryStr)
					keep = false
					break
				}
			}
		}

		if keep {
			filtered = append(filtered, e)
		}
	}

	return filtered
}

// transOpCtxt contains context needed to execute a single transition op.
type transOpCtxt struct {
	AFwd []*TransEntry
	ARev []*TransEntry
	ADB  decode.Block

	BFwd []*TransEntry
	BRev []*TransEntry
	BDB  decode.Block

	BRev1WayUp []*TransEntry
}

func newTransOpCtxt(coll *Collection, state *decode.DecodeState,
	op TransOp) *transOpCtxt {

	// We only filter the reverse routes in A and the forward routes in B;
	// everything else is unfiltered.  Filtering is only necessary to restrict
	// the transitions which we copy *from*.  When we replace a journey, we
	// want to copy *to* all the selectors.  In other words, filter the reads,
	// not the writes.
	aFwd := coll.GetUnfiltered(op.A)
	aRev := coll.GetFiltered(defs.LocPair{op.A.To, op.A.From})
	bFwd := coll.GetFiltered(op.B)
	bRev := coll.GetUnfiltered(defs.LocPair{op.B.To, op.B.From})

	if len(aFwd) == 0 || len(bFwd) == 0 ||
		len(aRev) == 0 || len(bRev) == 0 {

		log.Warnf("ignoring op %+v: no round trip", op)
		return nil
	}

	isFullyDelisted := func(lp defs.LocPair, entries []*TransEntry) bool {
		if len(entries) == 0 {
			log.Warnf("delisted to 0: %s,%s",
				LocationString(lp.From), LocationString(lp.To))
			return true
		}

		return false
	}

	aZIP := aFwd[0].FromBlock
	aDB := state.Blocks[aZIP.GameIdx][aZIP.BlockIdx]

	filtAFwd := delistEntries(aFwd, false)
	if isFullyDelisted(op.A, filtAFwd) {
		return nil
	}

	filtARev := delistEntries(aRev, true)
	if isFullyDelisted(op.A, filtARev) {
		return nil
	}

	bZIP := bRev[0].FromBlock
	bDB := state.Blocks[bZIP.GameIdx][bZIP.BlockIdx]

	filtBFwd := delistEntries(bFwd, true)
	if isFullyDelisted(op.B, filtBFwd) {
		return nil
	}

	filtBRev := delistEntries(bRev, false)
	if isFullyDelisted(op.B, filtBRev) {
		return nil
	}

	filtBRev1WayUp := delistEntries(coll.Get1WayUp(op.B.To), false)
	return &transOpCtxt{
		AFwd: filtAFwd,
		ARev: filtARev,
		ADB:  aDB,

		BFwd: filtBFwd,
		BRev: filtBRev,
		BDB:  bDB,

		BRev1WayUp: filtBRev1WayUp,
	}
}

// ExecTransOp modifies a pair of transitions according to the specified
// TransOp.  It modifies a TransOp's first two transitions as follows:
//
// 1. A.From --> A.to   BECOMES   A.From --> B.to
// 2. A.From <-- A.to   BECOMES   A.From <-- B.to
func ExecTransOp(coll *Collection, state *decode.DecodeState, op TransOp) {
	toe := newTransOpCtxt(coll, state, op)
	if toe == nil {
		return
	}

	// For example:
	// We are replacing the highpool->workshop transition with agcenter->cave.
	// A: highpool->workshop
	// B: agcenter->cave

	locStr := func(loc int) string {
		return fmt.Sprintf("%-3d %s", loc, LocationString(loc))
	}
	lpStr := func(lp defs.LocPair) string {
		s0 := fmt.Sprintf("[%s],", locStr(lp.From))
		s1 := fmt.Sprintf("[%s]", locStr(lp.To))
		return fmt.Sprintf("%-30s %-30s", s0, s1)
	}

	// Cycles through "identical" selectors to keep things interesting.
	selectCopySrc := func(idx int, entries []*TransEntry) (int, action.Transition) {
		idx = idx % len(entries)
		entry := entries[idx]

		return entry.Selector, entry.Trans
	}

	log.Infof("setting transition: %s <-- %s", lpStr(op.A), lpStr(op.B))

	// Replace highpool->workshop with agcenter->cave.
	for i, e := range toe.AFwd {
		srcSel, srcTrans := selectCopySrc(i, toe.BFwd)
		log.Debugf("replacing %s->%s(%d) with %s->%s(%d) (forward route)",
			locStr(op.B.From), locStr(op.B.To), e.Selector,
			locStr(op.A.From), locStr(op.A.To), srcSel)
		CopyTrans(toe.ADB.ActionTables.Transitions[e.Selector],
			srcTrans)
	}

	// Replace cave->agcenter with workshop->highpool.
	for i, e := range toe.BRev {
		srcSel, srcTrans := selectCopySrc(i, toe.ARev)
		log.Debugf("replacing %s->%s(%d) with %s->%s(%d) (reverse route)",
			locStr(op.B.To), locStr(op.B.From), e.Selector,
			locStr(op.A.To), locStr(op.A.From), srcSel)
		CopyTrans(toe.BDB.ActionTables.Transitions[e.Selector],
			srcTrans)
	}

	// Replace cave->worldmap with workshop->highpool.  The player should not
	// emerge in a completely different part of the world map, so we reroute
	// this transition to send the player the way he came.
	for i, e := range toe.BRev1WayUp {
		srcSel, srcTrans := selectCopySrc(i, toe.ARev)
		log.Debugf("replacing %s->%s(%d) with %s->%s(%d) (one way up)",
			locStr(op.B.To), locStr(e.Trans.Location), e.Selector,
			locStr(op.A.To), locStr(op.A.From), srcSel)
		CopyTrans(toe.BDB.ActionTables.Transitions[e.Selector],
			srcTrans)
	}
}
