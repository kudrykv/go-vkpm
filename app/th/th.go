package th

import (
	"context"
	"runtime/trace"
)

func RegionTask(ctx context.Context, typ string) (context.Context, func()) {
	region := trace.StartRegion(ctx, typ)
	ctx, task := trace.NewTask(ctx, typ)

	return ctx, func() {
		task.End()
		region.End()
	}
}
