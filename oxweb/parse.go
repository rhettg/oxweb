package oxweb

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
)

func ParseString(statement string) (fname string, args []string, err error) {
	// Match a word followed by a pair of parens with anything in between them.
	expressionReString := `[A-Z|a-z|0-9]+\((.+)\)`
	expressionRe, err := regexp.Compile(expressionReString)
	if !expressionRe.MatchString(statement) {
		return "", []string{}, fmt.Errorf("\"%v\" is not an Expression", statement)
	}
	// Now pull out the functiona name.
	fnameRe, err := regexp.Compile(`([A-Z|a-z|0-9]+)\((.+)\)$`)
	fnameMatches := fnameRe.FindStringSubmatch(statement)
	fname = fnameMatches[1]
	argsStr := fnameMatches[2]

	// Scan over the arguments text, keeping track of the level of parentheses
	// nesting. If we reach a comma at the top-level, end the currentWord
	// and add it to the list of arguments.
	parenLevel := 0
	quoted := false
	currentWord := []rune{}
	for _, c := range argsStr {
		switch c {
		case '(':
			parenLevel++
			continue
		case ')':
			parenLevel--
			continue
        case ' ':
            if !quoted {
                continue
            }
		// We keep track of whether we're inside of quotes, so
		// we know whether to ignore whitespace. The quotes are
		// later stripped off in ParseLiteral().
		case '"', '`', '\'':
			quoted = !quoted
		}

		if parenLevel == 0 && !quoted && c == ',' {
			args = append(args, string(currentWord))
			currentWord = []rune{}
		} else {
			//append(currentWord, c)
		}
	}
	// Don't forget to add the last word.
	args = append(args, string(currentWord))

	if parenLevel != 0 {
		return "", []string{}, fmt.Errorf("Unbalanced parentheses in \"%v\"", argsStr)
	}
	if quoted {
		return "", []string{}, fmt.Errorf("Unbalanced quote marks in \"%v\"", argsStr)
	}
	return fname, args, nil
}

type Expression interface {
	Setup(fname string, args []Expression) (err error)
	Evaluate(data JSONData) (result interface{}, err error)
	String() string
}

type Function struct {
	args []Expression
}

type Literal struct {
	value interface{}
}

func (l *Literal) Evaluate(data JSONData) (result interface{}, err error) {
	return l.value, nil
}

func (l *Literal) Setup(fname string, args []Expression) (err error) {
	// this ain't right
	return nil
}

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.value)
}

func ParseLiteral(literal string) (l *Literal, err error) {
	l = new(Literal)
	if i, err := strconv.Atoi(literal); err == nil {
		l.value = i
		return l, nil
	} else if f, err := strconv.ParseFloat(literal, 64); err == nil {
		l.value = f
		return l, nil
	} else if unquoted, err := strconv.Unquote(literal); err == nil {
		l.value = unquoted
		return l, nil
	}
	return nil, fmt.Errorf("Couldn't parse %s as a literal", literal)
}

func Parse(statement string) (expr Expression, err error) {
	// First try to parse literals
	if expr, err = ParseLiteral(statement); err == nil {
		return
	}

	// Base case: statement is a single expression (e.g. Foo(a,b))
	fname, args, err := ParseString(statement)

	// Then treat it as a GetDeep expression if it's not an expression
	if err != nil {
		expr, err := NewGetDeepExpression(statement)
		if err == nil {
			log.Printf("found a get deep expr: %v, args: %v", expr, args)
			return expr, err
		}
		log.Printf("couldn't parse get deep expr: ", err)
	}

	// Now start parsing the rest
	expressionArgs := []Expression{}
	for _, arg := range args {
		argExpr, err := Parse(arg)
		if err != nil {
			return nil, err
		}
		expressionArgs = append(expressionArgs, argExpr)
	}

	switch {
	case fname == "RandomSample":
		expr = new(RandomSample)
	case fname == "EveryNth":
		expr = new(EveryNth)
	case fname == "GetDeep":
		expr = new(GetDeepExpression)
	case fname == "Subtract" || fname == "Add" || fname == "Divide" || fname == "Multiply":
		expr = new(ArithmeticOperator)
	case fname == "RollingWindow":
		expr = new(RollingWindow)
	case fname == "TimedWindow":
		expr = new(TimedWindow)
	case fname == "WindowAve":
		expr = new(WindowAve)
	case fname == "As":
		expr = new(AsClause)

	default:
		return nil, fmt.Errorf("Unrecognized function name '%s'", fname)
	}
	err = expr.Setup(fname, expressionArgs)
	return
}
