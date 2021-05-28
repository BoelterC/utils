package tool

func Filter(ss []Type, test func(Type) bool) (ret []Type) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	if ret == nil {
		ret = []Type{}
	}
	return
}
