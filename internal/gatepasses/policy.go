package gatepasses

import "fmt"

// ResolveReturnability applies the tenant-configured gatepass type policy.
// OPTIONAL: requester may choose; REQUIRED: return cannot be disabled;
// NOT_ALLOWED: return cannot be enabled.
func ResolveReturnability(policy ReturnabilityPolicy, requested *bool, defaultValue bool) (bool, error) {
	switch policy {
	case ReturnabilityRequired:
		if requested != nil && !*requested {
			return false, fmt.Errorf("this gatepass type requires return")
		}
		return true, nil
	case ReturnabilityNotAllowed:
		if requested != nil && *requested {
			return false, fmt.Errorf("this gatepass type does not allow return")
		}
		return false, nil
	case ReturnabilityOptional:
		if requested != nil {
			return *requested, nil
		}
		return defaultValue, nil
	default:
		return false, fmt.Errorf("unsupported returnability policy: %s", policy)
	}
}
