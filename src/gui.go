package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type guiSession struct {
	Board      *Board
	Result     SolveResult
	ElapsedMS  float64
	Algorithm  string
	Heuristic  string
	Weighted   float64
	InputLabel string
	Frames     [][]string
	FrameMoves []string
	Error      string
}

type guiServer struct {
	mu      sync.RWMutex
	session *guiSession
}

func runGUI() {
	server := &guiServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleIndex)
	mux.HandleFunc("/solve", server.handleSolve)
	mux.HandleFunc("/download", server.handleDownload)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("Gagal memulai server GUI:", err)
		return
	}

	baseURL := "http://" + listener.Addr().String()
	go func() {
		_ = http.Serve(listener, mux)
	}()

	if err := openBrowser(baseURL); err != nil {
		fmt.Println("Server GUI aktif di", baseURL)
		fmt.Println("Gagal membuka browser otomatis:", err)
	} else {
		fmt.Println("GUI aktif di", baseURL)
	}

	select {}
}

func (s *guiServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(guiHTML))
}

func (s *guiServer) handleSolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		s.writeJSONError(w, http.StatusBadRequest, "Form tidak valid: "+err.Error())
		return
	}

	boardText := strings.TrimSpace(r.FormValue("boardText"))
	if boardText == "" {
		s.writeJSONError(w, http.StatusBadRequest, "Masukkan file papan terlebih dahulu.")
		return
	}

	board, err := LoadBoardFromText(boardText)
	if err != nil {
		s.writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	algorithm := strings.ToUpper(strings.TrimSpace(r.FormValue("algorithm")))
	if algorithm == "" {
		algorithm = "UCS"
	}
	heuristic := strings.ToUpper(strings.TrimSpace(r.FormValue("heuristic")))
	if heuristic == "" || heuristic == "-" {
		heuristic = "H1"
	}
	weighted := 1.8
	if rawWeighted := strings.TrimSpace(r.FormValue("weighted")); rawWeighted != "" {
		if parsed, parseErr := strconv.ParseFloat(rawWeighted, 64); parseErr == nil && parsed > 0 {
			weighted = parsed
		}
	}

	start := time.Now()
	result, solveErr := Solve(board, algorithm, heuristic, weighted)
	elapsed := float64(time.Since(start).Milliseconds())
	if solveErr != nil {
		s.writeJSONError(w, http.StatusBadRequest, solveErr.Error())
		return
	}

	frames := make([][]string, 0, len(result.StatePath))
	labels := make([]string, 0, len(result.StatePath))
	if len(result.StatePath) == 0 {
		frames = append(frames, RenderBoard(board, board.Start, 0))
		labels = append(labels, "Initial")
	} else {
		for i, state := range result.StatePath {
			frames = append(frames, RenderBoard(board, state.Pos, state.NextDigit))
			if i == 0 {
				labels = append(labels, "Initial")
			} else if i-1 < len(result.Moves) {
				labels = append(labels, fmt.Sprintf("Step %d: %c", i, result.Moves[i-1]))
			} else {
				labels = append(labels, fmt.Sprintf("Step %d", i))
			}
		}
	}

	session := &guiSession{
		Board:      board,
		Result:     result,
		ElapsedMS:  elapsed,
		Algorithm:  algorithm,
		Heuristic:  heuristic,
		Weighted:   weighted,
		InputLabel: "uploaded board",
		Frames:     frames,
		FrameMoves: labels,
	}

	s.mu.Lock()
	s.session = session
	s.mu.Unlock()

	response := map[string]any{
		"ok":         true,
		"found":      result.Found,
		"moves":      result.Moves,
		"totalCost":  result.TotalCost,
		"expanded":   result.Expanded,
		"elapsedMS":  elapsed,
		"algorithm":  algorithm,
		"heuristic":  heuristic,
		"weighted":   weighted,
		"frames":     frames,
		"frameMoves": labels,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(response)
}

func (s *guiServer) handleDownload(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	session := s.session
	s.mu.RUnlock()
	if session == nil {
		http.Error(w, "Belum ada hasil solusi.", http.StatusNotFound)
		return
	}

	filename := "solusi.txt"
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	if session.Result.Found {
		_, _ = fmt.Fprintf(w, "Algoritma: %s\n", session.Algorithm)
		_, _ = fmt.Fprintf(w, "Heuristik: %s\n", session.Heuristic)
		_, _ = fmt.Fprintf(w, "Weighted: %.3f\n", session.Weighted)
		_, _ = fmt.Fprintf(w, "Solusi: %s\n", session.Result.Moves)
		_, _ = fmt.Fprintf(w, "Cost: %d\n", session.Result.TotalCost)
		_, _ = fmt.Fprintf(w, "Iterasi: %d\n", session.Result.Expanded)
		_, _ = fmt.Fprintf(w, "Waktu(ms): %.3f\n\n", session.ElapsedMS)
		for i, st := range session.Result.StatePath {
			if i == 0 {
				_, _ = fmt.Fprintln(w, "Initial")
			} else {
				_, _ = fmt.Fprintf(w, "Step %d: %c\n", i, session.Result.Moves[i-1])
			}
			for _, line := range RenderBoard(session.Board, st.Pos, st.NextDigit) {
				_, _ = fmt.Fprintln(w, line)
			}
			_, _ = fmt.Fprintln(w)
		}
		return
	}
	_, _ = fmt.Fprintln(w, "Tidak ada solusi valid.")
}

func (s *guiServer) writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":    false,
		"error": message,
	})
}

func openBrowser(rawURL string) error {
	escaped := rawURL
	if runtime.GOOS == "windows" {
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", escaped).Start()
	}
	if runtime.GOOS == "darwin" {
		return exec.Command("open", escaped).Start()
	}
	return exec.Command("xdg-open", escaped).Start()
}

var guiHTML = `<!doctype html>
<html lang="en">
<head>
	<meta charset="utf-8" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<title>Ice Sliding Puzzle Solver</title>
	<style>
		:root {
			--bg: #07111f;
			--bg2: #0c1b2f;
			--card: rgba(9, 19, 35, 0.86);
			--card2: rgba(16, 31, 54, 0.9);
			--text: #eef4ff;
			--muted: #9fb3cf;
			--accent: #71d7ff;
			--accent2: #8a7dff;
			--ok: #56e39f;
			--warn: #ffbf69;
			--danger: #ff6b6b;
			--line: rgba(255,255,255,0.08);
			--shadow: 0 28px 80px rgba(0,0,0,0.35);
			--radius: 22px;
		}
		* { box-sizing: border-box; }
		body {
			margin: 0;
			min-height: 100vh;
			font-family: "Segoe UI", system-ui, sans-serif;
			color: var(--text);
			background:
				radial-gradient(circle at top left, rgba(113,215,255,0.16), transparent 28%),
				radial-gradient(circle at top right, rgba(138,125,255,0.18), transparent 25%),
				linear-gradient(160deg, var(--bg), var(--bg2));
		}
		.shell {
			width: min(1300px, calc(100vw - 32px));
			margin: 0 auto;
			padding: 24px 0 40px;
		}
		.hero {
			display: grid;
			gap: 18px;
			grid-template-columns: 1.4fr 0.9fr;
			align-items: end;
			margin-bottom: 18px;
		}
		.title {
			padding: 28px;
			border: 1px solid var(--line);
			border-radius: var(--radius);
			background: linear-gradient(135deg, rgba(15,30,54,0.92), rgba(10,18,33,0.88));
			box-shadow: var(--shadow);
		}
		.title h1 {
			margin: 0 0 10px;
			font-size: clamp(30px, 4vw, 54px);
			line-height: 0.96;
			letter-spacing: -0.05em;
		}
		.title p {
			margin: 0;
			color: var(--muted);
			max-width: 68ch;
		}
		.pillbar {
			display: flex;
			gap: 10px;
			flex-wrap: wrap;
			justify-content: flex-end;
		}
		.pill {
			border: 1px solid var(--line);
			background: rgba(255,255,255,0.05);
			color: var(--text);
			padding: 10px 14px;
			border-radius: 999px;
			font-size: 13px;
			backdrop-filter: blur(12px);
		}
		.grid {
			display: grid;
			grid-template-columns: 420px 1fr;
			gap: 18px;
		}
		.panel {
			border: 1px solid var(--line);
			background: var(--card);
			border-radius: var(--radius);
			box-shadow: var(--shadow);
			overflow: hidden;
		}
		.panel .head {
			padding: 18px 20px 14px;
			border-bottom: 1px solid var(--line);
			background: rgba(255,255,255,0.02);
		}
		.panel .head h2,
		.panel .head h3 {
			margin: 0;
			font-size: 16px;
			letter-spacing: 0.02em;
		}
		.panel .head .sub {
			margin-top: 6px;
			color: var(--muted);
			font-size: 13px;
		}
		.panel .body { padding: 18px 20px 20px; }
		.field {
			display: grid;
			gap: 8px;
			margin-bottom: 14px;
		}
		label { font-size: 13px; color: var(--muted); }
		input[type="text"], input[type="number"], select, textarea {
			width: 100%;
			border: 1px solid rgba(255,255,255,0.12);
			background: rgba(3, 10, 22, 0.75);
			color: var(--text);
			border-radius: 14px;
			padding: 12px 14px;
			outline: none;
			transition: border-color .18s ease, transform .18s ease, box-shadow .18s ease;
		}
		textarea {
			min-height: 260px;
			resize: vertical;
			font-family: Consolas, "Cascadia Mono", monospace;
			line-height: 1.35;
			white-space: pre;
		}
		input:focus, select:focus, textarea:focus {
			border-color: rgba(113,215,255,0.7);
			box-shadow: 0 0 0 3px rgba(113,215,255,0.15);
			transform: translateY(-1px);
		}
		.actions {
			display: flex;
			gap: 10px;
			flex-wrap: wrap;
			margin-top: 8px;
		}
		button, .linkbtn {
			border: 0;
			background: linear-gradient(135deg, var(--accent), var(--accent2));
			color: #06111e;
			font-weight: 700;
			padding: 12px 16px;
			border-radius: 14px;
			cursor: pointer;
			text-decoration: none;
			display: inline-flex;
			align-items: center;
			justify-content: center;
			transition: transform .18s ease, filter .18s ease;
		}
		button.secondary, .linkbtn.secondary {
			background: rgba(255,255,255,0.08);
			color: var(--text);
			border: 1px solid var(--line);
		}
		button:hover, .linkbtn:hover { transform: translateY(-1px); filter: brightness(1.05); }
		button:disabled { opacity: 0.55; cursor: not-allowed; transform: none; }
		.status {
			display: grid;
			grid-template-columns: repeat(4, minmax(0, 1fr));
			gap: 12px;
			margin-bottom: 14px;
		}
		.stat {
			padding: 16px;
			border-radius: 18px;
			background: rgba(255,255,255,0.04);
			border: 1px solid var(--line);
		}
		.stat .k { color: var(--muted); font-size: 12px; text-transform: uppercase; letter-spacing: .08em; }
		.stat .v { margin-top: 8px; font-size: 22px; font-weight: 700; }
		.stat .v.ok { color: var(--ok); }
		.stat .v.warn { color: var(--warn); }
		.workspace {
			display: grid;
			grid-template-columns: minmax(0, 1fr) 280px;
			gap: 16px;
			align-items: start;
		}
		.viewer {
			padding: 16px;
			border-radius: 18px;
			background: rgba(4, 12, 23, 0.92);
			border: 1px solid var(--line);
			min-height: 320px;
			overflow: hidden;
		}
		.board-shell {
			display: grid;
			gap: 12px;
		}
		.board-meta {
			display: flex;
			justify-content: space-between;
			align-items: center;
			gap: 12px;
			color: var(--muted);
			font-size: 13px;
			margin-bottom: 8px;
		}
		.board-grid {
			display: grid;
			gap: 4px;
			padding: 12px;
			border-radius: 18px;
			background:
				radial-gradient(circle at top, rgba(113,215,255,0.12), transparent 28%),
				linear-gradient(180deg, rgba(7, 18, 35, 0.98), rgba(12, 26, 46, 0.99));
			border: 1px solid rgba(255,255,255,0.07);
			box-shadow: inset 0 1px 0 rgba(255,255,255,0.03);
			position: relative;
			transform-origin: center;
			transition: transform .2s ease, filter .2s ease;
		}
		.board-grid.is-changing {
			transform: scale(0.992);
			filter: brightness(1.04);
		}
		.tile {
			position: relative;
			width: 42px;
			height: 42px;
			border-radius: 10px;
			display: flex;
			align-items: center;
			justify-content: center;
			font-weight: 800;
			font-size: 17px;
			line-height: 1;
			letter-spacing: 0.04em;
			user-select: none;
			transition: transform .22s cubic-bezier(.2,.9,.2,1), background-color .22s ease, box-shadow .22s ease, color .22s ease, opacity .22s ease;
		}
		.tile::after {
			content: '';
			position: absolute;
			inset: 0;
			border-radius: inherit;
			border: 1px solid rgba(255,255,255,0.03);
			pointer-events: none;
		}
		.tile.wall { background: linear-gradient(180deg, #1a2331, #0f1723); color: #60738d; }
		.tile.empty {
			background: linear-gradient(180deg, rgba(214,221,232,0.42), rgba(176,186,202,0.28));
			box-shadow: inset 0 1px 0 rgba(255,255,255,0.14), inset 0 -1px 0 rgba(0,0,0,0.18);
			color: transparent;
		}
		.tile.goal { background: linear-gradient(180deg, rgba(48,190,137,0.95), rgba(28,139,98,0.95)); color: #03140c; box-shadow: 0 0 0 1px rgba(86,227,159,0.18), 0 8px 18px rgba(86,227,159,0.18); }
		.tile.block { background: linear-gradient(180deg, rgba(173,100,255,0.9), rgba(103,61,194,0.9)); color: #fff; }
		.tile.digit { background: linear-gradient(180deg, rgba(255,191,105,0.95), rgba(214,136,36,0.95)); color: #2a1400; box-shadow: 0 0 0 1px rgba(255,191,105,0.22), 0 10px 22px rgba(255,191,105,0.14); }
		.tile.start { background: radial-gradient(circle at 30% 30%, #7ef0ff, #1293d0); color: #02121d; box-shadow: 0 0 0 1px rgba(113,215,255,0.22), 0 10px 24px rgba(19,147,208,0.25); transform: scale(1.03); }
		.tile.hazard { background: linear-gradient(180deg, rgba(255,110,110,0.92), rgba(183,57,57,0.92)); color: #fff3f3; }
		.tile.lit {
			animation: pulseGlow .7s ease-out;
		}
		.tile.moved {
			transform: translateY(-2px) scale(1.06);
		}
		.board-legend {
			display: flex;
			gap: 10px;
			flex-wrap: wrap;
			font-size: 12px;
			color: var(--muted);
		}
		.legend-item {
			display: inline-flex;
			align-items: center;
			gap: 8px;
			padding: 8px 10px;
			border-radius: 999px;
			background: rgba(255,255,255,0.04);
			border: 1px solid rgba(255,255,255,0.06);
		}
		.legend-swatch {
			width: 10px;
			height: 10px;
			border-radius: 999px;
		}
		@keyframes pulseGlow {
			0% { box-shadow: 0 0 0 0 rgba(113,215,255,0.0); }
			40% { box-shadow: 0 0 0 8px rgba(113,215,255,0.18); }
			100% { box-shadow: 0 0 0 0 rgba(113,215,255,0.0); }
		}
		.play {
			display: grid;
			gap: 12px;
			padding: 16px;
			border-radius: 18px;
			background: rgba(255,255,255,0.04);
			border: 1px solid var(--line);
		}
		.play .row {
			display: flex;
			gap: 10px;
			flex-wrap: wrap;
		}
		.slider {
			width: 100%;
		}
		.footer-note {
			color: var(--muted);
			font-size: 12px;
			margin-top: 12px;
		}
		.error {
			color: #ffd2d2;
			background: rgba(255, 107, 107, 0.12);
			border: 1px solid rgba(255, 107, 107, 0.25);
			padding: 12px 14px;
			border-radius: 14px;
			margin-top: 12px;
			display: none;
		}
		.muted { color: var(--muted); }
		.hidden { display: none !important; }
		@media (max-width: 1100px) {
			.grid, .hero, .workspace { grid-template-columns: 1fr; }
			.pillbar { justify-content: flex-start; }
			.status { grid-template-columns: repeat(2, minmax(0, 1fr)); }
		}
		@media (max-width: 640px) {
			.shell { width: calc(100vw - 20px); }
			.title, .panel .body, .panel .head { padding-left: 16px; padding-right: 16px; }
			.status { grid-template-columns: 1fr; }
			pre { font-size: 14px; }
		}
	</style>
</head>
<body>
	<div class="shell">
		<div class="hero">
			<div class="title">
				<h1>Ice Sliding Puzzle Solver</h1>
				<p>Upload a puzzle file, run UCS / GBFS / A* / WASTAR, then inspect the solved path with playback, statistics, and a downloadable report.</p>
			</div>
		</div>

		<div class="grid">
			<div class="panel">
				<div class="head">
					<h2>Input</h2>
					<div class="sub">Choose a file, or paste the raw board text below.</div>
				</div>
				<div class="body">
					<form id="solveForm">
						<div class="field">
							<label for="fileInput">Input file</label>
							<input id="fileInput" type="file" accept=".txt,text/plain" />
						</div>
						<div class="field">
							<label for="boardText">Board text</label>
							<textarea id="boardText" name="boardText" placeholder="Paste puzzle text here..."></textarea>
						</div>
						<div class="field">
							<label for="algorithm">Algorithm</label>
							<select id="algorithm" name="algorithm">
								<option value="UCS">UCS</option>
								<option value="GBFS">GBFS</option>
								<option value="A*">A*</option>
								<option value="WASTAR">WASTAR</option>
							</select>
						</div>
						<div class="field">
							<label for="heuristic">Heuristic</label>
							<select id="heuristic" name="heuristic">
								<option value="-">-</option>
								<option value="H1">H1</option>
								<option value="H2">H2</option>
								<option value="H3">H3</option>
								<option value="H4">H4</option>
								<option value="H5">H5</option>
							</select>
						</div>
						<div class="field">
							<label for="weighted">Weighted factor</label>
							<input id="weighted" name="weighted" type="number" min="1" step="0.1" value="1.8" />
						</div>
						<div class="actions">
							<button id="solveBtn" type="button">Run solver</button>
							<a class="linkbtn secondary" id="downloadBtn" href="/download" target="_blank">Download result</a>
						</div>
					</form>
					<div id="errorBox" class="error"></div>
					<div class="footer-note">Tip: selecting a file auto-fills the text box so you can edit before solving.</div>
				</div>
			</div>

			<div class="panel">
				<div class="head">
					<h2>Result</h2>
					<div class="sub">Solve output and step playback.</div>
				</div>
				<div class="body">
					<div class="status">
						<div class="stat"><div class="k">Found</div><div id="statFound" class="v warn">-</div></div>
						<div class="stat"><div class="k">Cost</div><div id="statCost" class="v">-</div></div>
						<div class="stat"><div class="k">Expanded</div><div id="statExpanded" class="v">-</div></div>
						<div class="stat"><div class="k">Elapsed</div><div id="statElapsed" class="v">-</div></div>
					</div>
					<div class="workspace">
						<div class="viewer">
							<div class="board-shell">
								<div class="board-meta">
									<div id="boardTitle">Load a puzzle to see the board here.</div>
									<div id="boardCoords" class="muted">0 x 0</div>
								</div>
								<div id="boardGrid" class="board-grid"></div>
								<div class="board-legend">
									<span class="legend-item"><span class="legend-swatch" style="background:#1293d0"></span>Actor</span>
									<span class="legend-item"><span class="legend-swatch" style="background:#30be89"></span>Goal</span>
									<span class="legend-item"><span class="legend-swatch" style="background:#d68824"></span>Digit</span>
									<span class="legend-item"><span class="legend-swatch" style="background:#b73939"></span>Hazard</span>
								</div>
							</div>
						</div>
						<div class="play">
							<div><strong id="stepLabel">Playback</strong></div>
							<input id="frameSlider" class="slider" type="range" min="0" max="0" value="0" disabled />
							<div class="row">
								<button id="prevBtn" class="secondary" type="button" disabled>Prev</button>
								<button id="playBtn" class="secondary" type="button" disabled>Play</button>
								<button id="nextBtn" class="secondary" type="button" disabled>Next</button>
							</div>
							<div class="muted" id="frameCounter">Step 0 of 0</div>
							<div class="footer-note">Use the slider to inspect the reconstructed path one frame at a time.</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>

	<script>
		const fileInput = document.getElementById('fileInput');
		const boardText = document.getElementById('boardText');
		const solveForm = document.getElementById('solveForm');
		const errorBox = document.getElementById('errorBox');
		const solveBtn = document.getElementById('solveBtn');
		const statFound = document.getElementById('statFound');
		const statCost = document.getElementById('statCost');
		const statExpanded = document.getElementById('statExpanded');
		const statElapsed = document.getElementById('statElapsed');
		const boardGrid = document.getElementById('boardGrid');
		const boardTitle = document.getElementById('boardTitle');
		const boardCoords = document.getElementById('boardCoords');
		const stepLabel = document.getElementById('stepLabel');
		const frameSlider = document.getElementById('frameSlider');
		const frameCounter = document.getElementById('frameCounter');
		const prevBtn = document.getElementById('prevBtn');
		const playBtn = document.getElementById('playBtn');
		const nextBtn = document.getElementById('nextBtn');
		const downloadBtn = document.getElementById('downloadBtn');
		const algorithm = document.getElementById('algorithm');
		const heuristic = document.getElementById('heuristic');

		let frames = [];
		let labels = [];
		let frameSize = { rows: 0, cols: 0 };
		let playing = false;
		let playTimer = null;
		let cellNodes = [];
		let previousActorIndex = -1;

		function parseFrame(frame) {
			return frame.map((row) => row.split(''));
		}

		function tileClass(ch) {
			if (ch === 'X') return 'tile wall';
			if (ch === 'O') return 'tile goal';
			if (ch === 'L') return 'tile hazard';
			if (ch === 'Z') return 'tile start';
			if (ch >= '0' && ch <= '9') return 'tile digit';
			return 'tile empty';
		}

		function tileLabel(ch) {
			if (ch === 'X' || ch === '*') return '';
			if (ch === 'Z') return 'Z';
			return ch;
		}

		function findActorIndex(grid) {
			for (let r = 0; r < grid.length; r++) {
				for (let c = 0; c < grid[r].length; c++) {
					if (grid[r][c] === 'Z') return r * grid[r].length + c;
				}
			}
			return -1;
		}

		function ensureBoard(grid) {
			if (!grid.length) return;
			if (cellNodes.length && frameSize.rows === grid.length && frameSize.cols === grid[0].length) return;
			frameSize = { rows: grid.length, cols: grid[0].length };
			boardGrid.innerHTML = '';
			boardGrid.style.gridTemplateColumns = 'repeat(' + frameSize.cols + ', minmax(42px, 1fr))';
			cellNodes = [];
			for (let r = 0; r < frameSize.rows; r++) {
				for (let c = 0; c < frameSize.cols; c++) {
					const cell = document.createElement('div');
					cell.className = 'tile empty';
					cell.dataset.row = String(r);
					cell.dataset.col = String(c);
					const span = document.createElement('span');
					span.className = 'tile-text';
					cell.appendChild(span);
					boardGrid.appendChild(cell);
					cellNodes.push(cell);
				}
			}
		}

		function animateBoardSwap(nextActorIndex) {
			boardGrid.classList.remove('is-changing');
			void boardGrid.offsetWidth;
			boardGrid.classList.add('is-changing');
			if (previousActorIndex >= 0 && previousActorIndex < cellNodes.length) {
				cellNodes[previousActorIndex].classList.remove('moved');
				void cellNodes[previousActorIndex].offsetWidth;
				cellNodes[previousActorIndex].classList.add('moved');
			}
			if (nextActorIndex >= 0 && nextActorIndex < cellNodes.length) {
				cellNodes[nextActorIndex].classList.remove('moved');
				void cellNodes[nextActorIndex].offsetWidth;
				cellNodes[nextActorIndex].classList.add('moved');
			}
			previousActorIndex = nextActorIndex;
		}

		function renderBoard(frameIndex) {
			if (!frames.length) {
				boardTitle.textContent = 'No frames available.';
				boardCoords.textContent = '0 x 0';
				boardGrid.innerHTML = '';
				cellNodes = [];
				previousActorIndex = -1;
				return;
			}
			const safeIndex = Math.max(0, Math.min(frameIndex, frames.length - 1));
			const grid = parseFrame(frames[safeIndex] || []);
			ensureBoard(grid);
			boardTitle.textContent = labels[safeIndex] || ('Step ' + safeIndex);
			boardCoords.textContent = frameSize.rows + ' x ' + frameSize.cols;
			let actorIndex = -1;
			for (let r = 0; r < frameSize.rows; r++) {
				for (let c = 0; c < frameSize.cols; c++) {
					const idx = r * frameSize.cols + c;
					const cell = cellNodes[idx];
					const ch = grid[r][c];
					cell.className = tileClass(ch);
					const text = tileLabel(ch);
					cell.querySelector('.tile-text').textContent = text;
					if (ch === 'Z') actorIndex = idx;
				}
			}
			animateBoardSwap(actorIndex);
		}

		function showError(message) {
			errorBox.textContent = message;
			errorBox.style.display = 'block';
		}

		function clearError() {
			errorBox.textContent = '';
			errorBox.style.display = 'none';
		}

		function renderFrame(index) {
			if (!frames.length) {
				stepLabel.textContent = 'Playback';
				frameCounter.textContent = 'Step 0 of 0';
				renderBoard(-1);
				return;
			}
			const safeIndex = Math.max(0, Math.min(index, frames.length - 1));
			renderBoard(safeIndex);
			stepLabel.textContent = labels[safeIndex] || ('Step ' + safeIndex);
			frameSlider.value = String(safeIndex);
			frameCounter.textContent = 'Step ' + (safeIndex + 1) + ' of ' + frames.length;
		}

		function setPlaying(nextPlaying) {
			playing = nextPlaying;
			playBtn.textContent = playing ? 'Pause' : 'Play';
			if (playTimer) {
				clearInterval(playTimer);
				playTimer = null;
			}
			if (playing) {
				playTimer = setInterval(() => {
					const current = Number(frameSlider.value || 0);
					if (current >= frames.length - 1) {
						setPlaying(false);
						return;
					}
					renderFrame(current + 1);
				}, 700);
			}
		}

		fileInput.addEventListener('change', async () => {
			const file = fileInput.files && fileInput.files[0];
			if (!file) return;
			boardText.value = await file.text();
		});

		frameSlider.addEventListener('input', () => renderFrame(Number(frameSlider.value || 0)));
		prevBtn.addEventListener('click', () => renderFrame(Number(frameSlider.value || 0) - 1));
		nextBtn.addEventListener('click', () => renderFrame(Number(frameSlider.value || 0) + 1));
		playBtn.addEventListener('click', () => setPlaying(!playing));

		algorithm.addEventListener('change', () => {
			const active = algorithm.value === 'GBFS' || algorithm.value === 'A*' || algorithm.value === 'WASTAR';
			if (algorithm.value === 'UCS') {
				heuristic.value = '-';
			} else if (heuristic.value === '-') {
				heuristic.value = 'H1';
			}
			heuristic.disabled = !active;
			document.getElementById('weighted').disabled = algorithm.value !== 'WASTAR';
		});
		algorithm.dispatchEvent(new Event('change'));

		async function runSolver() {
			clearError();
			solveBtn.disabled = true;
			solveBtn.textContent = 'Solving...';
			setPlaying(false);

			const formData = new URLSearchParams();
			formData.set('boardText', boardText.value);
			formData.set('algorithm', algorithm.value);
			formData.set('heuristic', heuristic.value);
			formData.set('weighted', document.getElementById('weighted').value || '1.8');

			try {
				const response = await fetch('/solve', { method: 'POST', body: formData });
				const payload = await response.json();
				if (!response.ok || !payload.ok) {
					throw new Error(payload.error || 'Solve request failed');
				}

				frames = payload.frames || [];
				labels = payload.frameMoves || [];
				previousActorIndex = -1;
				statFound.textContent = payload.found ? 'Yes' : 'No';
				statFound.className = 'v ' + (payload.found ? 'ok' : 'warn');
				statCost.textContent = String(payload.totalCost ?? '-');
				statExpanded.textContent = String(payload.expanded ?? '-');
				statElapsed.textContent = Number(payload.elapsedMS || 0).toFixed(3) + ' ms';
				downloadBtn.classList.remove('hidden');

				if (!frames.length) {
					frames = [['No state available']];
					labels = ['Initial'];
				}

				frameSlider.disabled = frames.length <= 1;
				prevBtn.disabled = frames.length <= 1;
				nextBtn.disabled = frames.length <= 1;
				playBtn.disabled = frames.length <= 1;
				frameSlider.min = '0';
				frameSlider.max = String(Math.max(0, frames.length - 1));
				renderFrame(0);
				// Auto-start playback when the solver finishes and there is more than one frame
				if (frames.length > 1) {
					setPlaying(true);
				}
			} catch (err) {
				showError(err.message || String(err));
			} finally {
				solveBtn.disabled = false;
				solveBtn.textContent = 'Run solver';
			}
		}

		solveBtn.addEventListener('click', runSolver);
		solveForm.addEventListener('submit', (event) => {
			event.preventDefault();
			runSolver();
		});

		renderFrame(0);
	</script>
</body>
</html>`
