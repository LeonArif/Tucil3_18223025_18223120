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
        fmt.Fprintf(w, "Solusi: %s\n", res.Moves)
        fmt.Fprintf(w, "Cost: %d\n", res.TotalCost)
        fmt.Fprintf(w, "Iterasi: %d\n", res.Expanded)
        fmt.Fprintf(w, "Waktu(ms): %.3f\n\n", elapsed)
        for i, st := range res.StatePath{
            if i==0 { fmt.Fprintln(w, "Initial") } else { fmt.Fprintf(w, "Step %d: %s\n", i, string(res.Moves[i-1])) }
            for _, line := range RenderBoard(b, st.Pos, st.NextDigit){ fmt.Fprintln(w, line) }
            fmt.Fprintln(w, "")
        }
    } else {
        fmt.Fprintln(w, "Tidak ada solusi valid.")
    }
    w.Flush()
    return nil
}

func runCLI(){
    fmt.Println("Ice Sliding Puzzle Solver (Go)")
    filePath := promptLine(">> Masukkan file input (.txt): ")
    if filePath=="" { fmt.Println("Tidak ada input."); return }
    board, err := LoadBoard(filePath)
    if err!=nil { fmt.Println("Input tidak valid:", err); return }

    algorithm := strings.ToUpper(promptLine(">> Pilih algoritma (UCS/GBFS/A*/WASTAR): "))
    heuristic := "H1"
    if algorithm=="GBFS" || algorithm=="A*" || algorithm=="WASTAR"{
        heuristic = strings.ToUpper(promptLine(">> Pilih heuristic (H1/H2/H3/H4/H5): "))
    }
    start := time.Now()
    res, err := Solve(board, algorithm, heuristic, 1.8)
    if err!=nil { fmt.Println("Konfigurasi tidak valid:", err); return }
    elapsed := float64(time.Since(start).Milliseconds())

    if res.Found{
        fmt.Println("\nSolusi ditemukan")
        fmt.Println("Gerakan:", res.Moves)
        fmt.Println("Cost solusi:", res.TotalCost)
    } else { fmt.Println("\nTidak ada solusi valid.") }

    fmt.Printf("\nWaktu eksekusi: %.3f ms\n", elapsed)
    fmt.Printf("Banyak iterasi ditinjau: %d\n", res.Expanded)

    pb := strings.ToLower(promptLine(">> Jalankan playback? (Ya/Tidak): "))
    if (pb=="ya"||pb=="y"||pb=="yes") && res.Found{
        for i, st := range res.StatePath{
            fmt.Printf("\nStep %d\n", i)
            for _, l := range RenderBoard(board, st.Pos, st.NextDigit){ fmt.Println(l) }
            if i < len(res.StatePath)-1 { _ = promptLine("Tekan Enter untuk langkah berikutnya...") }
        }
    }

    save := strings.ToLower(promptLine(">> Simpan solusi ke file? (Ya/Tidak): "))
    if save=="ya"||save=="y"||save=="yes"{
        out := promptLine(">> Lokasi file output (mis. test/solusi.txt): ")
        if out=="" { out = "test/solusi.txt" }
        if err := saveSolution(out, board, res, elapsed); err!=nil { fmt.Println("Gagal menyimpan file:", err) } else { fmt.Println("Solusi disimpan ke:", out) }
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
