package main

import (
    "fmt"
    "io"
    "os"
    "go/token"
    goAst "go/ast"
    goParser "go/parser"
)

func main() {
    reader, err := os.Open("exampleSyntax.txt")
    defer reader.Close()

    if err != nil {
        //TODO: better error handling
        fmt.Println("File open error: " + err.Error())
        return
    }

    //outputLexerResults(reader)
    outputParserResults(reader)
    //fmt.Println(getExampleAst().String())
}

func getExampleAst() *Container {
    return &Container{
        []Import{
            Import{"pkg", "my/pkg"},
            Import{"other", "my/other"},
        },
        []Reg{
            Reg{
                []string{},
                Resolver{
                    "pkg",
                    "NewThing",
                    []string{"_", "Named"},
                },
                false,
            },
            Reg{
                []string{"Named", "OtherName"},
                Resolver{
                    "pkg",
                   "NewOtherThing",
                    []string{"Named", "OtherName"},
                },
                false,
            },
            Reg{
                []string{"Named", "OtherName"},
                Resolver{
                    "pkg",
                    "NewOtherThing",
                    []string{},
                },
                false,
            },
        },
        []List{
            List{
                []string{"MyList"},
                []string{"Named", "OtherName"},
            },
        },
        Resolver{
            "pkg",
            "NewCompositionRoot",
            []string{"MyList"},
        },
    }
}

func outputParserResults(reader io.Reader) {
    container, err := Parse(NewLexer(reader))
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(container.String())
    }
}

func draftParseGo() {
    fileSet := token.NewFileSet()
    t, err := goParser.ParseFile(fileSet, "main.go", nil, goParser.ParseComments)
    if err != nil {
        //TODO: better error handling
        fmt.Println("Parse error: " + err.Error())
        return
    }

    goAst.Inspect(t, func(n goAst.Node) bool {
        switch x := n.(type) {
        case *goAst.FuncDecl:
            fmt.Println(x.Name.Name)
        }
        return true
    })
}

