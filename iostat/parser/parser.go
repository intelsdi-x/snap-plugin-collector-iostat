package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	NsVendor = "intel"
	NsType   = "iostat"

	defaultEmptyTokenAcceptance = 5
)

type parser struct {
	// this structure is used in parsing iostat command output
	firstLine   bool // set true if next interval is exepected
	titleLine   bool // set true if the line is a title
	emptyTokens int  // numbers of empty tokens got from data channel

	statType    string   // type of statistics (for example cpu or device statistic)
	statSubType string   // subtype of statistics (for example sda)
	statNames   []string // names of statistics

	stat   string   // statictic namespace, includes statType, statSubType and StatName
	stats  []string // slice of statistics, after parsing process it's equivalent to IOSTAT.keys
	values []string // slice of statictics' values

	keys []string
	data map[string]float64
}

func New() *parser {
	return &parser{
		keys: []string{},
		data: map[string]float64{},
	}
}

func (p *parser) Parse(reader io.Reader) ([]string, map[string]float64, error) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		err := p.parse(scanner.Text())
		if err != nil {
			log.WithFields(log.Fields{
				"line":  scanner.Text(),
				"error": err,
			}).Debug("failed to parse line")
		}
	}

	return p.keys, p.data, nil
}

func (p *parser) parse(data string) error {
	line := strings.Fields(data)
	if len(line) == 0 {
		// slice "line" is empty
		p.emptyTokens++
		if p.emptyTokens > defaultEmptyTokenAcceptance {
			return errors.New("more empty data slice than acceptance level")
		}
		return nil
	}
	p.emptyTokens = 0

	if p.titleLine {
		// skip the title line
		p.titleLine = false
		return nil
	}
	if p.firstLine {
		//next interval separated by a newline
		p.firstLine = false
		p.stats = []string{}
		p.values = []string{}
		return nil
	}

	if strings.HasSuffix(line[0], ":") {
		if len(line) > 1 {
			p.statType = strings.ToLower(strings.TrimSuffix(line[0], ":"))
			p.statNames = replaceByPerSec(line[1:])
			return nil
		}
	}

	if len(p.statNames) == 0 || len(p.statType) == 0 {
		return errors.New("Invalid format of iostat output data")
	}
	if len(line) > len(p.statNames) {
		// subType is defined
		p.statSubType = line[0]
		line = line[1:]
	} else {
		p.statSubType = ""
	}

	if len(line) == len(p.statNames) && len(p.statNames) != 0 {
		for i, sname := range p.statNames {
			if p.statSubType != "" {
				p.stat = p.statType + "/" + p.statSubType + "/" + sname
			} else {
				p.stat = p.statType + "/" + sname
			}

			p.stats = append(p.stats, p.stat)
			p.values = append(p.values, line[i])
		}
	}

	if strings.ToLower(p.statSubType) == "all" {
		// all available metrics keys collected
		p.firstLine = true // for next scan skip first line

		if len(p.keys) == 0 {

			if len(p.stat) == 0 {
				return errors.New("can not retrive iostat metrics namespace")
			}

			p.keys = make([]string, len(p.stats))
			for i, s := range p.stats {
				p.keys[i] = joinNamespace(createNamespace(s))
			}
		}

		if len(p.values) != len(p.keys) {
			// number of values has to be equivalent to number of keys
			return errors.New("invalid parsing iostat output")
		}

		for i, val := range p.values {
			v, err := strconv.ParseFloat(strings.TrimSpace(strings.Replace(val, ",", ".", 1)), 64)
			if err == nil {
				p.data[p.keys[i]] = v
			} else {
				fmt.Fprintln(os.Stderr, "invalid metric value", err)
			}
		}
	}

	return nil
}

// replacePerSec turns "/s" into "_per_sec"
func replaceByPerSec(slice []string) []string {
	for i, str := range slice {
		slice[i] = strings.Replace(str, "/s", "_per_sec", 1)
	}
	return slice
}

func joinNamespace(ns []string) string {
	return "/" + strings.Join(ns, "/")
}

// createNamespace returns namespace slice of strings composed from: vendor, type and ceph-daemon name
func createNamespace(name string) []string {
	return []string{NsVendor, NsType, name}
}
