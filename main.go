package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/huandu/xstrings"
)

type CodeGen struct {
	TableName              string
	AliasTableNameInSelect string
	ListSearchColumn       string
	Columns                []Column
}

type RefTable struct {
	Name          string
	SelectModel   string
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
			Name:                     "address",
			DataType:                 "string",
			DataTypeLen:              200,
			IsNull:                   true,
			IsIndexed:                false,
			IncludedInSearchDropDown: false,
		},
		/*{
			Name:                     "mobile_country_code",
			DataType:                 "string",
			DataTypeLen:              10,
			IsNull:                   true,
			IncludedInSearchDropDown: true,
		},
		{
			Name:                     "mobile_no",
			DataType:                 "string",
			DataTypeLen:              20,
			IsNull:                   true,
			IncludedInSearchDropDown: true,
		},
		{
			Name:     "created_by",
			DataType: "uuid",
			RefTable: &RefTable{Name: "users", SelectModel: "IdEmailModel", SelectColumns: []string{"id", "email"}},
			IsNull:   false,
		},
		{
			Name:     "created_at",
			DataType: "time",
			IsNull:   false,
		},
		{
			Name:     "updated_by",
			DataType: "uuid",
			RefTable: &RefTable{Name: "users", SelectModel: "IdEmailModel", SelectColumns: []string{"id", "email"}},
			IsNull:   true,
		},
		{
			Name:      "updated_at",
			DataType:  "time",
			IsNull:    true,
			IsIndexed: true,
		},*/
	}

	codeGen := CodeGen{TableName: "organizations", ListSearchColumn: "name", AliasTableNameInSelect: "org", Columns: columns}

	//----------- func
	createColumn := func(column Column) string {

		var dataType = ""
		switch column.DataType {
		case "uuid":
			dataType = "uuid"
		case "time":
			dataType = "TIMESTAMP WITH TIME ZONE"
		case "string":
			dataType = fmt.Sprintf("varchar(%d)", column.DataTypeLen)
		case "number":
			dataType = "integer"
		}

		var isNull = "NULL"
		if !column.IsNull {
			isNull = "NOT NULL"
		}

		var ref = ""
		if column.RefTable != nil {
			ref = fmt.Sprintf("REFERENCES %s(id)", column.RefTable.Name)
		}

		return fmt.Sprintf("\"%s\" %s %s %s", column.Name, dataType, isNull, ref)
	}

	left := func(str string, len int) string {
		return str[0:len]
	}

	notNull := func(n interface{}) bool {
		return n != nil
	}

	sub := func(n1 int, n2 int) int {
		return n1 - n2
	}

	kebabCase := func(str string) string {
		return xstrings.ToKebabCase(str)
	}

	camelCase := func(str string) string {
		return xstrings.ToCamelCase(str)
	}

	pascalCase := func(str string) string {
		return xstrings.ToPascalCase(str)
	}

	createColumnForStruct := func(column Column) string {
		goDataType := columnDataTypeToGoDataType(column)
		return fmt.Sprintf("%s\t%s\t`json:\"%s\" db:\"%s\"`", xstrings.ToPascalCase(column.Name), goDataType, xstrings.ToCamelCase(column.Name), column.Name)
	}

	createColumnForStructInputDTO := func(column Column) string {
		//Name              string                  `json:"name" validate:"required"`
		goDataType := columnDataTypeToGoDataType(column)

		validate := ""
		if !column.IsNull {
			validate = " validate:\"required\""
		}

		return fmt.Sprintf("%s\t%s\t`json:\"%s\"%s`", xstrings.ToPascalCase(column.Name), goDataType, xstrings.ToCamelCase(column.Name), validate)
	}

	generateSelectForRefColumn := func(columns []Column) string {
		//"uCreatedBy"."id" AS "createdBy.id",
		//"uCreatedBy"."email" AS "createdBy.email",

		selectColumns := []string{}
		for _, column := range columns {
			if column.RefTable != nil {

				for _, selCol := range column.RefTable.SelectColumns {
					sel := fmt.Sprintf("%s%s.\"%s\" AS \"%s.%s\"", column.RefTable.Name[0:1], xstrings.ToPascalCase(column.Name), selCol, xstrings.ToCamelCase(column.Name), selCol)
					selectColumns = append(selectColumns, sel)
				}
			}
		}
		selectColumnsWithComma := strings.Join(selectColumns, ",\n")
		return selectColumnsWithComma
	}

	joinInSelect := func(cg CodeGen) string {
		/*
			LEFT JOIN "users" "uPerson" ON "uPerson"."person_id" = "p"."id"
			INNER JOIN "users" "uCreatedBy" ON "uCreatedBy"."id" = "p"."created_by"
			LEFT JOIN "users" "uUpdatedBy" ON "uUpdatedBy"."id" = "p"."updated_by"
		*/
		columns = cg.Columns

		joins := []string{}
		for _, column := range columns {
			if column.RefTable != nil {
				joinType := "INNER JOIN"
				if column.IsNull {
					joinType = "LEFT JOIN"
				}

				alias := fmt.Sprintf("%s%s", column.RefTable.Name[0:1], xstrings.ToPascalCase(column.Name))

				join := fmt.Sprintf("%s \"%s\" %s ON %s.id = %s.\"%s\"", joinType, column.RefTable.Name, alias, alias, cg.AliasTableNameInSelect, column.Name)

				joins = append(joins, join)
			}
		}

		return strings.Join(joins, "\n")
	}
	//----------- func - end

	//read template folder
	templateFiles, err := os.ReadDir("./tmpl/")
	if err != nil {
		log.Fatal(err)
	}

	for _, templateFile := range templateFiles {
		templateFileName := path.Join("./tmpl", templateFile.Name())

		baseName := path.Base(templateFileName)
		tmpl, err := template.New(baseName).Funcs(template.FuncMap{
			"left":                          left,
			"notNull":                       notNull,
			"createColumn":                  createColumn,
			"sub":                           sub,
			"camelCase":                     camelCase,
			"pascalCase":                    pascalCase,
			"createColumnForStruct":         createColumnForStruct,
			"generateSelectForRefColumn":    generateSelectForRefColumn,
			"joinInSelect":                  joinInSelect,
			"createColumnForStructInputDTO": createColumnForStructInputDTO,
			"kebabCase":                     kebabCase,
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

func columnDataTypeToGoDataType(column Column) string {
	goDataType := ""
	switch column.DataType {
	case "uuid":
		goDataType = "string"
	case "time":
		goDataType = "time.Time"
	case "string":
		goDataType = "string"
	case "number":
		goDataType = "int"
		if column.DataTypeLen >= 0 {
			goDataType = fmt.Sprintf("int%d", column.DataTypeLen)
		}
	}

	if column.RefTable != nil {
		goDataType = column.RefTable.SelectModel
	}

	nullIndicator := ""
	if column.IsNull {
		nullIndicator = "*"
	}

	return fmt.Sprintf("%s%s", nullIndicator, goDataType)
}
