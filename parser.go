package main

import (
    "fmt"
    "strings"
)

// Here's the grammar for our DSL:
// file ::= import+ root regOrLazyOrList*
// import ::= Import importStuff
// importStuff ::= OpenParen importValue+ CloseParen
//               | importValue
// importValue ::= String
//               | Ident String
// regOrLazyOrList ::= reg | lazy | list
// lazy ::= Lazy Ident regStuff
// reg ::= Reg Ident regStuff
// regStuff ::= resolver
//            | anotherTag* Ident resolver
// list ::= List Ident anotherTag* Ident anotherTag*
// anotherTag ::= Comma Ident
// root ::= Root Ident resolver
// resolver ::= Period Ident Ident*

func Parse(lex Lexer) (c *Container, err error) {
    c = &Container{}

    // Move to the first token.
    lex.Next()

    // Parse the imports (required).
    c.Imports, err = parseImports(lex)
    if err != nil {
        return
    }

    // Parse the root (required).
    c.Root, err = parseRoot(lex)
    if err != nil {
        return
    }

    // Parse reg or lazy or list.
    c.Regs, c.Lists, err = parseAllRegOrLazyOrList(lex)
    if err != nil {
        return
    }

    // Make sure there aren't any more tokens thrown in.
    err = expect(lex, RegTok, LazyTok, ListTok, EOFTok)

    return
}

func parseImports(lex Lexer) ([]Import, error) {
    var imports []Import

    // Get the first import.
    err := expect(lex, ImportTok)
    if err != nil {
        return nil, err
    }

    // Get the remaining imports.
    for lex.Current() == ImportTok {
        imp, err := parseImport(lex)
        if err != nil {
            return nil, err
        }
        imports = append(imports, imp...)
    }

    return imports, nil
}

func parseRoot(lex Lexer) (Resolver, error) {
    if err := expect(lex, RootTok); err != nil {
        return Resolver{}, err
    }
    lex.Next()

    return parseResolver(lex)
}

func parseImport(lex Lexer) ([]Import, error) {
    assert(lex, ImportTok)
    lex.Next()

    var imps []Import
    if lex.Current() == OpenParenTok {
        lex.Next()

        for lex.Current() != CloseParenTok {
            imp, err := parseImportValue(lex)
            if err != nil {
                return nil, err
            }
            imps = append(imps, imp)
        }
        lex.Next()
    } else {
        imp, err := parseImportValue(lex)
        if err != nil {
            return nil, err
        }
        imps = append(imps, imp)
    }

    return imps, nil
}

func parseImportValue(lex Lexer) (imp Import, err error) {
    switch lex.Current() {
        case IdentTok:
            imp.Alias = lex.Value()
            lex.Next()

            err = expect(lex, StringTok)
            if err != nil {
                return
            }
            imp.Package = strings.Trim(lex.Value(), "\"")
            lex.Next()

        case StringTok:
            imp.Package = strings.Trim(lex.Value(), "\"")
            lex.Next()

            parts := strings.Split(imp.Package, "/")
            imp.Alias = parts[len(parts) - 1]
        
        default:
            err = shouldBeOneOf(lex, IdentTok, StringTok)
    }
    return
}

func parseAllRegOrLazyOrList(lex Lexer) (regs []Reg, lists []List, err error) {
    for {
        switch lex.Current() {
            case LazyTok:
                var reg *Reg
                reg, err = parseRegStuff(lex)
                if err != nil {
                    return
                }
                regs = append(regs, *reg)

            case RegTok:
                var reg *Reg
                reg, err = parseRegStuff(lex)
                if err != nil {
                    return
                }
                regs = append(regs, *reg)

            case ListTok:
                var list *List
                list, err = parseList(lex)
                if err != nil {
                    return
                }
                lists = append(lists, *list)

            default:
                return
        }
    }
}

func parseRegStuff(lex Lexer) (reg *Reg, err error) {
    assert(lex, RegTok, LazyTok)
    reg = &Reg{Lazy: lex.Current() == LazyTok}
    lex.Next()

    var pkgOrFirstTag string
    pkgOrFirstTag, err = parseIdent(lex)
    if err != nil {
        return
    }

    switch lex.Current() {
        case PeriodTok:
            reg.Resolver, err = parseRemainingResolver(lex, pkgOrFirstTag)

        case CommaTok: fallthrough
        case IdentTok:
            reg.Tags, err = parseMoreTags(lex, pkgOrFirstTag)
            if err != nil {
                return
            }
            reg.Resolver, err = parseResolver(lex)

        default:
            err = shouldBeOneOf(
                lex,
                PeriodTok,
                CommaTok,
                IdentTok,
            )
            return
    }

    return
}

func parseResolver(lex Lexer) (res Resolver, err error) {
    var first string
    first, err = parseIdent(lex)
    if err != nil {
        return
    }

    err = expect(lex, PeriodTok)
    if err != nil {
        return
    }
    
    res, err = parseRemainingResolver(lex, first)
    return
}

func parseRemainingResolver(lex Lexer, pkg string) (res Resolver, err error) {
    assert(lex, PeriodTok)
    lex.Next()

    res = Resolver{Package: pkg}

    res.Func, err = parseIdent(lex)
    if err != nil {
        return
    }

    res.Deps = parseDeps(lex)

    return
}

func parseDeps(lex Lexer) (deps []string) {
    for lex.Current() == IdentTok {
        deps = append(deps, lex.Value())
        lex.Next()
    }
    return
}

func parseList(lex Lexer) (list *List, err error) {
    assert(lex, ListTok)
    lex.Next()

    list = &List{}

    list.Tags, err = parseTags(lex)
    if err != nil {
        return
    }
    list.ResolveTo, err = parseTags(lex)
    if err != nil {
        return
    }

    return
}

func parseTags(lex Lexer) (tags []string, err error) {
    var first string
    first, err = parseIdent(lex)
    if err != nil {
        return
    }
    tags, err = parseMoreTags(lex, first)
    return
}

func parseMoreTags(lex Lexer, firstTag string) (tags []string, err error) {
    tags = []string{firstTag}
    for lex.Current() == CommaTok {
        lex.Next()

        var tag string
        tag, err = parseIdent(lex)
        if err != nil {
            return
        }
        tags = append(tags, tag)
    }
    return
}

func parseIdent(lex Lexer) (ident string, err error) {
    err = expect(lex, IdentTok)
    if err != nil {
        return
    }
    ident = lex.Value()
    lex.Next()
    return
}

func shouldBeOneOf(lex Lexer, expected ...Token) error {
    exp := make([]string, len(expected))
    for i, e := range expected {
        exp[i] = e.String()
    }
    return fmt.Errorf(
        "Got token '%s' but expected one of '%s'",
        lex.Current(),
        strings.Join(exp, "', '"),
    )
}

func expect(lex Lexer, expected ...Token) error {
    actual := lex.Current()
    for _, exp := range expected {
        if actual == exp {
            return nil
        }
    }
    return shouldBeOneOf(lex, expected...)
}

func assert(lex Lexer, expected ...Token) {
    actual := lex.Current()
    for _, exp := range expected {
        if actual == exp {
            return
        }
    }
    panic("PARSER BUG: " + shouldBeOneOf(lex, expected...).Error())
}

