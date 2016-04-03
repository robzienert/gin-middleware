package oauth

func stringInSlice(val string, list []string) bool {
	for _, s := range list {
		if s == val {
			return true
		}
	}
	return false
}

