// +build gofuzz

package runtimeutil

// Fuzz example
func Fuzz(data []byte) int {
	// img, err := png.Decode(bytes.NewReader(data))
	// if err != nil {
	// 	if img != nil {
	// 		panic("img != nil on error")
	// 	}
	// 	return 0
	// }
	// var w bytes.Buffer
	// err = png.Encode(&w, img)
	// if err != nil {
	// 	panic(err)
	// }
	return 1
}
