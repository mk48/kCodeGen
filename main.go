package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"text/template"

	"github.com/huandu/xstrings"
)

type CodeGen struct {
	TableName string
	Columns   []Column
}

type Column struct {
	Name        string
	DataType    string //uuid, string, time,
	DataTypeLen int    // varchar(*)
	RefTable    string
	IsNull      bool

	IsIndexed bool
}

func main() {
	columns := []Column{
		{
			Name:      "user_id",
			DataType:  "uuid",
			RefTable:  "users",
			IsNull:    false,
			IsIndexed: true,
		},
		{
			Name:        "role",
			DataType:    "string",
			DataTypeLen: 20,
			IsNull:      false,
		},
		{
			Name:     "created_by",
			DataType: "uuid",
			RefTable: "users",
			IsNull:   false,
		},
		{
			Name:     "created_at",
			DataType: "time",
			IsNull:   false,
		},
		{
			Name:     "deleted_by",
			DataType: "uuid",
			RefTable: "users",
			IsNull:   true,
		},
		{
			Name:      "deleted_at",
			DataType:  "time",
			IsNull:    true,
			IsIndexed: true,
		},
	}

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
		if column.RefTable != "" {
			ref = fmt.Sprintf("REFERENCES %s(id)", column.RefTable)
		}

		return fmt.Sprintf("%s %s %s %s", column.Name, dataType, isNull, ref)
	}

	sub := func(n1 int, n2 int) int {
		return n1 - n2
	}

	camelCase := func(str string) string {
		return xstrings.ToCamelCase(str)
	}
	//----------- func - end

	codeGen := CodeGen{TableName: "users_role", Columns: columns}

	//read template folder
	templateFiles, err := os.ReadDir("./tmpl/")
	if err != nil {
		log.Fatal(err)
	}

	for _, templateFile := range templateFiles {
		templateFileName := path.Join("./tmpl", templateFile.Name())

		baseName := path.Base(templateFileName)
		tmpl, err := template.New(baseName).Funcs(template.FuncMap{
			"createColumn": createColumn,
			"sub":          sub,
			"camelCase":    camelCase,
		}).ParseFiles(templateFileName)
		if err != nil {
			panic(err)
		}

		/*f, err := os.Create("./out.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()*/

		err = tmpl.Execute(os.Stdout, codeGen)
		if err != nil {
			panic(err)
		}
	}
} // end main
