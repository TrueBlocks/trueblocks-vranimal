package primitives

import (
	"math"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/base"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/solid/euler"
	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

func TranslationalSweep(s *base.Solid, f *base.Face, dir vec.SFVec3f) {
	for _, l := range f.Loops {
		first := l.GetFirstHe()
		if first == nil {
			continue
		}

		scan := first.Next
		v := scan.Vertex
		_, _, _ = euler.Lmev2(scan, scan, v.Loc.Add(dir))

		for scan != first {
			v = scan.Next.Vertex
			_, _, _ = euler.Lmev2(scan.Next, scan.Next, v.Loc.Add(dir))
			_, _, _ = euler.Lmef(scan.Prev, scan.Next.Next)
			scan = scan.Next.GetMate().Next
		}
		_, _, _ = euler.Lmef(scan.Prev, scan.Next.Next)
	}
}

func RotationalSweep(s *base.Solid, nFaces int) {
	if nFaces < 1 {
		return
	}

	closedFig := false
	var headf, tailf *base.Face
	if !s.IsWire() {
		if !s.IsLamina() {
			return
		}
		closedFig = true
		he := s.Faces.GetFirstHe()
		_, _, _ = euler.Lmev2(he, he.GetMate().Next, he.Vertex.Loc)
		_ = euler.Lkef(he.Prev.GetMate())
	}

	headf = s.Faces
	tailf = sweepWire(s, nFaces)

	if closedFig {
		_ = euler.Lkfmrh(tailf, headf)
		loopGlue(s, headf)
	} else {
		if headf.IsDegenerate() {
			collapseFace(s, headf)
		}
		if tailf != nil && tailf.IsDegenerate() {
			collapseFace(s, tailf)
		}
	}

	s.Renumber()
}

func sweepWire(s *base.Solid, nFaces int) *base.Face {
	if !s.IsWire() || s.Faces == nil {
		return nil
	}

	angle := 2 * math.Pi / float64(nFaces)
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)

	rotate := func(v vec.SFVec3f) vec.SFVec3f {
		return vec.SFVec3f{
			X: v.X,
			Y: v.Y*cosA - v.Z*sinA,
			Z: v.Y*sinA + v.Z*cosA,
		}
	}

	first := s.Faces.GetFirstHe()
	if first == nil {
		return nil
	}

	for first.Edge != first.Next.Edge {
		first = first.Next
	}
	last := first.Next
	for last.Edge != last.Next.Edge {
		last = last.Next
	}

	cfirst := first
	var scan *base.HalfEdge
	var tailf *base.Face

	for i := nFaces; i > 1; i-- {
		v := rotate(cfirst.Next.Vertex.Loc)
		_, _, _ = euler.Lmev2(cfirst.Next, cfirst.Next, v)
		scan = cfirst.Next
		for scan != last.Next {
			v = rotate(scan.Prev.Vertex.Loc)
			_, _, _ = euler.Lmev2(scan.Prev, scan.Prev, v)
			_, _, _ = euler.Lmef(scan.Prev.Prev, scan.Next)
			scan = scan.Next.Next.GetMate()
		}
		last = scan
		cfirst = cfirst.Next.Next.GetMate()
	}

	if scan == nil {
		return nil
	}

	tailf, _, _ = euler.Lmef(cfirst.Next, first.GetMate())
	for cfirst != scan {
		_, _, _ = euler.Lmef(cfirst, cfirst.Next.Next.Next)
		cfirst = cfirst.Prev.GetMate().Prev
	}

	return tailf
}

func loopGlue(_ *base.Solid, f *base.Face) {
	if f.NLoops() < 2 {
		return
	}

	h1 := f.GetFirstHe()
	h2Loop := f.GetSecondLoop()
	if h2Loop == nil {
		return
	}
	h2 := h2Loop.GetFirstHe()
	if h2 == nil {
		return
	}

	start := h2
	for !h1.Vertex.Loc.Eq(h2.Vertex.Loc) {
		h2 = h2.Next
		if h2 == start {
			return
		}
	}

	_, _ = euler.Lmekr(h1, h2)
	_ = euler.Lkev(h1.Prev)

	limit := 10000
	for h1.Next != h2 {
		h1next := h1.Next
		_, _, _ = euler.Lmef(h1.Next, h1.Prev)
		_ = euler.Lkev(h1.Next)
		_ = euler.Lkef(h1.GetMate())
		h1 = h1next
		limit--
		if limit <= 0 {
			return
		}
	}

	_ = euler.Lkef(h1.GetMate())
}

func collapseFace(_ *base.Solid, f *base.Face) {
	if f == nil || f.GetSecondLoop() != nil {
		return
	}

	he := f.GetFirstHe()
	if he == nil {
		return
	}
	hen := he.Next
	hep := he.Prev

	_ = euler.Lkef(he.GetMate())

	he = hen
	for he != hep {
		hen = he.Next
		_ = euler.Lkev(he)
		he = hen
	}
	_ = euler.Lkev(he)
}
