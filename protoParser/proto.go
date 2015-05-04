package protoParser

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"

	"uuzu.com/hanxl/utils"
)

type Proto utils.Map
type Opcode utils.Map

type service utils.Map

func (this *service) get(key interface{}) interface{} {
	for _, v := range this.Items {
		if v.Key == key {
			return v.Value
		}
	}
	return nil
}

func Parse(path string) *Proto {
	ret := new(Proto)
	f, e := os.Open(path)
	if e != nil {
		return nil
	}
	r := bufio.NewReader(f)
	e = parseEnum(r, ret)
	if e != nil || len(ret.Items) == 0 {
		return nil
	}
	return ret
}

func ParseOpcode(path string) *Opcode {
	f, e := os.Open(path)
	if e != nil {
		return nil
	}
	r := bufio.NewReader(f)
	s, e := parseService(r)
	if e != nil || s == nil {
		return nil
	}
	ret, e := parseCommand(r, s)
	if e != nil {
		return nil
	}
	return ret
}

func parseService(r *bufio.Reader) (*service, error) {
	patternBegin := regexp.MustCompile(`enum[\s\t]+server[\s\t]*\{.*`)
	pattern := regexp.MustCompile(`[\s\t]*(\w+)[\s\t]*=[\s\t]*(\d*).*`)
	patternEnd := regexp.MustCompile(`\}.*`)
	var ret *service = nil
	for {
		line, e := utils.ReadLine(r)
		if e != nil {
			return ret, nil
		}
		if ret == nil {
			if patternBegin.MatchString(line) {
				ret = new(service)
			}
			continue
		}
		if patternEnd.MatchString(line) {
			return ret, nil
		}
		arr := pattern.FindStringSubmatch(line)
		if len(arr) > 0 {
			ret.Items = append(ret.Items, utils.Pair{
				Key:   strings.ToLower(arr[1]),
				Value: arr[2],
			})
		}
	}
	return ret, nil
}

func parseCommand(r *bufio.Reader, s *service) (*Opcode, error) {
	patternBegin := regexp.MustCompile(`enum[\s\t]*OP[\s\t]*\{.*`)
	pattern := regexp.MustCompile(`[\s\t]*[CS]M_(\w+)[\s\t]*=[\s\t]*(\d+).*`)
	patternEnd := regexp.MustCompile(`\}.*`)
	patternMajorBegin := regexp.MustCompile(`[\s\t]*begin_server_(\w+)[\s\t]*=[\s\t]*\d*.*`)
	patternMajorEnd := regexp.MustCompile(`[\s\t]*end_server_(\w+)[\s\t]*=[\s\t]*\d*.*`)
	var ret *Opcode = nil
	var major *string = nil
	for {
		line, e := utils.ReadLine(r)
		if e != nil {
			return ret, nil
		}
		if ret == nil {
			if patternBegin.MatchString(line) {
				ret = new(Opcode)
			}
			continue
		}
		if patternEnd.MatchString(line) {
			return ret, nil
		}
		if major == nil {
			arr := patternMajorBegin.FindStringSubmatch(line)
			if len(arr) > 0 {
				major = &arr[1]
			}
			continue
		}
		arr := patternMajorEnd.FindStringSubmatch(line)
		if len(arr) > 0 && arr[1] == *major {
			major = nil
			continue
		}
		arr = pattern.FindStringSubmatch(line)
		if len(arr) > 0 {
			majorStr, _ := s.get(*major).(string)
			ret.Items = append(ret.Items, utils.Pair{
				Key:   "k" + majorStr + "_" + arr[2],
				Value: utils.FormatName(arr[1]),
			})
		}
	}
	return ret, nil
}

func parseEnum(r *bufio.Reader, p *Proto) error {
	patternBegin := regexp.MustCompile(`enum[\s\t]+(\w+)[\s\t]*\{.*`)
	pattern := regexp.MustCompile(`[\s\t]*(\w+)[\s\t]*=[\s\t]*(\d+).*`)
	patternEnd := regexp.MustCompile(`\}.*`)
	key := ""
	var item *utils.Map = nil
	for {
		line, e := utils.ReadLine(r)
		if e != nil {
			return nil
		}
		if item == nil {
			arr := patternBegin.FindStringSubmatch(line)
			if len(arr) > 0 {
				key = arr[1]
				item = new(utils.Map)
			}
			continue
		}
		if patternEnd.MatchString(line) {
			p.Items = append(p.Items, utils.Pair{
				Key:   key,
				Value: *item,
			})
			item = nil
			continue
		}
		arr := pattern.FindStringSubmatch(line)
		if len(arr) > 0 {
			intValue, e := strconv.Atoi(arr[2])
			if e != nil {
				return e
			}
			item.Items = append(item.Items, utils.Pair{
				Key:   arr[1],
				Value: intValue,
			})
		}
	}
	return nil
}
