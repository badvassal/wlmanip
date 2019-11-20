package wlmanip

import (
	"github.com/badvassal/wllib/defs"
)

// Sub locations are areas within another map that should be considered
// separate.
const (
	SubLocationMin = 256

	SubLocationHighpoolCave            = 256
	SubLocationHighpoolCommunityCenter = 257
	SubLocationHighpoolWorkshop        = 258

	SubLocationAgCenterRootCellar = 259

	SubLocationDesertNomadsTent = 260

	SubLocationUglysHideoutAlley = 261

	SubLocationNeedlesBishopsOffice = 262
	SubLocationNeedlesGarage        = 263
	SubLocationNeedlesPoliceStation = 264
	SubLocationNeedlesAmmoBunker    = 265

	SubLocationDarwinBlackMarket     = 266
	SubLocationDarwinLab             = 267
	SubLocationDarwinBlackGilaTavern = 268

	SubLocationLasVegasJail         = 269
	SubLocationLasVegasProtonAxRoom = 270

	SubLocationSpadesCasinoWineCellar = 271
	SubLocationSpadesCasinoLevel2     = 272
	SubLocationSpadesCasinoBasement   = 273
)

var SubLocationNameMap = map[int]string{
	SubLocationHighpoolCave:            "HighpoolCave",
	SubLocationHighpoolCommunityCenter: "HighpoolCommunityCenter",
	SubLocationHighpoolWorkshop:        "HighpoolWorkshop",
	SubLocationAgCenterRootCellar:      "AgCenterRootCellar",
	SubLocationDesertNomadsTent:        "DesertNomadsTent",
	SubLocationUglysHideoutAlley:       "UglysHideoutAlley",
	SubLocationNeedlesBishopsOffice:    "NeedlesBishopsOffice",
	SubLocationNeedlesGarage:           "NeedlesGarage",
	SubLocationNeedlesPoliceStation:    "NeedlesPoliceStation",
	SubLocationNeedlesAmmoBunker:       "NeedlesAmmoBunker",
	SubLocationDarwinBlackMarket:       "DarwinBlackMarket",
	SubLocationDarwinLab:               "DarwinLab",
	SubLocationDarwinBlackGilaTavern:   "DarwinBlackGilaTavern",
	SubLocationLasVegasJail:            "LasVegasJail",
	SubLocationLasVegasProtonAxRoom:    "LasVegasProtonAxRoom",
	SubLocationSpadesCasinoWineCellar:  "SpadesCasinoWineCellar",
	SubLocationSpadesCasinoLevel2:      "SpadesCasinoLevel2",
	SubLocationSpadesCasinoBasement:    "SpadesCasinoBasement",
}

type SubLocDesc struct {
	GameIdx  int
	BlockIdx int
	Selector int
}

// SubLocMap indicates which transitions actually lead to or from a sub
// location (as opposed to a regular location).  A value of -1 represents the
// parent regular location.
var SubLocMap = map[SubLocDesc]defs.LocPair{
	// Highpool.
	SubLocDesc{0, defs.Block0Highpool, 1}: defs.LocPair{-1, SubLocationHighpoolCave},
	SubLocDesc{0, defs.Block0Highpool, 2}: defs.LocPair{SubLocationHighpoolCave, -1},
	SubLocDesc{0, defs.Block0Highpool, 3}: defs.LocPair{-1, SubLocationHighpoolCommunityCenter},
	SubLocDesc{0, defs.Block0Highpool, 4}: defs.LocPair{SubLocationHighpoolCommunityCenter, -1},
	SubLocDesc{0, defs.Block0Highpool, 5}: defs.LocPair{-1, SubLocationHighpoolWorkshop},
	SubLocDesc{0, defs.Block0Highpool, 6}: defs.LocPair{SubLocationHighpoolWorkshop, -1},

	// AgCenter.
	SubLocDesc{0, defs.Block0AgCenter, 3}:   defs.LocPair{-1, SubLocationAgCenterRootCellar},
	SubLocDesc{0, defs.Block0VerminCave, 0}: defs.LocPair{SubLocationAgCenterRootCellar, -1},

	// Desert Nomads.
	SubLocDesc{0, defs.Block0DesertNomads, 1}:  defs.LocPair{-1, SubLocationDesertNomadsTent},
	SubLocDesc{0, defs.Block0DesertNomads, 12}: defs.LocPair{-1, SubLocationDesertNomadsTent},
	SubLocDesc{0, defs.Block0DesertNomads, 13}: defs.LocPair{SubLocationDesertNomadsTent, -1},

	// Quartz
	SubLocDesc{0, defs.Block0Quartz, 92}:      defs.LocPair{-1, SubLocationUglysHideoutAlley},
	SubLocDesc{0, defs.Block0UglysHideout, 1}: defs.LocPair{SubLocationUglysHideoutAlley, -1},

	// Needles
	SubLocDesc{0, defs.Block0Needles, 8}:       defs.LocPair{-1, SubLocationNeedlesBishopsOffice},
	SubLocDesc{0, defs.Block0PoliceStation, 6}: defs.LocPair{SubLocationNeedlesBishopsOffice, -1},
	SubLocDesc{0, defs.Block0Needles, 9}:       defs.LocPair{-1, SubLocationNeedlesPoliceStation},
	SubLocDesc{0, defs.Block0PoliceStation, 3}: defs.LocPair{SubLocationNeedlesPoliceStation, -1},
	SubLocDesc{0, defs.Block0Needles, 10}:      defs.LocPair{-1, SubLocationNeedlesGarage},
	SubLocDesc{0, defs.Block0PoliceStation, 2}: defs.LocPair{SubLocationNeedlesGarage, -1},
	SubLocDesc{0, defs.Block0Needles, 19}:      defs.LocPair{-1, SubLocationNeedlesAmmoBunker},
	SubLocDesc{0, defs.Block0WastePit, 5}:      defs.LocPair{SubLocationNeedlesAmmoBunker, -1},

	// Darwin Village.
	SubLocDesc{1, defs.Block1Darwin, 3}:  defs.LocPair{-1, SubLocationDarwinBlackMarket},
	SubLocDesc{1, defs.Block1Darwin, 1}:  defs.LocPair{SubLocationDarwinBlackMarket, -1},
	SubLocDesc{1, defs.Block1Darwin, 10}: defs.LocPair{SubLocationDarwinBlackMarket, -1},
	SubLocDesc{1, defs.Block1Darwin, 11}: defs.LocPair{SubLocationDarwinBlackMarket, -1},
	SubLocDesc{1, defs.Block1Darwin, 4}:  defs.LocPair{-1, SubLocationDarwinLab},
	SubLocDesc{1, defs.Block1Darwin, 2}:  defs.LocPair{SubLocationDarwinLab, -1},
	SubLocDesc{1, defs.Block1Darwin, 5}:  defs.LocPair{-1, SubLocationDarwinBlackGilaTavern},
	SubLocDesc{1, defs.Block1Darwin, 6}:  defs.LocPair{-1, SubLocationDarwinBlackGilaTavern},
	SubLocDesc{1, defs.Block1Darwin, 0}:  defs.LocPair{SubLocationDarwinBlackGilaTavern, -1},
	SubLocDesc{1, defs.Block1Darwin, 9}:  defs.LocPair{SubLocationDarwinBlackGilaTavern, -1},

	// Las Vegas
	SubLocDesc{1, defs.Block1LasVegas, 1}:   defs.LocPair{-1, SubLocationLasVegasJail},
	SubLocDesc{1, defs.Block1FatFreddys, 6}: defs.LocPair{SubLocationLasVegasJail, defs.LocationLasVegas},
	SubLocDesc{1, defs.Block1LasVegas, 3}:   defs.LocPair{-1, SubLocationLasVegasProtonAxRoom},
	SubLocDesc{1, defs.Block1FatFreddys, 5}: defs.LocPair{SubLocationLasVegasProtonAxRoom,
		defs.LocationLasVegas},

	// Spades Casino
	SubLocDesc{1, defs.Block1SpadesCasino, 1}: defs.LocPair{-1, SubLocationSpadesCasinoWineCellar},
	SubLocDesc{1, defs.Block1SpadesCasino, 2}: defs.LocPair{SubLocationSpadesCasinoWineCellar, -1},
	SubLocDesc{1, defs.Block1SpadesCasino, 3}: defs.LocPair{-1, SubLocationSpadesCasinoLevel2},
	SubLocDesc{1, defs.Block1SpadesCasino, 4}: defs.LocPair{SubLocationSpadesCasinoLevel2, -1},
	SubLocDesc{1, defs.Block1SpadesCasino, 7}: defs.LocPair{-1, SubLocationSpadesCasinoBasement},
	// Normally this transition leads out to Las Vegas.  Make it lead back to
	// Spade's Casino to create a round trip.
	SubLocDesc{1, defs.Block1SpadesCasino, 5}: defs.LocPair{SubLocationSpadesCasinoBasement,
		defs.LocationSpadesCasino},
}

// IntraTransitions indicates which transitions should be considered as
// "intra".  That is, transitions which take the player from one section to
// another of a single area.
var IntraTransitions = []defs.LocPair{
	defs.LocPair{defs.LocationBloodTempleTop, defs.LocationBloodTempleBottom},
	defs.LocPair{defs.LocationBloodTempleBottom, defs.LocationBloodTempleTop},
	defs.LocPair{defs.LocationLasVegasSewersWest, defs.LocationLasVegasSewersEast},
	defs.LocPair{defs.LocationLasVegasSewersEast, defs.LocationLasVegasSewersWest},
	defs.LocPair{defs.LocationNeedlesDowntownWest, defs.LocationNeedlesDowntownEast},
	defs.LocPair{defs.LocationNeedlesDowntownEast, defs.LocationNeedlesDowntownWest},
}

// LocationDepthMap indicates the "depth" of each location in the game.  The
// world map has a depth of 0; everything deeper has a greater depth.  This is
// used when setting transition data.  It doesn't make sense to replace a
// shallow-to-deep transition with a deep-to-shallow one.  For example,
// Highpool-to-cave is a shallow-to-deep transition.  It would be confusing if
// this were replaced by courthouse-to-quartz.
var LocationDepthMap = map[int]int{
	defs.LocationWorldMap:                  0,
	defs.LocationQuartz:                    1,
	defs.LocationScottsBar:                 2,
	defs.LocationStageCoachInn:             2,
	defs.LocationUglysHideout:              2,
	defs.LocationQuartzDerelictBuildings:   2,
	defs.LocationCourthouse:                2,
	defs.LocationSleeperBaseLevel1:         1,
	defs.LocationDesertNomads:              1,
	defs.LocationAgCenter:                  1,
	defs.LocationHighpool:                  1,
	defs.LocationLasVegasDerelictBuildings: 2,
	defs.LocationLasVegas:                  1,
	defs.LocationSleeperBaseLevel2:         2,
	defs.LocationSleeperBaseLevel3:         2,
	defs.LocationBaseCochiseOutside:        2,
	defs.LocationBaseCochiseLevel1:         3,
	defs.LocationBaseCochiseLevel3:         3,
	defs.LocationBaseCochiseLevel2:         3,
	defs.LocationBaseCochiseLevel4:         3,
	defs.LocationDarwin:                    1,
	defs.LocationDarwinBase:                2,
	defs.LocationFinstersBrain:             2,
	defs.LocationLasVegasSewersWest:        3,
	defs.LocationLasVegasSewersEast:        3,
	defs.LocationNeedles:                   1,
	defs.LocationBloodTempleTop:            2,
	defs.LocationBloodTempleBottom:         2,
	defs.LocationVerminCave:                2,
	defs.LocationWastePit:                  2,
	defs.LocationNeedlesDowntownEast:       3,
	defs.LocationNeedlesDowntownWest:       3,
	defs.LocationPoliceStation:             2,
	defs.LocationGuardianCitadelEntrance:   2,
	defs.LocationGuardianCitadelOuter:      3,
	defs.LocationTempleMushroom:            2,
	defs.LocationFaranBrygos:               2,
	defs.LocationFatFreddys:                2,
	defs.LocationSpadesCasino:              2,
	defs.LocationGuardianCitadelInner:      3,
	defs.LocationMineShaft:                 1,
	defs.LocationSavageVillage:             1,

	SubLocationHighpoolCave:            2,
	SubLocationHighpoolCommunityCenter: 2,
	SubLocationHighpoolWorkshop:        2,
	SubLocationAgCenterRootCellar:      2,
	SubLocationDesertNomadsTent:        2,
	SubLocationUglysHideoutAlley:       2,
	SubLocationNeedlesBishopsOffice:    2,
	SubLocationNeedlesGarage:           2,
	SubLocationNeedlesPoliceStation:    2,
	SubLocationNeedlesAmmoBunker:       2,
	SubLocationDarwinBlackMarket:       2,
	SubLocationDarwinLab:               2,
	SubLocationDarwinBlackGilaTavern:   2,
	SubLocationLasVegasJail:            2,
	SubLocationLasVegasProtonAxRoom:    2,
	SubLocationSpadesCasinoWineCellar:  3,
	SubLocationSpadesCasinoLevel2:      3,
	SubLocationSpadesCasinoBasement:    3,
}

// LocationPostSewersMap indicates which locations the player is expected to
// explore only after completing the sewers.
var LocationPostSewersMap = map[int]bool{
	defs.LocationSleeperBaseLevel1:       true,
	defs.LocationSleeperBaseLevel2:       true,
	defs.LocationSleeperBaseLevel3:       true,
	defs.LocationBaseCochiseOutside:      true,
	defs.LocationBaseCochiseLevel1:       true,
	defs.LocationBaseCochiseLevel3:       true,
	defs.LocationBaseCochiseLevel2:       true,
	defs.LocationBaseCochiseLevel4:       true,
	defs.LocationFinstersBrain:           true,
	defs.LocationGuardianCitadelEntrance: true,
	defs.LocationGuardianCitadelOuter:    true,
	defs.LocationGuardianCitadelInner:    true,
}

// LocationReadXListMap contains blacklisted and whitelisted transitions.
var LocationXListPairMap = map[defs.LocPair]TransXListPair{
	defs.LocPair{defs.LocationDesertNomads, defs.LocationDesertNomads}: TransXListPair{
		Read: TransXList{
			// This transition puts the player back in the tent (only if the
			// player has already entered the tent once).  If the
			// DesertNomads->Tent transition gets replaced, entering the tent
			// should never actually transport the player to the tent
			// sublocation.
			Black: []int{12},
		},
	},

	defs.LocPair{defs.LocationNeedles, defs.LocationNeedlesDowntownWest}: TransXListPair{
		Read: TransXList{
			// There are several ways to enter DowntownWest from Needles that
			// can be confusing when they replace other transitions.
			// FixupTransitions() makes transition 0 user friendly.  That is
			// the only one we want to use.
			White: []int{20},
		},
	},
}
