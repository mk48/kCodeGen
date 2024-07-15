package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type CrudGen struct {
	TableName              string
	AliasTableNameInSelect string
	ListSearchColumn       string
	Columns                []Column
	IsHistoryTableNeeded   bool
}

type RefTable struct {
	Name          string
	Model         string
	SelectColumns []string
}

type Column struct {
	Name                     string
	DataType                 string //uuid, string, time, number
	DataTypeLen              int    // varchar(*)
	RefTable                 *RefTable
	IsNull                   bool
	IsIndexed                bool
	IncludedInSearchDropDown bool
}

type ManyToManyLink struct {
	Table1 string
	Table2 string
	Alias  string
}

func main() {
	columns := []Column{
		{
			Name:                     "name",
			DataType:                 "string",
			DataTypeLen:              50,
			IsNull:                   false,
			IsIndexed:                true,
			IncludedInSearchDropDown: true,
		},
		{
			Name:                     "description",
			DataType:                 "string",
			DataTypeLen:              150,
			IsNull:                   true,
			IsIndexed:                false,
			IncludedInSearchDropDown: false,
		},
		{
			Name:                     "type",
			DataType:                 "string",
			DataTypeLen:              50,
			IsNull:                   false,
			IsIndexed:                true,
			IncludedInSearchDropDown: true,
		},
	}
	singleTableCrud := CrudGen{TableName: "tag", ListSearchColumn: "name", AliasTableNameInSelect: "t", IsHistoryTableNeeded: true, Columns: columns}
	fmt.Printf(singleTableCrud.TableName) //just to avoid unused error

	many2many := ManyToManyLink{Table1: "event", Table2: "photo", Alias: "ep"}

	//read template folder
	templateFiles, err := os.ReadDir("./tmplMany2ManyLink/")
	if err != nil {
		log.Fatal(err)
	}

	for _, templateFile := range templateFiles {
		templateFileName := path.Join("./tmplMany2ManyLink", templateFile.Name())

		baseName := path.Base(templateFileName)
		tmpl, err := template.New(baseName).Funcs(template.FuncMap{
			"left":                            left,
			"notNull":                         notNull,
			"isNull":                          isNull,
			"createColumn":                    createColumn,
			"sub":                             sub,
			"camelCase":                       camelCase,
			"pascalCase":                      pascalCase,
			"createColumnForStruct":           createColumnForStruct,
			"generateSelectForRefColumn":      generateSelectForRefColumn,
			"joinInSelect":                    joinInSelect,
			"createColumnForStructInputDTO":   createColumnForStructInputDTO,
			"kebabCase":                       kebabCase,
			"sqlDataTypeToGoDataType":         sqlDataTypeToGoDataType,
			"sqlDataTypeToGoDataTypeForInput": sqlDataTypeToGoDataTypeForInput,
			//"StringsJoin": strings.Join,
		}).ParseFiles(templateFileName)
		if err != nil {
			panic(err)
		}

		fileExtension := filepath.Ext(templateFile.Name())
		fileNameWithoutExtension := strings.TrimSuffix(templateFile.Name(), fileExtension)
		f, err := os.Create("./out/" + fileNameWithoutExtension + ".go")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		err = tmpl.Execute(f, many2many)
		if err != nil {
			panic(err)
		}
	}
} // end main
