package record

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	decimalSeparator  = "."
	thousandSeparator = ","
)

// Reader is the interface for record readers.
type Reader interface {
	Read() ([]Record, error)
}

// Record contains details of a finanical record.
type Record struct {
	Time   time.Time
	Text   string
	Amount int64
}

func (r *Record) StringAmount() string {
	s := strconv.FormatInt(r.Amount, 10)
	off := len(s) - 2
	return s[:off] + "," + s[off:]
}

func (r *Record) String() string {
	return fmt.Sprintf("%s\t%s\t%s", r.Time.Format("2006-01-02"), r.Text, r.StringAmount())
}

type defaultReader struct {
	rd       io.Reader
	replacer *strings.Replacer
}

func NewReader(rd io.Reader) Reader {
	return &defaultReader{
		rd:       rd,
		replacer: strings.NewReplacer(decimalSeparator, "", thousandSeparator, ""),
	}
}

func (d *defaultReader) parseAmount(s string) (int64, error) {
	v := d.replacer.Replace(s)
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (r *defaultReader) Read() ([]Record, error) {
	c := csv.NewReader(r.rd)
	c.Comma = ';'
	var rs []Record
	line := 0
	for {
		record, err := c.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		line++
		if len(record) < 4 {
			continue
		}
		t, err := time.Parse("02.01.2006", record[0])
		if err != nil {
			return nil, errors.Wrapf(err, "invalid time found on line %d: %q", line, record[0])
		}
		text := record[2]
		amount, err := r.parseAmount(record[3])
		if err != nil {
			return nil, errors.Wrapf(err, "invalid amount found on line %d: %q", line, record[3])
		}
		rs = append(rs, Record{Time: t, Text: text, Amount: amount})
	}
	return rs, nil
}