package cmd

// result holds the outcome of processing a single action file.
type result struct {
	actionPath string
	errs       []string
}
