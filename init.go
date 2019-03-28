// +build go1.12

package ocmod

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	PathKey, _    = tag.NewKey("path")
	VersionKey, _ = tag.NewKey("version")
	ProgramKey, _ = tag.NewKey("program")
)

var GoModInfo = stats.Int64("go_mod_info", "Go Module information", stats.UnitDimensionless)

var (
	GoModInfoView = &view.View{
		Measure:     GoModInfo,
		TagKeys:     []tag.Key{PathKey, VersionKey, ProgramKey},
		Aggregation: view.Count(),
	}
)

// build module dependency information. Populated at build-time.
var (
	buildInfo, ok = debug.ReadBuildInfo()
	info          string
	version       map[string]string
)

func init() {
	var versions []string
	if ok {
		for _, dep := range buildInfo.Deps {
			d := dep
			if dep.Replace != nil {
				d = dep.Replace
			}
			versions = append(versions, d.Path+": "+d.Version)
		}
	}

	info = fmt.Sprintf("(%s)", strings.Join(versions, ", "))

	version = make(map[string]string)
	if ok {
		for _, dep := range buildInfo.Deps {
			d := dep
			if dep.Replace != nil {
				d = dep.Replace
			}
			version[d.Path] = d.Version
		}
	}

	for _, dep := range buildInfo.Deps {
		d := dep
		if dep.Replace != nil {
			d = dep.Replace
		}
		stats.RecordWithTags(context.Background(),
			[]tag.Mutator{
				tag.Upsert(PathKey, d.Path),
				tag.Upsert(VersionKey, d.Version),
			},
			GoModInfo.M(1),
		)
	}
}
