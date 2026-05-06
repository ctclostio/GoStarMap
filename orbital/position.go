package orbital

// PositionToRenderUnits converts astronomical position (AU) to render units
// using your current scale system (1 AU = 150 units)
func PositionToRenderUnits(positionAU [3]float64, auScale float64) [3]float32 {
	return [3]float32{
		float32(positionAU[0] * auScale),
		float32(positionAU[2] * auScale), // Swap Y and Z for your coordinate system
		float32(positionAU[1] * auScale), // raylib uses Y-up, astronomy uses Z-up
	}
}
