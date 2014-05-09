package main

import (
    "io"
    "fmt"
    "errors"
    "strconv"
    "text/scanner"
)

type Token int

const (
    // Identifiers
    IdentTok Token = iota

    // Literals
    StringTok

    // Keywords
    ImportTok
    RegTok
    LazyTok
    ListTok
    RootTok

    // Punctuation
    OpenParenTok
    CloseParenTok
    CommaTok
    PeriodTok

    // Sentinals
    EOFTok
)

func (self Token) String() string {
    switch self {
        case IdentTok: return "Ident"
        case StringTok: return "String"
        case ImportTok: return "Import"
        case RegTok: return "Reg"
        case LazyTok: return "Lazy"
        case ListTok: return "List"
        case RootTok: return "Root"
        case OpenParenTok: return "OpenParen"
        case CloseParenTok: return "CloseParen"
        case CommaTok: return "Comma"
        case PeriodTok: return "Period"
        case EOFTok: return "EOF"
        default: panic("Not actually a token value.")
    }
}

type Lexer interface {
    Next() Token
    Current() Token
    Value() string
    Error() error
}

type lexerStuff struct {
    s scanner.Scanner
    current Token
    value string
    error error
}

func identToToken(ident string) Token {
    switch ident {
        case "import": return ImportTok
        case "reg": return RegTok
        case "lazy": return LazyTok
        case "list": return ListTok
        case "root": return RootTok
        default: return IdentTok
    }
}

var i int
func (self *lexerStuff) Next() Token {
    self.value = ""
    tok := self.s.Scan()
    if self.error != nil {
        return EOFTok
    }
    switch tok {
        case scanner.Ident:
            value := self.s.TokenText()
            self.current = identToToken(value)
            if self.current == IdentTok {
                self.value = value
            }

        case scanner.String:
            self.value = self.s.TokenText()
            self.current = StringTok

        case ',': self.current = CommaTok
        case '(': self.current = OpenParenTok
        case ')': self.current = CloseParenTok
        case '.': self.current = PeriodTok

        case scanner.EOF:
            self.current = EOFTok

        default:
            self.error = fmt.Errorf("Invalid token %s", strconv.QuoteRune(tok))
            self.current = EOFTok
    }
    return self.current
}

func (self *lexerStuff) Current() Token {
    return self.current
}

func (self *lexerStuff) Value() string {
    return self.value
}

func (self *lexerStuff) Error() error {
    return self.error
}

func NewLexer(reader io.Reader) Lexer {
    lex := new(lexerStuff)
    lex.s = scanner.Scanner{}
    lex.s.Init(reader)
    lex.s.Error = func(s *scanner.Scanner, msg string) {
        lex.error = errors.New(msg)
    }
    lex.s.Mode = scanner.ScanIdents |
                 scanner.ScanStrings |
                 scanner.ScanComments |
                 scanner.SkipComments
    return lex
}
