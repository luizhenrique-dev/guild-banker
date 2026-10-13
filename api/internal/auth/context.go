package auth

import "context"

type principalKey struct{}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalKey{}).(Principal)
	return principal, ok
}

func ClaimsFromContext(ctx context.Context) (Claims, bool) {
	principal, ok := PrincipalFromContext(ctx)
	return principal.Claims, ok
}
