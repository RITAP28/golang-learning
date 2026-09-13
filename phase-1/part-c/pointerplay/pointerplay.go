package pointerplay

type MyInt int
func (input *MyInt) Double() {
	*input *= 2
}
