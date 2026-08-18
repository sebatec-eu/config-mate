package hostsharing

import "fmt"

// ErrShortPath is returned when a path lacks enough components to identify
// a PAC, user, and domain (requires at least 7 path segments for domains,
// 5 for users, or 3 for PAC-only paths).
var ErrShortPath = fmt.Errorf("cannot detect PAC/user/domain from path")

// ErrNoPAC is returned by user.PAC and user.User when no PAC segment was
// found in the parsed path (e.g. a non-Hostsharing dev path like
// /srv/doms/example.com/...).
var ErrNoPAC = fmt.Errorf("no PAC in path")

// ErrNoUser is returned by user.User when the parsed path has no
// Domain-Admin or Email-User sub-account segment.
var ErrNoUser = fmt.Errorf("no user in path")
