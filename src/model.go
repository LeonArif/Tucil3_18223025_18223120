package main

type Position struct{
	R int
	C int
}

type State struct{
	Pos Position
	NextDigit int
}

type Board struct{
	Rows int
	Cols int
	Grid [][]rune
	Costs [][]int
	Start Position
	Goal Position
	DigitPositions map[int]Position
}

func (b *Board) MaxDigit() int{
	if len(b.DigitPositions)==0 { return -1 }
	m := -1
	for k := range b.DigitPositions{ if k>m { m=k } }
	return m
}

func (b *Board) InBounds(r,c int) bool{ return r>=0 && r<b.Rows && c>=0 && c<b.Cols }

func (b *Board) Tile(r,c int) rune{ return b.Grid[r][c] }

func (b *Board) TraversalCost(r,c int) int{ return b.Costs[r][c] }

func RenderBoard(b *Board, actor Position, nextDigit int) []string{
	out := make([]string, 0, b.Rows)
	for r:=0; r<b.Rows; r++{
		row := make([]rune, b.Cols)
		for c:=0; c<b.Cols; c++{
			if r==actor.R && c==actor.C{ row[c]='Z'; continue }
			ch := b.Grid[r][c]
			if ch>='0' && ch<='9' {
				d := int(ch - '0')
				if d < nextDigit { row[c]='*'; continue }
			}
			if ch=='Z' { row[c]='*' } else { row[c]=ch }
		}
		out = append(out, string(row))
	}
	return out
}
