# Bool Test Failure Analysis

**Related issue:** [#37 — Bool Phase 3: Fix remaining 93 failing boolean tests](https://github.com/TrueBlocks/trueblocks-3d/issues/37)

## Are All Tests Valid?

Review of the test harness and test geometries reveals several concerns suggesting
that some "failures" may be test problems rather than code bugs.

### 1. No Input Validation

The test harness (`boolTestCase` in `bool_test.go`) calls `BoolOp(a, b, op)` without
ever running `VerifyDetailed()` on the input solids. If `MakeSphere` or lamina-based
prism builders produce slightly malformed B-reps, the bool pipeline gets blamed for
downstream corruption.

**Action:** Add `VerifyDetailed(a)` / `VerifyDetailed(b)` before `BoolOp` to expose
bad inputs.

### 2. Group 10 — Seven Identical Tests

All 7 cases in `TestBool_Group10_HexagonPrisms` use the exact same geometry with no
per-case translation or rotation. If Case 0 Union fails, all 7 Union tests fail
identically. This inflates the failure count from 1 real failure to 7.

**Action:** Either add per-case transforms (as all other groups do) or collapse to a
single representative case.

### 3. Mathematically Degenerate Inputs

Several groups test geometry that is inherently degenerate for B-rep CSG:

| Test | Degeneracy | Notes |
|------|-----------|-------|
| Group 6 Case 0 — face-on-face | Coplanar shared face, zero-volume intersection | **Actually passes** |
| Group 6 Case 2 — edge-on-edge | Shared edge, zero-area contact | Fails (VV classify) |
| Group 6 Case 4 — vertex-on-vertex | Single-point contact | **Actually passes** |
| Group 6 Case 5 — identical cubes | A ≡ B, degenerate for all three ops | **Actually passes** |
| Group 5 Case 0 — 4×4 sphere | Only 14 faces, barely a polyhedron | |

After testing, the "degenerate" cases (face-on-face, vertex-on-vertex, identical) all
pass. The failures in Group 6 are Cases 1, 2, 6 — partial overlap, edge-on-edge, and
slight twist — which exercise the VV classify pipeline, not degenerate math.

### 4. Group 13 = Group 6 With Operands Reversed

Group 13 (SphereVsCube) reruns similar contact scenarios with sphere as A instead of
cube. Failures here compound the sphere-tessellation issue with VV-contact issues.

---

## Difficulty Categorization

### Easy (~25–30 tests) — Test fixes or one-line code fixes

| Category | Tests | Description |
|----------|-------|-------------|
| Group 10 dedup | ~7 | All 7 identical Union failures are 1 real bug or test error |
| Phase 3b Lkef LoopOut | varies | One-line fix: `if keepF.LoopOut == killLoop { keepF.LoopOut = keepLoop }` |
| Degenerate reclassification | ~10–15 | face-on-face, identical solids, vertex-vertex contact → mark as unsupported |
| Input validation | 0 fixed | But immediately clarifies which failures are caused by bad inputs |

### Medium (~30–35 tests) — Clear fix path, moderate effort

| Category | Tests | Description |
|----------|-------|-------------|
| Groups 3–4 (rotated elongated cubes) | ~20 | Some ops pass, others fail; face-selection logic in Finish |
| Group 5 (sphere vs cube) | ~5–7 | Only Union fails; multi-zone LoopGlue fix (Phase 3a) |
| Groups 11–12 (L-prisms) | ~6–9 | Non-convex geometry stresses classify; specific Z-offset cases |

### Hard (~25–30 tests) — Deep algorithmic work

| Category | Tests | Description |
|----------|-------|-------------|
| "Through" cases (Groups 0–2) | ~3–6 | B passes through A creating 2+ intersection zones; LoopGlue only handles nLoops==2 |
| VV classify (Groups 6–9) | ~10–15 | Vertex-vertex coincidence requires sector analysis; ~400 lines of C++ sector-walking |
| Group 13 (sphere-as-A, all ops) | ~15 | Combines sphere tessellation + VV issues + reversed operand order |

---

## Changes Applied

### 1. Input validation added to `boolTestCase`
Calls `VerifyDetailed()` on both inputs before `BoolOp`. Found that **Group 14**
(CubeVsLShape) had an invalid input B — `makeLShape` used `Merge` to combine two
cubes into a multi-shell solid, violating the single-shell Euler formula assumption.

### 2. `makeLShape` rebuilt as swept lamina
Replaced the two-bar `Merge` hack with a proper 8-vertex L-shaped polygon extruded
along Z. All Group 14 tests now have valid B-rep input. This exposed 13 new real
failures (previously masked by bad input), added to `knownFailingBoolTests`.

### 3. Group 10 collapsed from 7 to 1 case
C++ source confirms all 7 hexagon prism cases use identical `{0,0,0}` translations.
Reduced to a single representative `Case0_overlapping`, eliminating 6 redundant
entries from the known-failing map.

### 4. Lkef LoopOut preservation
Added `if keepF.LoopOut == killLoop { keepF.LoopOut = keepLoop }` in `euler.go`.
Prevents nil LoopOut when the killed loop happens to be the outer loop. Testing
showed this fix alone doesn't resolve any of the 101 remaining failures (all are
Euler formula violations at root), but it's a correctness improvement.

### 5. Degenerate cases — not the problem
Testing confirmed that face-on-face, vertex-on-vertex, and identical-solid cases
all pass. The actual failures are in VV classify paths (partial overlap, edge-on-edge,
slight twist), not mathematical degeneracy.

## Updated Counts

| Metric | Before | After |
|--------|--------|-------|
| Passing | 276 | 268 |
| Skipped (known-failing) | 94 | 101 |
| Total subtests | 370+ | 369 |
| Invalid inputs | 13 (silent) | 0 |
| Redundant tests | 6 | 0 |

Note: "Passing" decreased because 13 tests that previously silently passed with bad
input now correctly skip as known-failing with valid input. The test suite is now
more honest.

---

## Key Observations

### C++ vs Go LoopGlue Comparison

The Go `LoopGlue` faithfully ports the C++ version. Both assume exactly 2 loops
(outer + one inner). When the bool pipeline produces faces with 3+ loops (multi-zone
"through" geometry), both versions only glue the first pair and leave orphans.

This is not a porting bug — it is a limitation of the original algorithm.

### Euler Operator API Simplification

The Go port simplified Euler operators from 2-parameter (`Lkev(he1, he2)`) to
1-parameter (`Lkev(he)`) by deriving the mate internally via `GetMate()`. This is
correct — the mate relationship is guaranteed by the data structure. Not a bug source.

### Nil LoopOut Is Always Secondary

Every "nil outer loop" failure co-occurs with an Euler formula violation. The nil
LoopOut fixup in `bool_finish.go` catches most cases, but when the underlying Euler
topology is already wrong, the fixup assigns `Loops[0]` which may not be the true
outer loop.
