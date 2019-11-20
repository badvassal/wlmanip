package wlmanip

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/badvassal/wllib/decode"
	"github.com/badvassal/wllib/defs"
	"github.com/badvassal/wllib/gen"
	"github.com/badvassal/wllib/gen/wlerr"
)

// LocationString produces a user-friendly string for the given location code.
// It accepts an exact location (i.e., either a regular location or a sub
// location).
func LocationString(loc int) string {
	s := SubLocationNameMap[loc]
	if s != "" {
		return s
	}

	return defs.LocationString(loc)
}

// ParseLocation converts a string to an exact location code.
func ParseLocation(s string) (int, error) {
	for k, v := range SubLocationNameMap {
		if s == v {
			return k, nil
		}
	}

	return defs.ParseLocation(s)
}

func ParseLocationNoCase(s string) (int, error) {
	for k, v := range SubLocationNameMap {
		if strings.EqualFold(s, v) {
			return k, nil
		}
	}

	return defs.ParseLocationNoCase(s)
}

// LocationFullString is an embellished form of LocationString.
func LocationFullString(loc int) string {
	return fmt.Sprintf("%d (%s)", loc, LocationString(loc))
}

// TransitionIsIntra indicates whether a transition is marked as "intra" (i.e.,
// within the same general area).
func TransitionIsIntra(lp defs.LocPair) bool {
	for _, entry := range IntraTransitions {
		if lp.From == entry.From && lp.To == entry.To {
			return true
		}
	}

	return false
}

// FixupTransitions converts some relative transitions to absolute.  The
// transitions that it modifies are specified in a hardcoded list.
func FixupTransitions(state *decode.DecodeState) error {
	relToAbs := func(gameIdx int, blockIdx int, selector int,
		baseCoords gen.Point) error {

		t := state.Blocks[gameIdx][blockIdx].ActionTables.Transitions[selector]
		if !t.Relative {
			return wlerr.Errorf("failed to convert transition to absolute: "+
				"game=%d block=%d selector=%d: transition not relative",
				gameIdx, blockIdx, selector)
		}
		oldT := *t

		t.MakeAbsolute(baseCoords)

		log.Debugf("converted relative transition to absolute: "+
			"game=%d block=%d selector=%d %+v --> %+v",
			gameIdx, blockIdx, selector, oldT, t)

		return nil
	}

	//// Make some relative transitions absolute.

	// Needles --> Downtown East.
	err := relToAbs(0, defs.Block0Needles, 11, gen.Point{12, 22})
	if err != nil {
		return err
	}
	err = relToAbs(0, defs.Block0Needles, 22, gen.Point{12, 22})
	if err != nil {
		return err
	}

	// Needles --> Downtown West.
	err = relToAbs(0, defs.Block0Needles, 20, gen.Point{15, 24})
	if err != nil {
		return err
	}

	// Downtown West --> Needles.
	err = relToAbs(0, defs.Block0NeedlesDowntownWest, 2, gen.Point{20, 30})
	if err != nil {
		return err
	}
	err = relToAbs(0, defs.Block0NeedlesDowntownWest, 5, gen.Point{20, 30})
	if err != nil {
		return err
	}

	return nil
}
