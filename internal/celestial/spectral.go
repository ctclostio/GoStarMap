// Package celestial holds pure-Go astronomy helpers that don't depend on
// raylib. Keeping them here lets `go test ./...` run in environments that
// don't have raylib.dll on the path (CI, headless boxes).
package celestial

import "math/rand"

// SpectralType represents Morgan-Keenan stellar classification.
type SpectralType int

const (
	TypeO SpectralType = iota // Blue, very rare
	TypeB                     // Blue-white
	TypeA                     // White
	TypeF                     // Yellow-white
	TypeG                     // Yellow (like our Sun)
	TypeK                     // Orange
	TypeM                     // Red, most common
)

// RandomType returns a spectral type drawn from observed Milky Way ratios:
// O 0.003%, B 0.13%, A 0.6%, F 3%, G 7.6%, K 12.1%, M ~76.6%.
func RandomType() SpectralType {
	r := rand.Float32()
	switch {
	case r < 0.00003:
		return TypeO
	case r < 0.00133:
		return TypeB
	case r < 0.00733:
		return TypeA
	case r < 0.03733:
		return TypeF
	case r < 0.11333:
		return TypeG
	case r < 0.23433:
		return TypeK
	default:
		return TypeM
	}
}
