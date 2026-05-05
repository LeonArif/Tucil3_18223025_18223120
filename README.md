# Tucil3_18223025_18223120

Ice Sliding Puzzle Solver untuk Tugas Kecil 3 IF2211 Strategi Algoritma 2025/2026.

## Fitur

Spesifikasi wajib:
- Parser input `.txt` dengan validasi format, ukuran papan, simbol, urutan digit, dan matriks cost.
- Solver pathfinding:
	- `UCS`
	- `GBFS`
	- `A*`
- Menampilkan:
	- solusi gerakan
	- total cost solusi
	- visualisasi papan per langkah solusi
	- banyak iterasi/konfigurasi yang ditinjau
	- waktu eksekusi (ms, hanya proses pencarian)
- Playback setelah proses selesai (CLI):
	- maju/mundur step
	- lompat ke step tertentu
- Simpan hasil solusi ke file `.txt`.

Bonus yang dikerjakan:
- GUI penuh untuk input, eksekusi, visualisasi, dan playback (maju/mundur/play/pause/jump via slider).
- Algoritma tambahan: `WASTAR` (Weighted A*).
- Dua heuristik tambahan: `H4`, `H5` (di luar `H1`, `H2`, `H3`).

## Struktur Repository

- `src/` : source code program.
- `bin/` : executable (disiapkan, dapat diisi saat packaging).
- `test/` : test case dan contoh output solusi.
- `doc/` : laporan PDF.

## Requirement

- Python 3.10+ (disarankan 3.11+).
- Library tambahan tidak wajib (GUI menggunakan `tkinter` bawaan Python).

## Format Input

File input `.txt` berisi:
1. Baris 1: `N M`
2. `N` baris peta (`*`, `X`, `L`, `Z`, `O`, `0`..`9`)
3. `N` baris cost traversal (masing-masing `M` bilangan bulat non-negatif)

Contoh ada di `test/sample1.txt`.

## Cara Menjalankan

Mode CLI:

```bash
python src/main.py
```

Mode GUI:

```bash
python src/main.py --gui
```

## Contoh Alur CLI

1. Masukkan path file input (contoh: `test/sample1.txt`)
2. Pilih algoritma (`UCS/GBFS/A*/WASTAR`)
3. Jika perlu, pilih heuristic (`H1/H2/H3/H4/H5`)
4. Program menampilkan solusi, cost, iterasi, dan waktu
5. Opsional playback
6. Opsional simpan output ke `.txt`

## Catatan Implementasi

- State solver: `(posisi_aktor, next_digit_yang_harus_dilewati)`.
- Game over jika:
	- sliding keluar papan,
	- melewati `L`,
	- melewati digit lebih besar sebelum urutannya.
- Goal valid jika aktor berhenti tepat di `O` dan semua digit wajib sudah dilalui.

## Author

- 18223025
- 18223120