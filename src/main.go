package main

import (
    "bufio"
    "flag"
    "fmt"
    "os"
    "strings"
    "time"
)

func promptLine(prompt string) string{
    fmt.Print(prompt)
    rd := bufio.NewReader(os.Stdin)
    s, _ := rd.ReadString('\n')
    return strings.TrimSpace(s)
}

func saveSolution(path string, b *Board, res SolveResult, elapsed float64) error{
    f, err := os.Create(path)
    if err!=nil { return err }
    defer f.Close()
    w := bufio.NewWriter(f)
    if res.Found{
        fmt.Fprintf(w, "Solution: %s\n", res.Moves)
        fmt.Fprintf(w, "Cost: %d\n", res.TotalCost)
        fmt.Fprintf(w, "Iteration: %d\n", res.Expanded)
        fmt.Fprintf(w, "Time(ms): %.3f\n\n", elapsed)
        for i, st := range res.StatePath{
            if i==0 { fmt.Fprintln(w, "Initial") } else { fmt.Fprintf(w, "Step %d: %s\n", i, string(res.Moves[i-1])) }
            for _, line := range RenderBoard(b, st.Pos, st.NextDigit){ fmt.Fprintln(w, line) }
            fmt.Fprintln(w, "")
        }
    } else {
        fmt.Fprintln(w, "No valid solution found.")
    }
    w.Flush()
    return nil
}

func runCLI(){
    fmt.Println("Ice Sliding Puzzle Solver (Go)")
    filePath := promptLine(">> Enter input file (.txt): ")
    if filePath=="" { fmt.Println("No input provided."); return }
    board, err := LoadBoard(filePath)
    if err!=nil { fmt.Println("Invalid input:", err); return }

    algorithm := strings.ToUpper(promptLine(">> Choose algorithm (UCS/GBFS/A*/WASTAR): "))
    heuristic := "H1"
    if algorithm=="GBFS" || algorithm=="A*" || algorithm=="WASTAR"{
        heuristic = strings.ToUpper(promptLine(">> Choose heuristic (H1/H2/H3/H4/H5): "))
    }
    start := time.Now()
    res, err := Solve(board, algorithm, heuristic, 1.8)
    if err!=nil { fmt.Println("Invalid configuration:", err); return }
    elapsed := float64(time.Since(start).Milliseconds())

    if res.Found{
        fmt.Println("\nSolution found")
        fmt.Println("Moves:", res.Moves)
        fmt.Println("Solution cost:", res.TotalCost)
    } else { fmt.Println("\nNo valid solution found.") }

    fmt.Printf("\nExecution time: %.3f ms\n", elapsed)
    fmt.Printf("Number of iterations explored: %d\n", res.Expanded)

    pb := strings.ToLower(promptLine(">> Run playback? (Yes/No): "))
    if (pb=="yes"||pb=="y") && res.Found{
        for i, st := range res.StatePath{
            fmt.Printf("\nStep %d\n", i)
            for _, l := range RenderBoard(board, st.Pos, st.NextDigit){ fmt.Println(l) }
            if i < len(res.StatePath)-1 { _ = promptLine("Press Enter for next step...") }
        }
    }

    save := strings.ToLower(promptLine(">> Save solution to file? (Yes/No): "))
    if save=="yes"||save=="y"{
        out := promptLine(">> Output file location (e.g., test/solution.txt): ")
        if out=="" { out = "test/solution.txt" }
        if err := saveSolution(out, board, res, elapsed); err!=nil { fmt.Println("Failed to save file:", err) } else { fmt.Println("Solution saved to:", out) }
    }
}

func main(){
    cliMode := flag.Bool("cli", false, "launch CLI")
    flag.Parse()
    if *cliMode {
        runCLI()
        return
    }
    runGUI()
}
