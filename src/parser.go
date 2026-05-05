package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)

type ParseError string

func (e ParseError) Error() string { return string(e) }

func readNonEmptyLines(path string) ([]string, error){
    f, err := os.Open(path)
    if err!=nil { return nil, err }
    defer f.Close()
    return readNonEmptyLinesFromScanner(bufio.NewScanner(f))
}

func readNonEmptyLinesFromText(text string) ([]string, error){
    return readNonEmptyLinesFromScanner(bufio.NewScanner(strings.NewReader(text)))
}

func readNonEmptyLinesFromScanner(s *bufio.Scanner) ([]string, error){
    lines := []string{}
    for s.Scan(){
        t := strings.TrimRight(s.Text(), "\n\r")
        if strings.TrimSpace(t)=="" { continue }
        lines = append(lines, t)
    }
    return lines, s.Err()
}

func LoadBoardFromText(text string) (*Board, error){
    lines, err := readNonEmptyLinesFromText(text)
    if err!=nil { return nil, err }
    return parseBoardLines(lines)
}

func LoadBoard(path string) (*Board, error){
    lines, err := readNonEmptyLines(path)
    if err!=nil { return nil, err }
    return parseBoardLines(lines)
}

func parseBoardLines(lines []string) (*Board, error){
    if len(lines) < 3 { return nil, ParseError("Input terlalu pendek.") }

    header := strings.Fields(lines[0])
    if len(header)!=2 { return nil, ParseError("Baris pertama harus berisi dua bilangan: N M.") }
    n, err := strconv.Atoi(header[0])
    if err!=nil { return nil, ParseError("N dan M harus berupa bilangan bulat.") }
    m, err := strconv.Atoi(header[1])
    if err!=nil { return nil, ParseError("N dan M harus berupa bilangan bulat.") }
    if n<=0 || m<=0 { return nil, ParseError("Ukuran papan harus positif.") }

    expected := 1 + n + n
    if len(lines)!=expected { return nil, ParseError(fmt.Sprintf("Jumlah baris tidak sesuai. Ditemukan %d, seharusnya %d.", len(lines), expected)) }

    gridLines := lines[1:1+n]
    costLines := lines[1+n:1+n+n]

    allowed := map[rune]bool{}
    for _, ch := range "*XLOZ0123456789" { allowed[ch]=true }

    var startPos Position
    var goalPos Position
    gotStart := false
    gotGoal := false
    digitPositions := map[int]Position{}
    grid := make([][]rune, n)
    for r, line := range gridLines{
        if len(line) != m { return nil, ParseError(fmt.Sprintf("Panjang baris papan ke-%d tidak sama dengan M=%d.", r+1, m)) }
        row := []rune(line)
        for c, ch := range row{
            if !allowed[ch] { return nil, ParseError(fmt.Sprintf("Karakter tidak valid '%c' di (%d, %d).", ch, r, c)) }
            if ch=='Z' {
                if gotStart { return nil, ParseError("Aktor Z harus tepat satu.") }
                startPos = Position{r,c}; gotStart = true
            } else if ch=='O' {
                if gotGoal { return nil, ParseError("Tujuan O harus tepat satu.") }
                goalPos = Position{r,c}; gotGoal = true
            } else if ch>='0' && ch<='9' {
                d := int(ch - '0')
                if _,ok:=digitPositions[d]; ok { return nil, ParseError(fmt.Sprintf("Digit %d muncul lebih dari sekali.", d)) }
                digitPositions[d] = Position{r,c}
            }
        }
        grid[r] = row
    }
    if !gotStart { return nil, ParseError("Aktor Z tidak ditemukan.") }
    if !gotGoal { return nil, ParseError("Tujuan O tidak ditemukan.") }
    if len(digitPositions)>0{
        maxd := -1
        keys := map[int]bool{}
        for k := range digitPositions{ if k>maxd { maxd=k }; keys[k]=true }
        for i:=0;i<=maxd;i++{ if !keys[i]{ return nil, ParseError("Digit harus berurutan mulai dari 0 tanpa celah.") } }
    }

    costs := make([][]int, n)
    for r, line := range costLines{
        parts := strings.Fields(line)
        if len(parts)!=m { return nil, ParseError(fmt.Sprintf("Jumlah cost pada baris cost ke-%d tidak sama dengan M=%d.", r+1, m)) }
        row := make([]int, m)
        for i, p := range parts{
            v, err := strconv.Atoi(p)
            if err!=nil { return nil, ParseError(fmt.Sprintf("Cost tidak valid pada baris cost ke-%d.", r+1)) }
            if v<0 { return nil, ParseError("Cost tidak boleh negatif.") }
            row[i]=v
        }
        costs[r]=row
    }

    return &Board{
        Rows:n, Cols:m, Grid:grid, Costs:costs, Start:startPos, Goal:goalPos, DigitPositions:digitPositions,
    }, nil
}
