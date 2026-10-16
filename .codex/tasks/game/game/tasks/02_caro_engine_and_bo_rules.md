# Task 02: Caro Engine & Best-Of (BO) Set Rules

## Objective
Implement local 15x15 Caro matrix logic, 5-in-a-row win detection algorithm, BO set score manager, and 5-minute countdown timer.

## Instructions
1. Create pkg/engine/board.go:
   - Matrix [15][15]int (0: Empty, 1: X, 2: O).
   - PlaceTile(x, y, symbol) bool: Place tile if cell is empty.
   - CheckWin(x, y) (bool, [][2]int): Check horizontal, vertical, and diagonals around (x, y) for 5 consecutive matching symbols.
2. Create pkg/engine/bo_manager.go:
   - Track scoreP1 and scoreP2.
   - Calculate target wins: target = (maxGames / 2) + 1.
   - IsSetFinished() bool: Return true when P1 or P2 reaches target wins.
   - Implement 300-second (5 minutes) countdown timer when set is finished. Lock board inputs.
