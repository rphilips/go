package dfile

import (
	"regexp"
	"strings"

	qfnmatch "brocade.be/base/fnmatch"
)

// Terms : predefined value
var Terms = map[string]bool{
	`project`:     true,
	`qpath`:       true,
	`qdir`:        true,
	`qrelpath`:    true,
	`basename`:    true,
	`ext`:         true,
	`version`:     true,
	`mostype`:     true,
	`mclib`:       true,
	`systemname`:  true,
	`os`:          true,
	`systemgroup`: true,
	`systemroles`: true,
}

// Operator : available operators
var Operator = map[string]bool{
	`sortsAfter`:    true,
	`sortsBefore`:   true,
	`contains`:      true,
	`isEqualTo`:     true,
	`fileMatches`:   true,
	`regexpMatches`: true,
	`isIn`:          true,
	`isInstanceOf`:  true,
	`isEqualTrueAs`: true,
	`isPrefixOf`:    true,
	`isSuffixOf`:    true,
	`startsWith`:    true,
	`endsWith`:      true,
}

// Eval checks if a guard matches
func Eval(guard []string, env map[string]string) bool {
	if len(guard) == 0 {
		return true
	}

	stack := []bool{}
	st := -1
	value := ""
	term := ""
	for _, p := range guard {
		switch {
		case p == "not":
			stack[st] = !stack[st]
		case p == "and":
			stack[st-1] = stack[st] && stack[st-1]
			st--
		case p == "or":
			stack[st-1] = stack[st] || stack[st-1]
			st--
		case p == "true":
			stack = append(stack, true)
			st++
		case p == "false":
			stack = append(stack, false)
			st++
		case Operator[p] || strings.HasPrefix(p, "not-"):
			operand := p
			b := eval(term, operand, value, env)
			term = ""
			value = ""
			stack = append(stack, b)
			st++
		default:
			if term == "" {
				term = p
			} else {
				value = p
			}
		}
	}
	return stack[0]
}

func eval(term, operand, value string, env map[string]string) bool {
	if strings.HasPrefix(operand, "not-") {
		return !eval(term, operand[4:], value, env)
	}
	termc := env[term]
	//fmt.Printf("term: %s; termc: %s; value: %s; operand: %s\n", term, termc, value, operand)
	switch operand {
	case "fileMatches":
		return qfnmatch.Match(value, termc)
	case "regexpMatches":
		result, err := regexp.MatchString(value, termc)
		if err != nil {
			return false
		}
		return result
	case "isEqualTo":
		return value == termc
	case "sortsAfter":
		return strings.Compare(termc, value) > 0
	case "sortsBefore":
		return strings.Compare(termc, value) < 0
	case "contains":
		return strings.Contains(termc, value)
	case "isPrefixOf":
		return strings.HasPrefix(value, termc)
	case "startsWith":
		return strings.HasPrefix(termc, value)
	case "isSuffixOf":
		return strings.HasSuffix(value, termc)
	case "endsWith":
		return strings.HasSuffix(termc, value)
	case "isIn":
		return strings.Contains(value, termc)
	case "isEqualTrueAs":
		return yes(value) == yes(termc)
	case "isInstanceOf":
		if value == "empty" {
			return termc == ""
		}
		if termc == "" {
			return false
		}
		if value == "intlit" {
			re := regexp.MustCompile(`^[+-]?[0-9]+$`)
			return re.MatchString(termc)
		}
		if value == "numlit" {
			if strings.Count(termc, ".") > 1 {
				return false
			}
			if strings.Count(termc, "E") > 1 {
				return false
			}
			termc = strings.Replace(termc, ".", "", -1)
			re := regexp.MustCompile(`^[+-]?[0-9]+$`)
			result := re.MatchString(termc)
			if result {
				return result
			}
			re = regexp.MustCompile(`^[+-]?[0-9]*E[+-]?[0-9]+$`)
			result = re.MatchString(termc)
			return result
		}
		if value == "strlit" {
			return strings.HasPrefix(termc, `"`) && strings.HasSuffix(termc, `"`)
		}
		if value == "name" {
			re := regexp.MustCompile(`^[%a-zA-Z][a-zA-Z0-9]*$`)
			return re.MatchString(termc)
		}
		if value == "lvn" {
			return lvn(termc)
		}
		if value == "gvn" {
			return strings.HasPrefix(termc, "^") || strings.HasPrefix(termc, "@")
		}
		if value == "glvn" {
			result := strings.HasPrefix(termc, "^") || strings.HasPrefix(termc, "@")
			if result {
				return true
			}
			return lvn(termc)
		}
		if value == "expritem" {
			result := lvn(termc)
			if result {
				return false
			}
			termc = reduce(termc)
			if string(termc[0]) == "^" {
				return false
			}
			re := regexp.MustCompile(`^\.[@a-zA-Z%]`)
			return !re.MatchString(termc)
		}
		if value == "actualname" {
			re := regexp.MustCompile(`^[%a-zA-Z][a-zA-Z0-9]*$`)
			if re.MatchString(termc) {
				return true
			}
			return strings.HasPrefix(termc, "@")
		}
		if value == "actual" {
			if strings.HasPrefix(termc, ".") {
				termc = termc[1:]
				if termc == "" {
					return false
				}
				re := regexp.MustCompile(`^[%a-zA-Z][a-zA-Z0-9]*$`)
				if re.MatchString(termc) {
					return true
				}
				return strings.HasPrefix(termc, "@")
			}

			termc = reduce(termc)
			if strings.HasPrefix(termc, "(") {
				return false
			}
			if strings.Contains(termc, ",") {
				return false
			}
			return true
		}
		return false
	default:
		return false
	}
}

func lvn(term string) bool {
	if term == "" {
		return false
	}
	if term != "@" && strings.HasPrefix(term, "@") {
		return true
	}
	first := string(term[0])
	if !strings.Contains("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ%", first) {
		return false
	}
	term = reduce(term)
	re := regexp.MustCompile(`^[%a-zA-Z0-9]*$`)
	return re.MatchString(term)
}

func reduce(term string) string {
	re := regexp.MustCompile(`"[^"]*"`)
	term = re.ReplaceAllString(term, `"1"`)
	re = regexp.MustCompile(`"1"+`)
	term = re.ReplaceAllString(term, `"1"`)
	re = regexp.MustCompile(`\([^()]*\)`)
	for {
		x := term
		term = re.ReplaceAllString(term, `a`)
		if x == term {
			break
		}
	}
	return term
}

func yes(value string) bool {
	t := strings.ToLower(value)
	if t != "" && strings.Contains("jy1t", string(t[0])) {
		return true
	}
	return false
}
