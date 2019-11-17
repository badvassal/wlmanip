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

// CopyTrans replaces a destination transition with a source.  A few fields in
// the destintion are preserved to maintain data integrity in the parent MSQ
// block.
func CopyTrans(dst *action.Transition, src action.Transition) {
	dst.Relative = src.Relative
	dst.Prompt = src.Prompt
	dst.LocX = src.LocX
	dst.LocY = src.LocY
	dst.Derelict = src.Derelict
	dst.Location = src.Location
}

// filterSrcEntries applies white lists and black lists to a list of entries.
func filterSrcEntries(entries []*TransEntry) []*TransEntry {
	if len(entries) == 0 {
		return nil
	}

	var filtered []*TransEntry
	for _, e := range entries {
		keep := true

		xlist := LocationReadXListMap[defs.LocPair{e.FromLoc, e.Trans.Location}]
		if len(xlist.White) > 0 {
			keep = false
			for _, w := range xlist.White {
				if e.Selector == w {
					log.Debugf("whitelisting %s,%d",
						LocationString(e.FromLoc), e.Selector)
					keep = true
					break
				}
			}
		}

		if !keep {
			log.Debugf("delisting %s,%d",
				LocationString(e.FromLoc), e.Selector)
		} else {
			for _, b := range xlist.Black {
				if e.Selector == b {
					log.Debugf("blacklisting %s,%d",
						LocationString(e.FromLoc), e.Selector)
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
}

func newTransOpCtxt(coll *Collection, state *decode.DecodeState, op TransOp) *transOpCtxt {
	fromA, toA := coll.GetFromTo(op.A)
	fromB, toB := coll.GetFromTo(op.B)

	if len(fromA) == 0 || len(fromB) == 0 ||
		len(toA) == 0 || len(toB) == 0 {

		log.Warnf("ignoring op %+v: no round trip", op)
		return nil
	}

	dstZIP := fromB[0].FromBlock
	dstDB := state.Blocks[dstZIP.GameIdx][dstZIP.BlockIdx]
	filtFromA := filterSrcEntries(fromA)

	if len(filtFromA) == 0 {
		log.Warnf("filtered to 0: %s,%s",
			LocationString(op.A.From), LocationString(op.A.To))
		return nil
	}

	srcZIP := toA[0].FromBlock
	srcDB := state.Blocks[srcZIP.GameIdx][srcZIP.BlockIdx]
	filtToB := filterSrcEntries(toB)

	if len(filtToB) == 0 {
		log.Warnf("filtered to 0: %s,%s",
			LocationString(op.B.From), LocationString(op.B.To))
		return nil
	}

	return &transOpCtxt{
		FromA: filtFromA,
		ToA:   toA,
		SrcDB: srcDB,

		FromB: fromB,
		ToB:   filtToB,
		DstDB: dstDB,
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
	selectCopySrc := func(idx int, entries []*TransEntry) action.Transition {
		return entries[idx%len(entries)].Trans
	}

	log.Infof("setting transition: %s <-- %s", lpStr(op.B), lpStr(op.A))

	// Replace highpool->workshop with agcenter->cave.
	for i, e := range toe.FromB {
		log.Debugf("replacing %s->%s(%d) with %s->%s (forward route)",
			locStr(op.B.From), locStr(op.B.To), e.Selector,
			locStr(op.A.From), locStr(op.A.To))
		CopyTrans(toe.DstDB.ActionTables.Transitions[e.Selector],
			selectCopySrc(i, toe.FromA))
	}

	// Replace cave->agcenter with workshop->highpool.
	for i, e := range toe.ToA {
		log.Debugf("replacing %s->%s(%d) with %s->%s (reverse route)",
			locStr(op.A.To), locStr(op.A.From), e.Selector,
			locStr(op.B.To), locStr(op.B.From))
		CopyTrans(toe.SrcDB.ActionTables.Transitions[e.Selector],
			selectCopySrc(i, toe.ToB))
	}

	// Replace cave->worldmap with workshop->highpool.  The player should not
	// emerge in a completely different part of the world map, so we reroute
	// this transition to send the player the way he came.
	for i, e := range coll.Get1WayUp(op.A.To) {
		log.Debugf("replacing %s->%s(%d) with %s->%s (one way up)",
			locStr(op.A.To), locStr(e.Trans.Location), e.Selector,
			locStr(op.B.To), locStr(op.B.From))
		CopyTrans(toe.SrcDB.ActionTables.Transitions[e.Selector],
			selectCopySrc(i, toe.ToB))
	}
}
