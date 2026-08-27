package main

// $rPg(Services/Cache/Go cache resolution, cache, go)
// $~ Explains the cache fallback used by this tiny Go service.
func resolve(hit bool) string { if hit { return "cached" }; return "origin" }
