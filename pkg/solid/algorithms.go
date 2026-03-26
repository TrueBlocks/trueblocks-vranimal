package solid

import (
	"math"

	"github.com/TrueBlocks/trueblocks-vranimal/pkg/vec"
)

// ---------------------------------------------------------------------------
// Hit types for intersection records
// ---------------------------------------------------------------------------

const (
	NoHit           = 0
	FaceHit         = 1
	VertexHit       = 2
	EdgeHit         = 3
	InsideLoopNoHit = 4
)

// IntersectRecord records information about an intersection calculation.
type IntersectRecord struct {
	He   *HalfEdge
	Vert *Vertex
	Type int
}

// NewIntersectRecord creates a default no-hit record.
func NewIntersectRecord() IntersectRecord {
	return IntersectRecord{Type: NoHit}
}

// RecordHit records a hit of the given type.
func (r *IntersectRecord) RecordHit(he *HalfEdge, v *Vertex, hitType int) {
	r.Type = hitType
	switch hitType {
	case FaceHit:
		r.He = nil
		r.Vert = nil
	case VertexHit:
		r.Vert = v
		r.He = nil
	case EdgeHit:
		r.He = he
		r.Vert = nil
	case InsideLoopNoHit:
		r.He = nil
		r.Vert = nil
	}
}

// GetType returns the hit type.
func (r *IntersectRecord) GetType() int { return r.Type }

// ---------------------------------------------------------------------------
// GetDominantComp returns the index (0, 1, 2) of the largest absolute
// component of a normal vector. Used to project 3D to 2D.
// ---------------------------------------------------------------------------

func GetDominantComp(n vec.SFVec3f) int {
	ax := float64(math.Abs(float64(n.X)))
	ay := float64(math.Abs(float64(n.Y)))
	az := float64(math.Abs(float64(n.Z)))
	if ax >= ay && ax >= az {
		return 0
	}
	if ay >= az {
		return 1
	}
	return 2
}

// ---------------------------------------------------------------------------
// Collinear returns true if three points are collinear.
// ---------------------------------------------------------------------------

func Collinear(a, b, c vec.SFVec3f) bool {
	ab := b.Sub(a)
	ac := c.Sub(a)
	cross := ab.Cross(ac)
	return cross.Length() < 1e-7
}

// ---------------------------------------------------------------------------
// IntersectsEdgeVertex tests if vertex v3 lies on the line through v1->v2.
// Returns the parametric t value along v1->v2.
// ---------------------------------------------------------------------------

func IntersectsEdgeVertex(v1, v2, v3 vec.SFVec3f) (float64, bool) {
	r1 := v2.Sub(v1)
	lenSq := r1.Dot(r1)
	if lenSq == 0 {
		if v1.Eq(v3) {
			return 0, true
		}
		return 0, false
	}

	r2 := v3.Sub(v1)
	t := r1.Dot(r2) / lenSq
	testV := v1.Add(r1.Scale(t))

	if testV.Eq(v3) {
		return t, true
	}
	return t, false
}

// ---------------------------------------------------------------------------
// ContainsOnEdge tests if vertex v3 lies on the segment v1->v2
// (parametric t in [0,1]).
// ---------------------------------------------------------------------------

func ContainsOnEdge(v1, v2, v3 vec.SFVec3f) bool {
	t, ok := IntersectsEdgeVertex(v1, v2, v3)
	if !ok {
		return false
	}
	return t >= -0.00001 && t <= 1.00001
}

// ---------------------------------------------------------------------------
// IntersectsEdgeEdge tests if two 2D-projected edges intersect.
// Projects to 2D by dropping the coordinate at index "drop".
// Returns parametric values (t1, t2) and whether they intersect.
// ---------------------------------------------------------------------------

func IntersectsEdgeEdge(v1, v2, v3, v4 vec.SFVec3f, drop int) (t1, t2 float64, ok bool) {
	var a1, a2, b1, b2, c1, c2 float64

	switch drop {
	case 0:
		a1 = v2.Y - v1.Y
		a2 = v2.Z - v1.Z
		b1 = v3.Y - v4.Y
		b2 = v3.Z - v4.Z
		c1 = v1.Y - v3.Y
		c2 = v1.Z - v3.Z
	case 1:
		a1 = v2.X - v1.X
		a2 = v2.Z - v1.Z
		b1 = v3.X - v4.X
		b2 = v3.Z - v4.Z
		c1 = v1.X - v3.X
		c2 = v1.Z - v3.Z
	default: // case 2
		a1 = v2.X - v1.X
		a2 = v2.Y - v1.Y
		b1 = v3.X - v4.X
		b2 = v3.Y - v4.Y
		c1 = v1.X - v3.X
		c2 = v1.Y - v3.Y
	}

	d := a1*b2 - a2*b1
	if d > -1e-7 && d < 1e-7 {
		return 0, 0, false
	}

	t1 = (c2*b1 - c1*b2) / d
	t2 = (a2*c1 - a1*c2) / d
	return t1, t2, true
}

// ---------------------------------------------------------------------------
// CrossingsTest implements the Graphics Gems IV crossing number algorithm.
// Tests if a 2D point is inside a 2D polygon. Returns true if inside.
// ---------------------------------------------------------------------------

func CrossingsTest(pgon [][2]float64, point [2]float64) bool {
	numverts := len(pgon)
	if numverts == 0 {
		return false
	}

	tx := point[0]
	ty := point[1]

	vtx0 := pgon[numverts-1]
	yflag0 := vtx0[1] >= ty
	inside := false

	for j := 0; j < numverts; j++ {
		vtx1 := pgon[j]
		yflag1 := vtx1[1] >= ty

		if yflag0 != yflag1 {
			xflag0 := vtx0[0] >= tx
			if xflag0 == (vtx1[0] >= tx) {
				if xflag0 {
					inside = !inside
				}
			} else {
				crossX := vtx1[0] - (vtx1[1]-ty)*(vtx0[0]-vtx1[0])/(vtx0[1]-vtx1[1])
				if crossX >= tx {
					inside = !inside
				}
			}
		}

		yflag0 = yflag1
		vtx0 = vtx1
	}

	return inside
}

// ---------------------------------------------------------------------------
// Loop containment methods
// ---------------------------------------------------------------------------

// CheckForContainment projects the loop to 2D and uses the crossings test
// to determine if pt is inside the loop.
func (l *Loop) CheckForContainment(pt vec.SFVec3f, drop int) bool {
	if l.HalfEdges == nil {
		return false
	}

	var pgon [][2]float64
	l.ForEachHe(func(he *HalfEdge) bool {
		v := he.Vertex.Loc
		var p [2]float64
		switch drop {
		case 0:
			p[0] = v.Y
			p[1] = v.Z
		case 1:
			p[0] = v.X
			p[1] = v.Z
		default:
			p[0] = v.X
			p[1] = v.Y
		}
		pgon = append(pgon, p)
		return true
	})

	var pt2 [2]float64
	switch drop {
	case 0:
		pt2[0] = pt.Y
		pt2[1] = pt.Z
	case 1:
		pt2[0] = pt.X
		pt2[1] = pt.Z
	default:
		pt2[0] = pt.X
		pt2[1] = pt.Y
	}

	return CrossingsTest(pgon, pt2)
}

// BoundaryContains checks if a vertex coincides with any vertex or
// lies on any edge of this loop.
func (l *Loop) BoundaryContains(v *Vertex, rec *IntersectRecord) bool {
	var found bool
	l.ForEachHe(func(he *HalfEdge) bool {
		if he.Vertex.Loc.Eq(v.Loc) {
			rec.RecordHit(nil, he.Vertex, VertexHit)
			found = true
			return false
		}
		return true
	})
	if found {
		return true
	}

	l.ForEachHe(func(he *HalfEdge) bool {
		if ContainsOnEdge(he.Vertex.Loc, he.Next.Vertex.Loc, v.Loc) {
			rec.RecordHit(he, nil, EdgeHit)
			found = true
			return false
		}
		return true
	})
	return found
}

// LoopContains checks if vertex v is inside this loop.
func (l *Loop) LoopContains(v *Vertex, drop int, rec *IntersectRecord) bool {
	if l.BoundaryContains(v, rec) {
		return true
	}
	return l.CheckForContainment(v.Loc, drop)
}

// ---------------------------------------------------------------------------
// Face containment
// ---------------------------------------------------------------------------

// FaceContains tests if vertex v is contained within face f.
func (f *Face) FaceContains(v *Vertex, rec *IntersectRecord) bool {
	drop := GetDominantComp(f.Normal)

	if f.LoopOut == nil || !f.LoopOut.LoopContains(v, drop, rec) {
		return false
	}

	if rec.Type == VertexHit || rec.Type == EdgeHit {
		return true
	}

	for _, l := range f.Loops {
		if l == f.LoopOut {
			continue
		}
		if l.NHalfEdges() < 3 {
			continue
		}
		if l.LoopContains(v, drop, rec) {
			if rec.Type == VertexHit || rec.Type == EdgeHit {
				return true
			}
			rec.RecordHit(nil, nil, InsideLoopNoHit)
			return false
		}
	}

	rec.RecordHit(nil, nil, FaceHit)
	return true
}

// ---------------------------------------------------------------------------
// Solid containment
// ---------------------------------------------------------------------------

// SolidContains tests if solid other is entirely contained within solid s.
func (s *Solid) SolidContains(other *Solid) bool {
	for v := other.Verts; v != nil; v = v.Next {
		conclusive := true
		for f := s.Faces; f != nil; f = f.Next {
			d := f.GetDistance(v.Loc)
			if d > -1e-5 && d < 1e-5 {
				conclusive = false
			} else if d > 0.00001 {
				return false
			}
		}
		if conclusive {
			return true
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// TranslationalSweep extrudes a face along a direction vector.
// ---------------------------------------------------------------------------

func (s *Solid) TranslationalSweep(f *Face, dir vec.SFVec3f) {
	for _, l := range f.Loops {
		first := l.GetFirstHe()
		if first == nil {
			continue
		}

		scan := first.Next
		v := scan.Vertex
		Lmev2(scan, scan, v.Loc.Add(dir))

		for scan != first {
			v = scan.Next.Vertex
			Lmev2(scan.Next, scan.Next, v.Loc.Add(dir))
			Lmef(scan.Prev, scan.Next.Next)
			scan = scan.Next.GetMate().Next
		}
		Lmef(scan.Prev, scan.Next.Next)
	}
}

// ---------------------------------------------------------------------------
// RotationalSweep revolves the solid around the X axis.
// ---------------------------------------------------------------------------

func (s *Solid) RotationalSweep(nFaces int) {
	if nFaces < 1 {
		return
	}

	closedFig := false
	var headf, tailf *Face
	if !s.IsWire() {
		if !s.IsLamina() {
			return
		}
		closedFig = true
		he := s.Faces.GetFirstHe()
		Lmev2(he, he.GetMate().Next, he.Vertex.Loc)
		Lkef(he.Prev.GetMate())
	}

	headf = s.Faces
	tailf = s.SweepWire(nFaces)

	if closedFig {
		Lkfmrh(tailf, headf)
		s.LoopGlue(headf)
	} else {
		if headf.IsDegenerate() {
			s.CollapseFace(headf)
		}
		if tailf != nil && tailf.IsDegenerate() {
			s.CollapseFace(tailf)
		}
	}

	s.Renumber()
}

// SweepWire sweeps a wire profile around the X axis. Returns the tail face.
func (s *Solid) SweepWire(nFaces int) *Face {
	if !s.IsWire() || s.Faces == nil {
		return nil
	}

	angle := 2 * math.Pi / float64(nFaces)
	cosA := float64(math.Cos(angle))
	sinA := float64(math.Sin(angle))

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

	// Find one end of the wire: advance while consecutive edges differ,
	// stop when they match (the turn-around point).
	for first.Edge != first.Next.Edge {
		first = first.Next
	}
	last := first.Next
	// Find the other end of the wire.
	for last.Edge != last.Next.Edge {
		last = last.Next
	}

	cfirst := first
	var scan *HalfEdge
	var tailf *Face

	for i := nFaces; i > 1; i-- {
		v := rotate(cfirst.Next.Vertex.Loc)
		Lmev2(cfirst.Next, cfirst.Next, v)
		scan = cfirst.Next
		for scan != last.Next {
			v = rotate(scan.Prev.Vertex.Loc)
			Lmev2(scan.Prev, scan.Prev, v)
			Lmef(scan.Prev.Prev, scan.Next)
			scan = scan.Next.Next.GetMate()
		}
		last = scan
		cfirst = cfirst.Next.Next.GetMate()
	}

	if scan == nil {
		return nil
	}

	tailf, _ = Lmef(cfirst.Next, first.GetMate())
	for cfirst != scan {
		Lmef(cfirst, cfirst.Next.Next.Next)
		cfirst = cfirst.Prev.GetMate().Prev
	}

	return tailf
}

// ---------------------------------------------------------------------------
// Twist applies a twist deformation to all vertices.
// ---------------------------------------------------------------------------

func (s *Solid) Twist(fn func(float64) float64) {
	for v := s.Verts; v != nil; v = v.Next {
		z := fn(v.Loc.Z)
		cosZ := float64(math.Cos(float64(z)))
		sinZ := float64(math.Sin(float64(z)))
		x := v.Loc.X
		y := v.Loc.Y
		v.Loc.X = x*cosZ - y*sinZ
		v.Loc.Y = x*sinZ + y*cosZ
	}
}

// ---------------------------------------------------------------------------
// Glue merges two solids at coincident faces.
// ---------------------------------------------------------------------------

func (s *Solid) Glue(s2 *Solid, f1, f2 *Face) {
	s.Merge(s2)
	Lkfmrh(f2, f1)
	s.LoopGlue(f1)
}

// LoopGlue joins duplicate loops at matching vertices.
func (s *Solid) LoopGlue(f *Face) {
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

	Lmekr(h1, h2)
	Lkev(h1.Prev)

	for h1.Next != h2 {
		h1next := h1.Next
		Lmef(h1.Next, h1.Prev)
		Lkev(h1.Next)
		Lkef(h1.GetMate())
		h1 = h1next
	}

	Lkef(h1.GetMate())
}

// ---------------------------------------------------------------------------
// Join connects two half-edges with an edge.
// ---------------------------------------------------------------------------

func (s *Solid) Join(he1, he2 *HalfEdge, swap bool) {
	oldFace := he1.GetFace()

	var newFace *Face
	if he1.Loop == he2.Loop {
		if he1.Prev.Prev != he2 {
			if swap {
				newFace, _ = Lmef(he2.Next, he1)
			} else {
				newFace, _ = Lmef(he1, he2.Next)
			}
		}
	} else {
		Lmekr(he1, he2.Next)
	}

	if he1.Next.Next != he2 {
		Lmef(he2, he1.Next)
		if newFace != nil && oldFace.GetSecondLoop() != nil {
			for len(oldFace.Loops) > 1 {
				l := oldFace.Loops[len(oldFace.Loops)-1]
				if l == oldFace.LoopOut {
					break
				}
				Lringmv(s, l, newFace, false)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Cut disconnects at an edge.
// ---------------------------------------------------------------------------

func (s *Solid) Cut(he *HalfEdge, setOp bool) *Face {
	if he.Edge == nil {
		return nil
	}

	he1 := he.Edge.He1
	he2 := he.Edge.He2

	if he1.Loop == he2.Loop {
		oldFace := he1.GetFace()
		if setOp {
			Lkemr(he2)
		} else {
			Lkemr(he1)
		}
		return oldFace
	}

	Lkef(he1)
	return nil
}

// ---------------------------------------------------------------------------
// MoveFace recursively moves a face and neighbors to another solid.
// ---------------------------------------------------------------------------

func (s *Solid) MoveFace(f *Face, target *Solid) {
	if f == nil || f.Marked2(VISITED) {
		return
	}

	f.Mark2 = VISITED

	// Remove from the face's own solid (which may differ from s
	// when null edges connect faces across two solids).
	owner := f.Solid
	owner.RemoveFace(f)
	target.AddFace(f)

	for _, l := range f.Loops {
		l.ForEachHe(func(he *HalfEdge) bool {
			mf := he.GetMateFace()
			if mf != nil && mf.Solid != target {
				mf.Solid.MoveFace(mf, target)
			}
			return true
		})
	}
}

// ---------------------------------------------------------------------------
// CollapseFace removes a degenerate face by collapsing its edges.
// ---------------------------------------------------------------------------

func (s *Solid) CollapseFace(f *Face) {
	if f == nil || f.GetSecondLoop() != nil {
		return
	}

	he := f.GetFirstHe()
	if he == nil {
		return
	}
	hen := he.Next
	hep := he.Prev

	Lkef(he.GetMate())

	he = hen
	for he != hep {
		hen = he.Next
		Lkev(he)
		he = hen
	}
	Lkev(he)
}

// ---------------------------------------------------------------------------
// RemoveColinearVerts removes collinear vertices from a face.
// ---------------------------------------------------------------------------

func (f *Face) RemoveColinearVerts() {
	for _, l := range f.Loops {
		if l.HalfEdges == nil {
			continue
		}

		start := l.HalfEdges
		he := start
		prevV := he.Prev.Vertex.Loc
		v := he.Vertex.Loc

		for {
			nextV := he.Next.Vertex.Loc

			canRemove := he.GetMate() != nil &&
				he.Prev.GetMate() != nil &&
				he.Prev.GetMate() == he.GetMate().Next &&
				Collinear(prevV, v, nextV)

			if canRemove {
				n := he.Next
				if he == start {
					if he.GetMate() == he.Prev {
						start = he.Prev.Prev
					} else {
						start = he.Prev
					}
				}
				Lkev(he)
				v = nextV
				he = n
			} else {
				he = he.Next
				prevV = v
				v = nextV
			}

			if he == start {
				break
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Triangulate splits all faces into triangles.
// ---------------------------------------------------------------------------

func (s *Solid) Triangulate() {
	for f := s.Faces; f != nil; f = f.Next {
		s.TriangulateFace(f)
	}
}

// TriangulateFace splits a single face into triangles via fan triangulation.
func (s *Solid) TriangulateFace(f *Face) {
	if f.LoopOut == nil {
		return
	}

	for f.LoopOut.NHalfEdges() > 3 {
		he := f.LoopOut.GetFirstHe()
		if he == nil {
			return
		}
		target := he.Next.Next
		Lmef(he, target)
		// After Lmef, he moved to the new face; f.LoopOut retains the
		// remaining polygon with one fewer half-edge.
	}
}

// ---------------------------------------------------------------------------
// Verify checks the Euler formula: F + V - 2 = E + H.
// ---------------------------------------------------------------------------

func (s *Solid) Verify() bool {
	nLoops := 0
	for f := s.Faces; f != nil; f = f.Next {
		nLoops += f.NLoops()
	}
	nHoles := nLoops - s.nFaces
	expected := s.nFaces + s.nVerts - 2
	got := s.nEdges + nHoles
	return expected == got
}

// ---------------------------------------------------------------------------
// Arc creates arc geometry via mev operators.
// ---------------------------------------------------------------------------

func (s *Solid) Arc(face *Face, startVertex *Vertex, cx, cy, radius, height, startAngle, endAngle float64, steps int) {
	angle := float64(startAngle) * math.Pi / 180.0
	stepSize := float64(endAngle-startAngle) * math.Pi / (180.0 * float64(steps))

	prevHe := startVertex.He
	for i := 0; i < steps; i++ {
		angle += stepSize
		x := cx + float64(math.Cos(angle))*radius
		y := cy + float64(math.Sin(angle))*radius
		Lmev(prevHe, vec.SFVec3f{X: x, Y: y, Z: height})
		prevHe = prevHe.Prev
	}
}

// ArcSweep performs a translational sweep along a sequence of direction vectors.
func (s *Solid) ArcSweep(f *Face, dirs []vec.SFVec3f) {
	for _, d := range dirs {
		s.TranslationalSweep(f, d)
	}
}
