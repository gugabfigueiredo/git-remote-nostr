package main

import "strings"

// lineFields splits a line into fields and tries to assign them to the provided pointers.
// If there are more fields than pointers, the last pointer will receive the remaining fields.
// If there are more pointers than fields, the remaining pointers will be emptied.
func lineFields(line string, fields ...*string) {
	parts := strings.Fields(line)

	for i, field := range fields {
		if i >= len(parts) {
			*field = ""
			continue
		}
		*field = parts[i]
		if i == len(fields)-1 {
			*field = strings.Join(parts[i:], " ")
		}
	}
}
