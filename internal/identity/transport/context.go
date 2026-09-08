package transport

import (
	"context"
	"github.com/ksamwang/PersonalContentPlatform/internal/identity/domain"
)

type principalKey struct{}

func withPrincipal(ctx context.Context, p domain.Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}
func Principal(ctx context.Context) (domain.Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(domain.Principal)
	return p, ok
}
