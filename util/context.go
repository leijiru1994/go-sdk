package util

import "context"

func UserIDFromCtx(ctx context.Context) uint64 {
	userID := ctx.Value("user_id")
	if userID == nil {
		return 0
	}
	return userID.(uint64)
}
