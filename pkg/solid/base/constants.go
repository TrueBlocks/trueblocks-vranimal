package base

const (
	ON    = 0
	ABOVE = 1
	BELOW = -1

	UNKNOWN      = 1 << 1
	BRANDNEW     = 1 << 2
	VISITED      = 1 << 3
	ALTERED      = 1 << 4
	CREASE       = 1 << 12
	PARTIALNORML = 1 << 13
	INVISIBLE    = 1 << 15
	DELETED      = 0xdddddddd
	POINTSET     = 1 << 20
	LINESET      = 1 << 21
	FACESET      = POINTSET | LINESET
	CALCED       = 1 << 25

	SOLIDNONE = 0
	SOLIDA    = 11
	SOLIDB    = 12

	BoolUnion        = 1
	BoolIntersection = 2
	BoolDifference   = 3

	PerFace          = 1
	PerVertex        = 2
	PerVertexPerFace = 3
)

const (
	ColorFlag    int64 = 0x1
	NormalFlag   int64 = 0x2
	TexCoordFlag int64 = 0x4
)

const (
	LOOSE     = 1 << 16
	NOT_LOOSE = 0
)

const (
	IN  = -1
	OUT = 1
)

const BigEps = 0.00001
