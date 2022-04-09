package ctf

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var RandomString = strings.Split("a b c d e f g h i j k l m n o p q r s t u v w x y z A B C D E F G H I J K L M N O P Q R S T U V W X Y Z 1 2 3 4 5 6 7 8 9", " ")
var Random *rand.Rand

func init() {
	Random = rand.New(rand.NewSource(time.Now().Unix()))
}

// randomKey generates a new random string using the constraints passed in.
func randomKey(length int, numbersOnly bool) string {
	var key string
	if !numbersOnly {
		for i := 0; i < length; i++ {			
			key += RandomString[Random.Int() % len(RandomString)]	
		}
	} else {
		for i := 0; i < length; i++ {
			val :=  Random.Int() % 10
			key += fmt.Sprintf("%v", val)
		}
	}

	return key
}
