package gyre

import "fmt"

type Resource map[string]string

func (r Resource) Clone() Resource {
	out := make(Resource, len(r))
	for k, v := range r {
		out[k] = v
	}
	return out
}

// Merge applies layers from weakest to strongest. Explicit conflicting
// service.name values are rejected instead of silently changing identity.
func MergeResource(layers ...Resource) (Resource, error) {
	out := Resource{}
	for _, layer := range layers {
		for k, v := range layer {
			if k == "service.name" && out[k] != "" && v != "" && out[k] != v {
				return nil, fmt.Errorf("gyre: conflicting service.name")
			}
			if v != "" {
				out[k] = v
			}
		}
	}
	return out, nil
}
