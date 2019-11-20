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
func delistEntries(entries []*TransEntry, isFrom bool, isRead bool) []*TransEntry {
	if len(entries) == 0 {
		return nil
	}

	var filtered []*TransEntry
	for _, e := range entries {
		keep := true

		var loc int
		if isFrom {
			loc = e.FromLoc
		} else {
			loc = e.Trans.Location
		}

		pair := LocationXListPairMap[defs.LocPair{e.FromLoc, e.Trans.Location}]
		var xlist TransXList
		if isRead {
			xlist = pair.Read
		} else {
			xlist = pair.Write
		}

		entryStr := fmt.Sprintf("%s,%d,", LocationString(loc), e.Selector)
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
	FromA []*TransEntry
	ToA   []*TransEntry
	SrcDB decode.Block

	FromB []*TransEntry
	ToB   []*TransEntry
	DstDB decode.Block

	ToA1WayUp []*TransEntry
}

func newTransOpCtxt(coll *Collection, state *decode.DecodeState, op TransOp) *transOpCtxt {
	fromA, toA := coll.GetFromTo(op.A)
	fromB, toB := coll.GetFromTo(op.B)

	if len(fromA) == 0 || len(fromB) == 0 ||
		len(toA) == 0 || len(toB) == 0 {

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

	dstZIP := fromB[0].FromBlock
	dstDB := state.Blocks[dstZIP.GameIdx][dstZIP.BlockIdx]

	filtFromA := delistEntries(fromA, true, true)
	if isFullyDelisted(op.A, filtFromA) {
		return nil
	}

	filtToA := delistEntries(toA, false, false)
	if isFullyDelisted(op.A, filtToA) {
		return nil
	}

	srcZIP := toA[0].FromBlock
	srcDB := state.Blocks[srcZIP.GameIdx][srcZIP.BlockIdx]

	filtFromB := delistEntries(fromB, true, false)
	if isFullyDelisted(op.B, filtFromB) {
		return nil
	}

	filtToB := delistEntries(toB, false, true)
	if isFullyDelisted(op.B, filtToB) {
		return nil
	}

	filtToA1WayUp := delistEntries(coll.Get1WayUp(op.A.To), false, false)
	if isFullyDelisted(op.A, filtToA) {
		return nil
	}

	return &transOpCtxt{
		FromA: filtFromA,
		ToA:   filtToA,
		SrcDB: srcDB,

		FromB: filtFromB,
		ToB:   filtToB,
		DstDB: dstDB,

		ToA1WayUp: filtToA1WayUp,
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
	// dst: highpool->workshop
	// src: agcenter->cave

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

	log.Infof("setting transition: %s <-- %s", lpStr(op.B), lpStr(op.A))

	// Replace highpool->workshop with agcenter->cave.
	for i, e := range toe.FromB {
		srcSel, srcTrans := selectCopySrc(i, toe.FromA)
		log.Debugf("replacing %s->%s(%d) with %s->%s(%d) (forward route)",
			locStr(op.B.From), locStr(op.B.To), e.Selector,
			locStr(op.A.From), locStr(op.A.To), srcSel)
		CopyTrans(toe.DstDB.ActionTables.Transitions[e.Selector],
			srcTrans)
	}

	// Replace cave->agcenter with workshop->highpool.
	for i, e := range toe.ToA {
		srcSel, srcTrans := selectCopySrc(i, toe.ToB)
		log.Debugf("replacing %s->%s(%d) with %s->%s(%d) (reverse route)",
			locStr(op.A.To), locStr(op.A.From), e.Selector,
			locStr(op.B.To), locStr(op.B.From), srcSel)
		CopyTrans(toe.SrcDB.ActionTables.Transitions[e.Selector],
			srcTrans)
	}

	// Replace cave->worldmap with workshop->highpool.  The player should not
	// emerge in a completely different part of the world map, so we reroute
	// this transition to send the player the way he came.
	for i, e := range toe.ToA1WayUp {
		srcSel, srcTrans := selectCopySrc(i, toe.ToB)
		log.Debugf("replacing %s->%s(%d) with %s->%s(%d) (one way up)",
			locStr(op.A.To), locStr(e.Trans.Location), e.Selector,
			locStr(op.B.To), locStr(op.B.From), srcSel)
		CopyTrans(toe.SrcDB.ActionTables.Transitions[e.Selector],
			srcTrans)
	}
}
