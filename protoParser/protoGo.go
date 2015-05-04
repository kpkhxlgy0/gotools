package protoParser

import (
	"bufio"
	"os"
	"regexp"

	"uuzu.com/hanxl/utils"
)

type Message struct {
	IsArr   bool
	TypeStr string
	Name    string
}

type Messages struct {
	Name  string
	Value []Message
}

type ProtoGo struct {
	Package  string
	Messages []Messages
}

func ParseGo(path string) *ProtoGo {
	ret := new(ProtoGo)
	f, e := os.Open(path)
	if e != nil {
		return nil
	}
	r := bufio.NewReader(f)
	e = parsePackage(r, ret)
	if e != nil {
		return nil
	}
	e = parseMessage(r, ret)
	if e != nil {
		return nil
	}
	return ret
}

func parsePackage(r *bufio.Reader, p *ProtoGo) error {
	pattern := regexp.MustCompile(`package[\s\t]+(\w+)[\s\t]*;.*`)
	for {
		line, e := utils.ReadLine(r)
		if e != nil {
			return nil
		}
		arr := pattern.FindStringSubmatch(line)
		if len(arr) > 0 {
			p.Package = arr[1]
			return nil
		}
	}
}

func parseMessage(r *bufio.Reader, p *ProtoGo) error {
	patternBegin := regexp.MustCompile(`message[\s\t]+(\w+)[\s\t]*\{.*`)
	patternEnd := regexp.MustCompile(`\}.*`)
	pattern := regexp.MustCompile(`[\s\t]*(required|optional|repeated)[\s\t]+(\w+)[\s\t]+(\w+)[\s\t]*=[\s\t]*\d*.*`)
	var messages *Messages = nil
	for {
		line, e := utils.ReadLine(r)
		if e != nil {
			return nil
		}
		if messages == nil {
			arr := patternBegin.FindStringSubmatch(line)
			if len(arr) > 0 {
				messages = new(Messages)
				messages.Name = arr[1]
				messages.Value = []Message{}
			}
			continue
		}
		if patternEnd.MatchString(line) {
			p.Messages = append(p.Messages, *messages)
			messages = nil
			continue
		}
		arr := pattern.FindStringSubmatch(line)
		if len(arr) > 0 && arr[3] != "one_key" && arr[3] != "sec_key" {
			m := Message{
				IsArr:   (arr[1] == "repeated"),
				TypeStr: arr[2],
				Name:    arr[3],
			}
			messages.Value = append(messages.Value, m)
		}
	}
	return nil
}
