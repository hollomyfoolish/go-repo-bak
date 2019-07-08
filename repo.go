package repo

import(
	"github.com/hollomyfoolish/go-repo/p2"
)

func Repo() string {
	return "github.com/hollomyfoolish/go-repo: " + p2.Echo()
}