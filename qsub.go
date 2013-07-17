package grideng

import "bytes"

// Turns a map of resources into a string.
func ResourcesString(res map[string]string) string {
	var b bytes.Buffer
	var i int
	for k, v := range res {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(v)
		i++
	}
	return b.String()
}
