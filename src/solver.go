package main

import (
    "errors"
)

var DIRECTIONS = []struct{ Name string; DR, DC int }{
    {"U", -1, 0}, {"D", 1, 0}, {"L", 0, -1}, {"R", 0, 1},
}

type SolveResult struct{
    Found bool
    Moves string
    TotalCost int
    Expanded int
    ExploredOrder []State
    StatePath []State
}

func initialState(b *Board) State{ return State{Pos:b.Start, NextDigit:0} }

func isGoal(b *Board, s State) bool{ return s.Pos==b.Goal && s.NextDigit > b.MaxDigit() }

func simulateSlide(b *Board, s State, dr, dc int) (*State, int){
    r,c := s.Pos.R, s.Pos.C
    rr,cc := r,c
    totalCost := 0
    nextDigit := s.NextDigit
    moved := false
    for {
        nr := rr + dr
        nc := cc + dc
        if !b.InBounds(nr,nc) { return nil, 0 }
        nxt := b.Tile(nr,nc)
        if nxt=='X' { break }
        moved = true
        if nxt=='L' { return nil, 0 }
        if nxt>='0' && nxt<='9' {
            d := int(nxt - '0')
            if d > nextDigit { return nil, 0 }
            if d == nextDigit { nextDigit++ }
        }
        totalCost += b.TraversalCost(nr,nc)
        rr,cc = nr,nc
    }
    if !moved { return nil, 0 }
    ns := State{Pos:Position{rr,cc}, NextDigit:nextDigit}
    return &ns, totalCost
}

func priority(algorithm string, g int, h float64, w float64) float64{
    switch algorithm {
    case "UCS": return float64(g)
    case "GBFS": return h
    case "A*": return float64(g) + h
    case "WASTAR": return float64(g) + w*h
    }
    panic("unknown algorithm")
}

func Solve(b *Board, algorithm, heuristicName string, weighted float64) (SolveResult, error){
    if algorithm!="UCS" && algorithm!="GBFS" && algorithm!="A*" && algorithm!="WASTAR"{
        return SolveResult{}, errors.New("Algoritma harus UCS, GBFS, A*, atau WASTAR.")
    }
    hfunc, ok := Heuristics[heuristicName]
    if !ok { return SolveResult{}, errors.New("Heuristik tidak dikenal: "+heuristicName) }

    initial := initialState(b)
    frontier := &PriorityQueue{}
    serial := 0
    gbest := map[State]int{ initial:0 }
    parent := map[State]*State{ initial:nil }
    parentMove := map[State]string{}
    exploredOrder := []State{}

    firstH := hfunc(b, initial)
    frontier.PushItem(PQItem{Priority:priority(algorithm,0,firstH,weighted), Serial:serial, State:initial})
    serial++

    expanded := 0
    var goalState *State

    for frontier.Len()>0 {
        it := frontier.PopItem()
        current := it.State
        currentG := gbest[current]
        expanded++
        exploredOrder = append(exploredOrder, current)
        if isGoal(b, current){ goalState = &current; break }
        for _, dir := range DIRECTIONS{
            nxt, stepCost := simulateSlide(b, current, dir.DR, dir.DC)
            if nxt==nil { continue }
            nxtG := currentG + stepCost
            old, ok := gbest[*nxt]
            if ok && nxtG >= old { continue }
            gbest[*nxt] = nxtG
            // store copy
            ncopy := *nxt
            parent[ncopy] = &current
            parentMove[ncopy] = dir.Name
            hval := hfunc(b, ncopy)
            pri := priority(algorithm, nxtG, hval, weighted)
            frontier.PushItem(PQItem{Priority:pri, Serial:serial, State:ncopy})
            serial++
        }
    }

    if goalState==nil{
        return SolveResult{Found:false, Moves:"", TotalCost:0, Expanded:expanded, ExploredOrder:exploredOrder, StatePath:[]State{}}, nil
    }

    // reconstruct path
    movesRev := []string{}
    pathRev := []State{}
    cur := *goalState
    for {
        pathRev = append(pathRev, cur)
        prev := parent[cur]
        if prev!=nil {
            movesRev = append(movesRev, parentMove[cur])
        }
        if prev==nil { break }
        cur = *prev
    }
    // reverse
    path := make([]State, len(pathRev))
    for i:=0;i<len(pathRev);i++{ path[i]=pathRev[len(pathRev)-1-i] }
    moves := ""
    for i:=len(movesRev)-1;i>=0;i--{ moves += movesRev[i] }

    return SolveResult{Found:true, Moves:moves, TotalCost: gbest[*goalState], Expanded:expanded, ExploredOrder:exploredOrder, StatePath:path}, nil
}
