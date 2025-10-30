package main

import (
	"fmt"
	"math"
	"time"
)

// Star represents a celestial body with 3D coordinates
type Star struct {
	Name      string
	X, Y, Z   float64
	Magnitude float64
}

// StarMap holds a collection of stars
type StarMap struct {
	Stars []Star
}

// NewStarMap creates a new star map
func NewStarMap() *StarMap {
	return &StarMap{
		Stars: make([]Star, 0),
	}
}

// AddStar adds a star to the map
func (sm *StarMap) AddStar(star Star) {
	sm.Stars = append(sm.Stars, star)
}

// Distance calculates the distance between two stars
func (sm *StarMap) Distance(s1, s2 Star) float64 {
	dx := s1.X - s2.X
	dy := s1.Y - s2.Y
	dz := s1.Z - s2.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// FindNearestStars finds the n nearest stars to a given star
func (sm *StarMap) FindNearestStars(target Star, n int) []Star {
	type starDistance struct {
		star Star
		dist float64
	}

	distances := make([]starDistance, 0)
	for _, star := range sm.Stars {
		if star.Name != target.Name {
			dist := sm.Distance(target, star)
			distances = append(distances, starDistance{star, dist})
		}
	}

	// Simple bubble sort for small datasets
	for i := 0; i < len(distances)-1; i++ {
		for j := 0; j < len(distances)-i-1; j++ {
			if distances[j].dist > distances[j+1].dist {
				distances[j], distances[j+1] = distances[j+1], distances[j]
			}
		}
	}

	nearest := make([]Star, 0)
	limit := n
	if limit > len(distances) {
		limit = len(distances)
	}
	for i := 0; i < limit; i++ {
		nearest = append(nearest, distances[i].star)
	}

	return nearest
}

// DisplayMap prints the star map
func (sm *StarMap) DisplayMap() {
	fmt.Println("\n=== Star Map ===")
	fmt.Printf("Total Stars: %d\n\n", len(sm.Stars))
	for i, star := range sm.Stars {
		fmt.Printf("%d. %s\n", i+1, star.Name)
		fmt.Printf("   Position: (%.2f, %.2f, %.2f)\n", star.X, star.Y, star.Z)
		fmt.Printf("   Magnitude: %.2f\n\n", star.Magnitude)
	}
}

func main() {
	fmt.Println("GoStarMap - Celestial Navigation System")
	fmt.Println("========================================")

	starMap := NewStarMap()

	// Add some well-known stars (approximate 3D coordinates in light-years)
	starMap.AddStar(Star{Name: "Sun", X: 0, Y: 0, Z: 0, Magnitude: -26.74})
	starMap.AddStar(Star{Name: "Proxima Centauri", X: 4.24, Y: 0.5, Z: -1.2, Magnitude: 11.13})
	starMap.AddStar(Star{Name: "Alpha Centauri A", X: 4.37, Y: 0.6, Z: -1.1, Magnitude: -0.01})
	starMap.AddStar(Star{Name: "Barnard's Star", X: 5.96, Y: 1.2, Z: 0.8, Magnitude: 9.53})
	starMap.AddStar(Star{Name: "Sirius", X: 8.6, Y: -2.5, Z: 3.1, Magnitude: -1.46})
	starMap.AddStar(Star{Name: "Betelgeuse", X: 548, Y: 120, Z: -85, Magnitude: 0.42})
	starMap.AddStar(Star{Name: "Rigel", X: 860, Y: -200, Z: 150, Magnitude: 0.13})

	// Display the star map
	starMap.DisplayMap()

	// Find nearest stars to Sun
	fmt.Println("=== 3 Nearest Stars to Sun ===")
	nearest := starMap.FindNearestStars(starMap.Stars[0], 3)
	for i, star := range nearest {
		dist := starMap.Distance(starMap.Stars[0], star)
		fmt.Printf("%d. %s (%.2f light-years away)\n", i+1, star.Name, dist)
	}

	// Performance benchmark
	fmt.Println("\n=== Performance Test ===")
	start := time.Now()
	for i := 0; i < 1000; i++ {
		starMap.FindNearestStars(starMap.Stars[0], 3)
	}
	elapsed := time.Since(start)
	fmt.Printf("1000 nearest star searches: %s\n", elapsed)
}
