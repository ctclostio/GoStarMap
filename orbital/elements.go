package orbital

import "math"

// OrbitalElements contains Keplerian orbital parameters
// All angles are in radians internally, but typically specified in degrees
type OrbitalElements struct {
	SemiMajorAxis         float64 // a (AU)
	Eccentricity          float64 // e (dimensionless, 0-1)
	Inclination           float64 // i (radians)
	LongitudeAscendingNode float64 // Ω (radians)
	ArgumentOfPeriapsis   float64 // ω (radians)
	MeanAnomalyAtEpoch    float64 // M₀ (radians at J2000.0 epoch)
	OrbitalPeriod         float64 // T (days)
}

// OrbitalState represents a body's position and velocity at a specific time
type OrbitalState struct {
	Position [3]float64 // x, y, z in AU (heliocentric)
	Velocity [3]float64 // vx, vy, vz in AU/day
	Time     float64    // days since J2000.0 epoch
}

// GetPlanetaryElements returns the validated orbital elements for solar system planets
// Data source: JPL Horizons System (J2000.0 epoch)
// All angles converted from degrees to radians
func GetPlanetaryElements(planetName string) OrbitalElements {
	// Conversion factor
	degToRad := math.Pi / 180.0

	elements := map[string]OrbitalElements{
		"Mercury": {
			SemiMajorAxis:          0.387,
			Eccentricity:           0.2056,
			Inclination:            7.005 * degToRad,
			LongitudeAscendingNode: 48.331 * degToRad,
			ArgumentOfPeriapsis:    29.124 * degToRad,
			MeanAnomalyAtEpoch:     174.796 * degToRad, // At J2000.0
			OrbitalPeriod:          87.969,
		},
		"Venus": {
			SemiMajorAxis:          0.723,
			Eccentricity:           0.0067,
			Inclination:            3.395 * degToRad,
			LongitudeAscendingNode: 76.680 * degToRad,
			ArgumentOfPeriapsis:    54.884 * degToRad,
			MeanAnomalyAtEpoch:     50.115 * degToRad,
			OrbitalPeriod:          224.701,
		},
		"Earth": {
			SemiMajorAxis:          1.000,
			Eccentricity:           0.0167,
			Inclination:            0.000 * degToRad, // Reference plane
			LongitudeAscendingNode: -11.26 * degToRad,
			ArgumentOfPeriapsis:    114.208 * degToRad,
			MeanAnomalyAtEpoch:     358.617 * degToRad,
			OrbitalPeriod:          365.256,
		},
		"Mars": {
			SemiMajorAxis:          1.524,
			Eccentricity:           0.0934,
			Inclination:            1.850 * degToRad,
			LongitudeAscendingNode: 49.558 * degToRad,
			ArgumentOfPeriapsis:    286.502 * degToRad,
			MeanAnomalyAtEpoch:     19.412 * degToRad,
			OrbitalPeriod:          686.980,
		},
		"Jupiter": {
			SemiMajorAxis:          5.203,
			Eccentricity:           0.0489,
			Inclination:            1.303 * degToRad,
			LongitudeAscendingNode: 100.464 * degToRad,
			ArgumentOfPeriapsis:    273.867 * degToRad,
			MeanAnomalyAtEpoch:     20.020 * degToRad,
			OrbitalPeriod:          4332.59,
		},
		"Saturn": {
			SemiMajorAxis:          9.537,
			Eccentricity:           0.0565,
			Inclination:            2.485 * degToRad,
			LongitudeAscendingNode: 113.665 * degToRad,
			ArgumentOfPeriapsis:    339.392 * degToRad,
			MeanAnomalyAtEpoch:     317.020 * degToRad,
			OrbitalPeriod:          10759.22,
		},
		"Uranus": {
			SemiMajorAxis:          19.191,
			Eccentricity:           0.0457,
			Inclination:            0.773 * degToRad,
			LongitudeAscendingNode: 74.006 * degToRad,
			ArgumentOfPeriapsis:    96.998 * degToRad,
			MeanAnomalyAtEpoch:     142.238 * degToRad,
			OrbitalPeriod:          30688.5,
		},
		"Neptune": {
			SemiMajorAxis:          30.069,
			Eccentricity:           0.0113,
			Inclination:            1.770 * degToRad,
			LongitudeAscendingNode: 131.784 * degToRad,
			ArgumentOfPeriapsis:    276.336 * degToRad,
			MeanAnomalyAtEpoch:     256.228 * degToRad,
			OrbitalPeriod:          60182.0,
		},
	}

	if elem, ok := elements[planetName]; ok {
		return elem
	}

	// Default to circular orbit at 1 AU if planet not found
	return OrbitalElements{
		SemiMajorAxis:       1.0,
		Eccentricity:        0.0,
		Inclination:         0.0,
		OrbitalPeriod:       365.25,
		MeanAnomalyAtEpoch:  0.0,
	}
}

// CalculatePlanetPosition is a convenience function that combines element lookup
// and position calculation for a named planet
func CalculatePlanetPosition(planetName string, timeDays float64) [3]float64 {
	elements := GetPlanetaryElements(planetName)
	return OrbitalElementsToPosition(elements, timeDays)
}
