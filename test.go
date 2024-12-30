package main

import (
	"fmt"

	"github.com/holiman/uint256"
)

func main() {
	var i, _ = uint256.FromDecimal("115792089237316195423570985008687907853269984665640564038663617003944771332748")

	fmt.Println(i)
	fmt.Println(i.Sign())
	fmt.Println(uint256.NewInt(0).Neg(i))
	fmt.Println(fmt.Sprintf("-%d", uint256.NewInt(0).Neg(i)))
}
