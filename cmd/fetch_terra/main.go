package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	shp "github.com/jonas-p/go-shp"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

const (
	sourceZip  = "https://naturalearth.s3.amazonaws.com/50m_cultural/ne_50m_admin_0_countries.zip"
	outputDir  = "data/terra/earth"
	outputFile = "countries.geojson"
)

func main() {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		panic(err)
	}

	fmt.Println("🌍 Downloading Natural Earth countries (1:50m)…")
	resp, err := http.Get(sourceZip)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("📦 Extracting archive…")
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		panic(fmt.Sprintf("Failed to open zip: %v", err))
	}

	tempDir, _ := os.MkdirTemp("", "ne_extract")
	var shpPath string
	for _, f := range zr.File {
		if filepath.Ext(f.Name) == ".shp" {
			shpPath = filepath.Join(tempDir, f.Name)
		}
		outPath := filepath.Join(tempDir, f.Name)
		os.MkdirAll(filepath.Dir(outPath), 0755)
		rc, _ := f.Open()
		out, _ := os.Create(outPath)
		io.Copy(out, rc)
		rc.Close()
		out.Close()
	}

	if shpPath == "" {
		fmt.Println("❌ No .shp found in archive")
		return
	}

	fmt.Println("🗺️ Converting Shapefile → GeoJSON (pure Go)…")

	shape, err := shp.Open(shpPath)
	if err != nil {
		panic(err)
	}
	defer shape.Close()

	fields := shape.Fields()
	fc := geojson.NewFeatureCollection()

	for shape.Next() {
		n, geom := shape.Shape()
		props := make(map[string]interface{})
		for i, f := range fields {
			val := shape.ReadAttribute(n, i)
			props[f.String()] = val
		}

		switch g := geom.(type) {
		case *shp.Polygon:
			points := make([]orb.Point, len(g.Points))
			for i, p := range g.Points {
				points[i] = orb.Point{p.X, p.Y}
			}
			poly := orb.Polygon{points}
			f := geojson.NewFeature(poly)
			f.Properties = props
			fc.Append(f)

		case *shp.PolyLine:
			points := make([]orb.Point, len(g.Points))
			for i, p := range g.Points {
				points[i] = orb.Point{p.X, p.Y}
			}
			line := orb.LineString(points)
			f := geojson.NewFeature(line)
			f.Properties = props
			fc.Append(f)
		default:
			continue
		}
	}

	outBytes, _ := json.MarshalIndent(fc, "", "  ")
	outPath := filepath.Join(outputDir, outputFile)
	os.WriteFile(outPath, outBytes, 0644)
	fmt.Printf("✅ Saved %s (%d features)\n", outPath, len(fc.Features))
}
