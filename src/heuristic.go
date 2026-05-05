package main

import "math"

func manhattan(a,b Position) int{ return abs(a.R-b.R)+abs(a.C-b.C) }
func abs(x int) int{ if x<0 { return -x }; return x }

func targetForState(b *Board, s State) Position{
    if s.NextDigit <= b.MaxDigit(){
        return b.DigitPositions[s.NextDigit]
    }
    return b.Goal
}

func H1(b *Board, s State) float64{
    t := targetForState(b,s)
    return float64(manhattan(s.Pos,t))
}

func H2(b *Board, s State) float64{
    t := targetForState(b,s)
    toTarget := manhattan(s.Pos,t)
    targetToGoal := manhattan(t,b.Goal)
    return float64(toTarget) + 0.5*float64(targetToGoal)
}

func H3(b *Board, s State) float64{
    t := targetForState(b,s)
    dr := abs(s.Pos.R - t.R)
    dc := abs(s.Pos.C - t.C)
    if dr>dc { return float64(dr) }
    return float64(dc)
}

func H4(b *Board, s State) float64{
    t := targetForState(b,s)
    dr := float64(s.Pos.R - t.R)
    dc := float64(s.Pos.C - t.C)
    return math.Sqrt(dr*dr + dc*dc)
}

func H5(b *Board, s State) float64{
    t := targetForState(b,s)
    base := manhattan(s.Pos,t)
    remaining := b.MaxDigit() - s.NextDigit + 1
    if remaining < 0 { remaining = 0 }
    return float64(base + 2*remaining)
}

type HeuristicFunc func(*Board, State) float64

var Heuristics = map[string]HeuristicFunc{
    "H1": H1,
    "H2": H2,
    "H3": H3,
    "H4": H4,
    "H5": H5,
}
