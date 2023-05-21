package evaluator

import (
	"protiumx.dev/simia/value"
)

var builtins = map[string]*value.Builtin{
	"len":    value.GetBuiltinByName("len"),
	"append": value.GetBuiltinByName("append"),
	"log":    value.GetBuiltinByName("log"),
}
