package navigation

import "time"

// Response is the tenant navigation contract consumed by web and mobile clients.
// Visibility is resolved on the server from the caller's current permissions.
type Response struct {
	Items       []Item    `json:"items"`
	GeneratedAt time.Time `json:"generated_at"`
}

// Item is a safe server-defined navigation item. Route and permission rules are
// never supplied by the client.
type Item struct {
	Code          string   `json:"code"`
	Label         string   `json:"label"`
	Icon          string   `json:"icon"`
	Href          string   `json:"href"`
	MatchPrefixes []string `json:"match_prefixes"`
	Placement     string   `json:"placement"`
	Order         int      `json:"order"`
}

// Definition belongs to the server-side registry.
type Definition struct {
	Item
	AnyPermissions []string
}
