package restapi

import "time"

// nowUTC is a function variable with a default implementation that returns
// the current time in UTC.
//
// The implementation may be replaced to provide a consistent time value for
// testing.
var nowUTC = func() time.Time { return time.Now().UTC() }
