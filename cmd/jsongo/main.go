package main

import (
	"fmt"
	"github.com/qwenode/gogo/ff"
	"github.com/qwenode/gogo/ss"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/twpayne/go-jsonstruct/v3"
)

var (
	abbreviations            = ""                                       // 类似 URL,ID,则会将此类转为大写
	omitEmptyTags            = "auto"                                   //
	packageComment           = "Code generated by JSONGO, DO NOT EDIT." // 包注释
	packageName              = ""
	skipUnparsableProperties = true
	stringTags               = false // 给tag 加上 ,string,没用
	structTagName            = ""    // tag名字
	typeComment              = ""    // 类注释
	typeName                 = ""    // 类名
	intType                  = ""    // 设置 int64,则所有int都为int64,默认int
	useJSONNumber            = false
	goFormat                 = true

	omitEmptyTagsType = map[string]jsonstruct.OmitEmptyTagsType{
		"never":  jsonstruct.OmitEmptyTagsNever,
		"always": jsonstruct.OmitEmptyTagsAlways,
		"auto":   jsonstruct.OmitEmptyTagsAuto,
	}
)

type SpecFile struct {
	Path string
	Name string
	Type string
}

func run() error {
	args := os.Args
	dir, _ := os.Getwd()
	specFiles := []SpecFile{}
	filepath.WalkDir(
		dir, func(path string, d fs.DirEntry, err error) error {
			lower := strings.ToLower(path)
			sf := SpecFile{
				Path: path,
				Name: upperFirst(ff.GetFileNameWithoutExtension(path)),
				Type: "",
			}
			if strings.HasSuffix(lower, ".json") {
				sf.Type = "json"
				specFiles = append(specFiles, sf)
			}
			if strings.HasSuffix(lower, ".yaml") {
				sf.Type = "yaml"
				specFiles = append(specFiles, sf)
			}
			return err
		},
	)
	if len(specFiles) <= 0 {
		log.Fatal().Msg("no spec files found")
		return nil
	}
	packageName = filepath.Base(dir)
	for _, arg := range args {
		if strings.Contains(arg, "-int=") {
			intType = ss.GetLastElemBySep(arg, "=")
		}
		if strings.Contains(arg, "-abbr=") {
			abbreviations = ss.GetLastElemBySep(arg, "=")
		}
	}
	for _, file := range specFiles {
		l := log.With().Str("Name", file.Name).Logger()
		l.Info().Msg("Parsing")
		options := []jsonstruct.GeneratorOption{
			jsonstruct.WithOmitEmptyTags(jsonstruct.OmitEmptyTagsAuto),
			jsonstruct.WithSkipUnparsableProperties(true),
			jsonstruct.WithStringTags(false),
			jsonstruct.WithUseJSONNumber(false),
			jsonstruct.WithGoFormat(true),
			jsonstruct.WithPackageComment(packageComment),
			jsonstruct.WithPackageName(packageName),
			jsonstruct.WithTypeName(file.Name),
		}
		if abbreviations != "" {
			options = append(options, jsonstruct.WithAbbreviations(strings.Split(abbreviations, ",")...))
		}
		if intType != "" {
			options = append(options, jsonstruct.WithIntType(intType))
		}
		saveTo := filepath.Join(dir, strings.ToLower(file.Name)+".json.go")
		generator := jsonstruct.NewGenerator(options...)
		if file.Type == "yaml" {
			options = append(options, jsonstruct.WithStructTagNames([]string{"json", "yaml"}))
			err := generator.ObserveYAMLFile(file.Path)
			if err != nil {
				log.Err(err).Msg("yaml observe failed")
				continue
			}
		} else {
			err := generator.ObserveJSONFile(file.Path)
			if err != nil {
				log.Err(err).Msg("json observe failed")
				continue
			}
		}
		bytes, err := generator.Generate()
		if err != nil {
			log.Err(err).Msg("generate failed")
			continue
		}
		l.Info().Msg("writing...")
		ff.PutContents(saveTo, bytes)
		l.Info().Msg("successfully generated")
	}
	return nil
}
func upperFirst(s string) string {
	return strings.Trim(cases.Title(language.English, cases.NoLower).String(s), "_")
}

//go:generate jsongo -int=int64
func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
