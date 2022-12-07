package slice

func Filter(ss []interface{}, test func(interface{}) bool) (ret []interface{}) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	if ret == nil {
		ret = []interface{}{}
	}
	return
}

func IndexOf(list []interface{}, elem interface{}) int {
	for i, v := range list {
		if elem == v {
			return i
		}
	}
	return -1
}

func Split(ls []interface{}, idx int) []interface{} {
	for i := idx; i < len(ls); i++ {
		if i < len(ls)-1 {
			ls[i] = ls[i+1]
		}
	}
	return ls[:len(ls)-1]
}
