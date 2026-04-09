package util

func ListStringToAny(in []string) (out []any) {
	out = make([]any, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}
func ListIntToAny(in []int) (out []any) {
	out = make([]any, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}
func ListInt64ToAny(in []int64) (out []any) {
	out = make([]any, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}
