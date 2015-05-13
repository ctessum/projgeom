/*
Package projgeom is performs geodesic reprojections on
Open GIS Consortium style geometry objects.
It is an interface between
	"github.com/pebbe/go-proj-4/proj"
and
	"github.com/twpayne/gogeom/geom"
*/
package projgeom

import (
	"io"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/lukeroth/gdal"
	"github.com/pebbe/go-proj-4/proj"
	"github.com/twpayne/gogeom/geom"
)

type UnsupportedGeometryError struct {
	Type reflect.Type
}

func (e UnsupportedGeometryError) Error() string {
	return "projgeom: unsupported geometry type: " + e.Type.String()
}

// Project geometry from src to dst projection. inputDegrees and outputDegrees are `true` if
// the input or output geometries is in units of degrees. We need to know this
// because the Proj4 library works in units of radians.
// Because I don't know whether to transform Z values from degrees to radians or
// not, Z values are not supported.
// I also don't know what to do with M values so they are not supported either.
func project(g geom.T, src, dst *proj.Proj, inputDegrees, outputDegrees bool) (geom.T, error) {
	if g == nil {
		return nil, nil
	}
	switch g.(type) {
	case geom.Point:
		point := g.(geom.Point)
		return projectPoint(&point, src, dst, inputDegrees, outputDegrees)
	//case geom.PointZ:
	//	pointZ := g.(geom.PointZ)
	//	return projectPointZ(&pointZ, src, dst)
	//case geom.PointM:
	//	pointM := g.(geom.PointM)
	//	return projectPointM(&pointM, src, dst)
	//case geom.PointZM:
	//	pointZM := g.(geom.PointZM)
	//	return projectPointZM(&pointZM, src, dst)
	case geom.LineString:
		lineString := g.(geom.LineString)
		return projectLineString(&lineString, src, dst, inputDegrees,
			outputDegrees)
	//case geom.LineStringZ:
	//	lineStringZ := g.(geom.LineStringZ)
	//	return projectLineStringZ(&lineStringZ, src, dst)
	//case geom.LineStringM:
	//	lineStringM := g.(geom.LineStringM)
	//	return projectLineStringM(&lineStringM, src, dst)
	//case geom.LineStringZM:
	//	lineStringZM := g.(geom.LineStringZM)
	//	return projectLineStringZM(&lineStringZM, src, dst)
	case geom.MultiLineString:
		multiLineString := g.(geom.MultiLineString)
		return projectMultiLineString(&multiLineString, src, dst,
			inputDegrees, outputDegrees)
	case geom.Polygon:
		polygon := g.(geom.Polygon)
		return projectPolygon(&polygon, src, dst,
			inputDegrees, outputDegrees)
	//case geom.PolygonZ:
	//	polygonZ := g.(geom.PolygonZ)
	//	return projectPolygonZ(&polygonZ, src, dst)
	//case geom.PolygonM:
	//	polygonM := g.(geom.PolygonM)
	//	return projectPolygonM(&polygonM, src, dst)
	//case geom.PolygonZM:
	//	polygonZM := g.(geom.PolygonZM)
	//	return projectPolygonZM(&polygonZM, src, dst), nil
	case geom.MultiPolygon:
		multiPolygon := g.(geom.MultiPolygon)
		return projectMultiPolygon(&multiPolygon, src, dst,
			inputDegrees, outputDegrees)
	default:
		return nil, &UnsupportedGeometryError{reflect.TypeOf(g)}
	}
}

type CoordinateTransform struct {
	src, dst                    *proj.Proj
	sameProj                    bool
	inputDegrees, outputDegrees bool
}

func NewCoordinateTransform(src, dst gdal.SpatialReference) (
	ct *CoordinateTransform, err error) {
	ct = new(CoordinateTransform)
	ct.sameProj = src.IsSame(dst)
	var srcproj, dstproj string
	if !ct.sameProj {
		srcproj, err = src.ToProj4()
		if err != nil && err.Error() != "No Error" {
			return
		}
		ct.inputDegrees = strings.Contains(srcproj, "longlat") ||
			strings.Contains(srcproj, "latlong")
		ct.src, err = proj.NewProj(srcproj)
		if err != nil {
			return
		}

		dstproj, err = dst.ToProj4()
		if err != nil && err.Error() != "No Error" {
			return
		}
		ct.outputDegrees = strings.Contains(dstproj, "longlat") ||
			strings.Contains(dstproj, "latlong")
		ct.dst, err = proj.NewProj(dstproj)
		if err != nil {
			return
		}
	}
	return
}

func (ct *CoordinateTransform) Reproject(g geom.T) (geom.T, error) {
	if ct.sameProj {
		return g, nil
	}
	g2, err := project(g, ct.src, ct.dst,
		ct.inputDegrees, ct.outputDegrees)
	return g2, err
}

// ReadPrj reads an ESRI '.prj' projection file and
// creates a corresponding spatial reference.
func ReadPrj(f io.Reader) (gdal.SpatialReference, error) {
	sr := gdal.CreateSpatialReference("")
	prj, err := ioutil.ReadAll(f)
	if err != nil {
		return gdal.SpatialReference{}, err
	}
	err = sr.FromWKT(string(prj))
	if err != nil && err.Error() == "No Error" {
		err = nil
	}
	return sr, err
}
