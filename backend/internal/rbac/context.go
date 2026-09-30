package rbac

import "context"

type decisionKey struct{}

func WithDecision(ctx context.Context, d Decision) context.Context {
    return context.WithValue(ctx, decisionKey{}, d)
}

func DecisionFromContext(ctx context.Context) (Decision, bool) {
    d, ok := ctx.Value(decisionKey{}).(Decision)
    return d, ok
}
