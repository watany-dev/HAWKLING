# Performance Improvement Review

## Current Issues

1. Unnecessary duplicate filtering in `main.go`
2. Inefficient use of AWS API requests
3. Inefficient slice initialization methods
4. Insufficient memory usage optimization in formatter
5. Filtering flags (--used, --unused, --days) not functioning correctly

## Improvement Suggestions

### 1. Optimize Filtering Logic (main.go)

Currently, in the `listRoles` function, multiple filter conditions (days, onlyUsed, onlyUnused) are applied independently, resulting in multiple slice traversals. This can be optimized to complete with a single traversal.

```go
// Before improvement: Multiple filtering loops
if days > 0 { /* filtering process */ }
if onlyUsed { /* filtering process */ }
if onlyUnused { /* filtering process */ }

// After improvement: Evaluate all conditions in one loop
var filteredRoles []aws.Role
for _, role := range roles {
    if (days > 0 && !role.IsUnused(days)) ||
       (onlyUsed && role.LastUsed == nil) ||
       (onlyUnused && role.LastUsed != nil) {
        continue
    }
    filteredRoles = append(filteredRoles, role)
}
```

### 2. Fix Filtering Flags (list.go)

There are issues with the filtering logic in the current implementation:

```go
// Issue: Filtering logic is reversed
if (c.options.OnlyUsed && !isUnused) || (c.options.OnlyUnused && isUnused) {
    filteredRoles = append(filteredRoles, role)
}
```

In this implementation:
1. When the `OnlyUsed` flag is set, it filters roles that are `!isUnused` (i.e., in use), but `isUnused` is incorrectly determined.
2. When the `OnlyUnused` flag is set, it filters roles that are `isUnused` (unused), but the consideration of days is insufficient.

Fix proposal:
```go
// Fix proposal: Clarify filtering logic
for _, role := range roles {
    // Properly handle the days > 0 condition
    isUnusedForDays := role.IsUnused(c.options.Days)
    
    // OnlyUsed: Show only roles that have been used (LastUsed != nil)
    if c.options.OnlyUsed && (role.LastUsed == nil) {
        continue
    }
    
    // OnlyUnused: Show only roles that have not been used within the specified days
    if c.options.OnlyUnused && !isUnusedForDays {
        continue
    }
    
    filteredRoles = append(filteredRoles, role)
}
```

### 3. Optimize AWS API Requests (awsclient.go)

In the `ListRoles` method, currently the retrieval of roles and the retrieval of the last used date for roles are separated. There may be potential to reduce API calls by maximizing the use of information provided by the AWS API.

Also, improving how goroutines are launched can reduce overhead.

```go
// Before improvement: Overhead of launching goroutines
for i := range roles {
    go func(i int) {
        // ...
    }(i)
}

// After improvement: Adopt worker pool pattern
```

### 4. Fix IsUnused Logic (iam.go)

The logic in the current `IsUnused` method also needs to be checked:

```go
// IsUnused checks if a role is unused for the specified number of days
func (r *Role) IsUnused(days int) bool {
    if r.LastUsed == nil {
        return true  // If LastUsed is nil, the role is unused
    }

    threshold := time.Now().AddDate(0, 0, -days)
    return r.LastUsed.Before(threshold)  // If it was last used before 'days' days ago, consider it unused
}
```

In this logic, if `LastUsed` is `nil` (never used), it is always treated as "unused". Also, the comparison with the threshold determines that a role is "unused" if the last used date is before (older than) the threshold. This is logically correct.

### 5. Optimize Slice Initialization

In many places, slices are initialized without specifying a pre-capacity. By specifying capacity in advance, memory allocations due to dynamic expansion can be reduced.

```go
// Before improvement
var unusedRoles []aws.Role

// After improvement
unusedRoles := make([]aws.Role, 0, len(roles))
```

### 6. Optimize Memory Usage in formatter.go

In the `FormatRolesAsJSON` function, all role information is stored in memory as JSON before being output. For cases with many roles, changing to a streaming output method can reduce memory usage.

## Implementation Priorities

1. **Highest**: Fix filtering flags (resolve functional issues)
   - ✅ Implementation complete: Modified filtering logic so that --used, --unused, --days flags function correctly
   - ✅ Tests added: Added tests for filtering logic
2. **High**: Optimize slice initialization (easy to implement and clear effect)
   - ✅ Partially implemented: Optimized initialization of filteredRoles slice
3. **Medium**: Integrate filtering logic (improve while maintaining code readability)
   - ✅ Implementation complete: Optimized filtering logic to complete in a single loop
4. **Medium**: Optimize goroutine management (currently has semaphores, so basically no problem)
5. **Low**: Stream output in JSON formatter (only effective with large amounts of data)

These improvements are expected to reduce memory usage and improve CPU usage efficiency. The effects will be particularly noticeable when processing large numbers of roles.