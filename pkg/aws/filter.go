package aws

// FilterOptions contains various filtering criteria
type FilterOptions struct {
	Days       int
	OnlyUsed   bool
	OnlyUnused bool
}

// FilterRoles filters roles based on specified options
// This unified implementation follows these rules:
// - Days=0: No days-based filtering
// - OnlyUsed=true: Show only roles that have been used at least once
// - OnlyUnused=true: Show only roles that have never been used
// - Days>0: Show roles not used in the specified days
// - Days>0 + OnlyUsed: Show roles that have been used at least once but not in the specified days
// - Days>0 + OnlyUnused: Show roles that have never been used (days parameter doesn't affect these)
func FilterRoles(roles []Role, options FilterOptions) []Role {
	// If both filters are enabled, return empty list (logical conflict)
	if options.OnlyUsed && options.OnlyUnused {
		return []Role{}
	}

	filteredRoles := make([]Role, 0, len(roles))

	for _, role := range roles {
		// OnlyUnused: Include only roles that have never been used (LastUsed == nil)
		if options.OnlyUnused && role.LastUsed != nil {
			continue
		}

		// OnlyUsed: Include only roles that have been used at least once (LastUsed != nil)
		if options.OnlyUsed && role.LastUsed == nil {
			continue
		}

		// Days filter: For roles not matching OnlyUnused, apply the days filter
		if options.Days > 0 && !options.OnlyUnused && !role.IsUnused(options.Days) {
			continue
		}

		filteredRoles = append(filteredRoles, role)
	}

	return filteredRoles
}
