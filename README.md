# GoStarMap

A Go-based celestial navigation and star mapping system for visualizing and calculating distances between stars in 3D space.

## Features

- 3D star coordinate system
- Distance calculations between celestial bodies
- Find nearest neighbors to any star
- Performance benchmarking
- Extensible star database

## Installation

```bash
go mod download
```

## Usage

Run the application:

```bash
go run main.go
```

Build the binary:

```bash
go build -o gostarmap
./gostarmap
```

## Project Structure

- `main.go` - Main application with Star and StarMap types
- `go.mod` - Go module definition

## Star Data

The application includes sample data for several well-known stars:
- Sun (reference point)
- Proxima Centauri
- Alpha Centauri A
- Barnard's Star
- Sirius
- Betelgeuse
- Rigel

## Future Enhancements

- Load star data from external databases (HYG, Hipparcos)
- 2D/3D visualization using graphics libraries
- REST API for star queries
- Constellation mapping
- Sky position calculations (RA/Dec)
- Time-based stellar positions
- Search by magnitude, spectral type, etc.
