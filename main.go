package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type CodeGen struct {
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
	DataType                 string //uuid, string, time,
	DataTypeLen              int    // varchar(*)
	RefTable                 *RefTable
	IsNull                   bool
	IsIndexed                bool
	IncludedInSearchDropDown bool
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
			Name:                     "dt",
			DataType:                 "time",
			IsNull:                   false,
			IsIndexed:                true,
			IncludedInSearchDropDown: false,
		},
		{
			Name:                     "location_type",
			DataType:                 "string",
			DataTypeLen:              50,
			IsNull:                   true,
			IsIndexed:                false,
			IncludedInSearchDropDown: false,
		},
		{
			Name:                     "address",
			DataType:                 "string",
			DataTypeLen:              200,
			IsNull:                   true,
			IsIndexed:                false,
			IncludedInSearchDropDown: false,
		},
		{
			Name:                     "notes",
			DataType:                 "string",
			DataTypeLen:              200,
			IsNull:                   true,
			IsIndexed:                false,
			IncludedInSearchDropDown: false,
		},
		{
			Name:      "organization_id",
			DataType:  "uuid",
			RefTable:  &RefTable{Name: "organizations", Model: "Organization", SelectColumns: []string{"id", "name"}},
			IsNull:    false,
			IsIndexed: true,
		},
		{
			Name:      "team_id",
			DataType:  "uuid",
			RefTable:  &RefTable{Name: "team", Model: "Team", SelectColumns: []string{"id", "name"}},
			IsNull:    false,
			IsIndexed: true,
		},
	}

	codeGen := CodeGen{TableName: "event", ListSearchColumn: "name", AliasTableNameInSelect: "e", IsHistoryTableNeeded: true, Columns: columns}

	//read template folder
	templateFiles, err := os.ReadDir("./tmpl/")
	if err != nil {
		log.Fatal(err)
	}

	for _, templateFile := range templateFiles {
		templateFileName := path.Join("./tmpl", templateFile.Name())

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

		err = tmpl.Execute(f, codeGen)
		if err != nil {
			panic(err)
		}
	}
} // end main
