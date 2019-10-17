package engines

func unique(slice []Result) []Result {
    keys := make(map[string]bool)
    list := []Result{} 
    for _, entry := range slice {
        if _, value := keys[entry.Link]; !value {
            keys[entry.Link] = true
            list = append(list, entry)
        }
    }    
    return list
}

func merge(slices [][]Result) []Result {
	var maxLen = 0
	for _, slice := range slices {
		len := len(slice)
		if len > maxLen {
			maxLen = len
		}
	}
	merged := []Result{}
	for i := 0; i < maxLen; i++ {
		for _, slice := range slices {
			if len(slice) > i {
				merged = append(merged, slice[i])
			}
		}
	}
	return merged
}