package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/simplify"
)

const (
	inputPath  = "data/terra/earth/countries.geojson"
	outputPath = "data/terra/earth/countries_simplified.geojson"
	tolerance  = 0.1 // degrees; smaller keeps more detail
)

func main() {
	fmt.Println("🗺️ Simplifying", inputPath)
	data, err := os.ReadFile(inputPath)
	if err != nil {
		panic(err)
	}

	fc, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		panic(err)
	}

	simplified := geojson.NewFeatureCollection()
	simplifier := simplify.DouglasPeucker(tolerance)

	for _, f := range fc.Features {
		switch g := f.Geometry.(type) {
		case orb.LineString:
			f.Geometry = simplifier.Simplify(g)
		case orb.MultiLineString:
			var newMulti orb.MultiLineString
			for _, line := range g {
				newMulti = append(newMulti, simplifier.Simplify(line).(orb.LineString))
			}
			f.Geometry = newMulti
		case orb.Polygon:
			var newPoly orb.Polygon
			for _, ring := range g {
				newPoly = append(newPoly, simplifier.Simplify(ring).(orb.Ring))
			}
			f.Geometry = newPoly
		case orb.MultiPolygon:
			var newMulti orb.MultiPolygon
			for _, poly := range g {
				var newPoly orb.Polygon
				for _, ring := range poly {
					newPoly = append(newPoly, simplifier.Simplify(ring).(orb.Ring))
				}
				newMulti = append(newMulti, newPoly)
			}
			f.Geometry = newMulti
		default:
			// leave Points etc. untouched
		}
		simplified.Append(f)
	}

	out, _ := json.MarshalIndent(simplified, "", "  ")
	os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err := os.WriteFile(outputPath, out, 0644); err != nil {
		panic(err)
	}
	fmt.Printf("✅ Simplified saved to %s (%d features)\n", outputPath, len(simplified.Features))
}
