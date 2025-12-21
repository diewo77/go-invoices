package gate

import "context"

// Policy defines authorization rules for a resource type.
// U is the user/subject type (e.g., uint for userID, *User for full user struct).
// Implementations check whether user may perform action on resource.
type Policy[U any] interface {
	// Can returns true if user is authorized to perform action on resource.
	// For list/create, resource may be nil (context-only check).
	Can(ctx context.Context, user U, action Action, resource any) bool
}
