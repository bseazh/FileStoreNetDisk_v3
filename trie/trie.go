package trie

import (
	"fmt"
	"unicode"
)

const N = 1e5
type UserFile struct{
	Username string
	Son [N][ 36 ] int
	Cnt [N] bool
	Idx int
}

func (user UserFile) init( username string ){
	user.Username = username
}

func (user UserFile)insert ( str string ){
	p := 0
	strLen := len(str)
	for i := 0 ; i < strLen ; i ++ {
		var u int
		if unicode.IsLower(rune(str[i])) {
			u = int(rune(str[i] - 'a')) + 10
		}else{
			u = int(rune(str[i] - '0'))
		}
		if user.Son[p][u] == 0 {
			user.Idx = user.Idx + 1
			user.Son[p][u] = user.Idx
		}
		p = user.Son[p][u]
		fmt.Printf("%d\n",u)
	}
	user.Cnt[p] = true
}

func (user UserFile)query ( str string ) bool {
	p := 0
	strLen := len(str)
	for i := 0 ; i < strLen ; i ++ {
		var u int
		if unicode.IsLower(rune(str[i])) {
			u = int(rune(str[i] - 'a')) + 10
		}else{
			u = int(rune(str[i] - '0'))
		}
		if user.Son[p][u] == 0 {
			return false
		}
		p = user.Son[p][u]
	}
	return true
}