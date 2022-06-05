package ctf

import (
	rand2 "crypto/rand"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"strings"
)

var RandomString = strings.Split("a b c d e f g h i j k l m n o p q r s t u v w x y z A B C D E F G H I J K L M N O P Q R S T U V W X Y Z 1 2 3 4 5 6 7 8 9", " ")
var Random *rand.Rand

func init() {
	num, err := rand2.Int(rand2.Reader, big.NewInt(9223372036854775806))
	if err != nil {
		log.Fatalf("Error generating seed for random: %v\n", err)
	}

	Random = rand.New(rand.NewSource(num.Int64()))
}

// randomKey generates a new random string using the constraints passed in.
// If numbersOnly is false, the combinations of this is 64^length.
func randomKey(length int, numbersOnly bool) string {
	var key string
	if !numbersOnly {
		for i := 0; i < length; i++ {
			key += RandomString[Random.Int()%len(RandomString)]
		}
	} else {
		for i := 0; i < length; i++ {
			val := Random.Int() % 10
			key += fmt.Sprintf("%v", val)
		}
	}

	return key
}
