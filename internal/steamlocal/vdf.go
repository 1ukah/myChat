package steamlocal

import (
	"errors"
	"strings"
)

type node struct {
	key  string
	vals [][2]string
	kids []*node
}

func (n *node) child(k string) *node {
	if n == nil {
		return nil
	}
	for _, c := range n.kids {
		if strings.EqualFold(c.key, k) {
			return c
		}
	}
	return nil
}

func (n *node) str(k string) string {
	if n == nil {
		return ""
	}
	for _, v := range n.vals {
		if strings.EqualFold(v[0], k) {
			return v[1]
		}
	}
	return ""
}

var errVDF = errors.New("steamlocal: bad VDF")

func parseVDF(s string) (*node, error) {
	p := &vdfParser{s: s}
	root := &node{}
	if err := p.object(root, true); err != nil {
		return nil, err
	}
	return root, nil
}

type vdfParser struct {
	s string
	i int
}

func (p *vdfParser) token() (tok string, quoted, ok bool) {
	for p.i < len(p.s) {
		c := p.s[p.i]
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n':
			p.i++
		case c == '/' && p.i+1 < len(p.s) && p.s[p.i+1] == '/':
			for p.i < len(p.s) && p.s[p.i] != '\n' {
				p.i++
			}
		case c == '{' || c == '}':
			p.i++
			return string(c), false, true
		case c == '"':
			p.i++
			var b strings.Builder
			for p.i < len(p.s) && p.s[p.i] != '"' {
				if p.s[p.i] == '\\' && p.i+1 < len(p.s) {
					p.i++
					switch p.s[p.i] {
					case 'n':
						b.WriteByte('\n')
					case 't':
						b.WriteByte('\t')
					default:
						b.WriteByte(p.s[p.i])
					}
				} else {
					b.WriteByte(p.s[p.i])
				}
				p.i++
			}
			p.i++
			return b.String(), true, true
		default:
			start := p.i
			for p.i < len(p.s) && !strings.ContainsRune(" \t\r\n{}\"", rune(p.s[p.i])) {
				p.i++
			}
			return p.s[start:p.i], true, true
		}
	}
	return "", false, false
}

func (p *vdfParser) object(n *node, top bool) error {
	for {
		key, quoted, ok := p.token()
		if !ok {
			if top {
				return nil
			}
			return errVDF
		}
		if !quoted && key == "}" {
			if top {
				return errVDF
			}
			return nil
		}
		if !quoted {
			return errVDF
		}
		val, vq, ok := p.token()
		if !ok {
			return errVDF
		}
		if !vq && val == "{" {
			c := &node{key: key}
			if err := p.object(c, false); err != nil {
				return err
			}
			n.kids = append(n.kids, c)
			continue
		}
		if !vq {
			return errVDF
		}
		n.vals = append(n.vals, [2]string{key, val})
	}
}
