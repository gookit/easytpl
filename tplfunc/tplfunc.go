// Package tplfunc provides a default FuncMap for use with text/template and html/template.
//
//   - string functions: join, trim, trimLeft, trimRight, trimPrefix, trimSuffix, trimSpace, repeat, replace, replaceAll, toUpper, toLower, title, toTitle, toCamel, toSnake, toKebab, toLowerCamel, toLowerSnake, toLowerKebab, toCamelLower, toSnakeLower, toKebabLower, toLowerCamelLower, toLowerSnakeLower, toLowerKebabLower, toCamelUpper, toSnakeUpper, toKebabUpper, toLowerCamelUpper, toLowerSnakeUpper, toLowerKebabUpper, toCamelTitle, toSnakeTitle, toKebabTitle
//   - math functions: add, max, mul, min, sub, div, mod, ceil, floor, round, roundEven, trunc, sqrt, pow, rand, randInt, randIntRange, randFloat, randFloatRange, randPerm, randShuffle, randChoice, randChoices, randSample, randSamples, randString, randStrings, randBytes, randByte, randRune, randRunes, randTime, randDate, randDateTime, randTimeDuration, randDateDuration, randDateTimeDuration, randTimeUnix, randDateUnix, randDateTimeUnix
//   - list functions: list, first, last, len, reverse, sort, shuffle, unique, contains, in, has, keys, values, chunk, chunkBy, chunkByNum, chunkBySize
//   - encoding functions: b64enc, b64dec, b32enc, b32dec
//   - path functions: base, dir, ext, clean, isAbs, osBase, osDir, osExt, osClean, osIsAbs
//   - hash functions: uuid, md5, sha1, sha256, sha512, crc32, crc64
//   - other functions: default, empty, coalesce, fromJson, toJson, toPrettyJson, toRawJson, dump
//
// Example:
//
//		import (
//	 	"github.com/gookit/easytpl/tplfunc"
//	 	"html/template"
//		)
//
//		// This example illustrates that the FuncMap *must* be set before the
//		// templates themselves are loaded.
//		tpl := template.Must(
//	 	template.New("base").Funcs(tplfunc.FuncMap()).ParseGlob("path/to/*.tpl")
//		)
//
// refer: https://github.com/Masterminds/sprig
package tplfunc

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gookit/goutil/strutil"
)

// FuncMap returns the default FuncMap.
func FuncMap() template.FuncMap {
	return simpleMergeMultiMap(stdFuncMap) // TODO add more funcs
}

// StdFuncMap returns the default template func map.
func StdFuncMap() template.FuncMap {
	return stdFuncMap
}

// stdFuncMap is the default template func map.
var stdFuncMap = map[string]any{
	// String:
	"join":  strings.Join,
	"trim":  strings.TrimSpace,
	"upper": strings.ToUpper,
	"lower": strings.ToLower,
	// - uppercase first char
	"ucFirst": strutil.UpFirst,
	"loFirst": strutil.LowerFirst,

	// OS:
	"env":       os.Getenv,
	"expandenv": os.ExpandEnv,

	// Paths:
	"base":  path.Base,
	"dir":   path.Dir,
	"clean": path.Clean,
	"ext":   path.Ext,
	"isAbs": path.IsAbs,

	// File paths:
	"osBase":  filepath.Base,
	"osClean": filepath.Clean,
	"osDir":   filepath.Dir,
	"osExt":   filepath.Ext,
	"osIsAbs": filepath.IsAbs,
}

// simpleMergeMultiMap merge multi any map[string]any data.
func simpleMergeMultiMap(mps ...map[string]any) map[string]any {
	newMp := make(map[string]any)
	for _, mp := range mps {
		for k, v := range mp {
			newMp[k] = v
		}
	}
	return newMp
}
