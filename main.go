package inifile

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Ini struct {
	name string
}

var (
	ErrSectionExists    = errors.New("Section already exists")
	ErrSectionNotExists = errors.New("Section doesn't exist")
	ErrKeyNotExists     = errors.New("Key doesn't exist")
)

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}

func join(lines []string) string {
	str := ""
	for _, l := range lines {
		str += l
	}

	return str
}

func NewIniFile(name string) (*Ini, error) {
	if !fileExists(name) {
		_, err := os.Create(name)
		if err != nil {
			return nil, err
		}
	}

	return &Ini{name}, nil
}

func (ini *Ini) open(permissions int) (*os.File, error) {
	return os.OpenFile(ini.name, permissions, 0644)
}

func (ini *Ini) sectionExists(name string) (int64, error) {
	file, err := ini.open(os.O_RDONLY)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	s, err := file.Stat()
	if err != nil {
		return -1, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, s.Size()))
	_, err = io.Copy(buf, file)
	if err != nil {
		return -1, err
	}

	offset := 0

	for {
		line, err := buf.ReadString('\n')
		if err != nil && err != io.EOF {
			return -1, err
		}

		offset += len(line)

		if err == io.EOF {
			return -1, nil
		}

		if strings.TrimSpace(line) == fmt.Sprintf("[%s]", name) {
			return int64(offset), nil
		}
	}
}

func (ini *Ini) AddSection(name string) error {
	offset, err := ini.sectionExists(name)
	if err != nil {
		return err
	}

	file, err := ini.open(os.O_WRONLY)
	if err != nil {
		return err
	}
	defer file.Close()

	s, err := file.Stat()
	if err != nil {
		return err
	}

	if offset == -1 {

		_, err = file.WriteAt([]byte(fmt.Sprintf("\n[%s]\n", name)), s.Size())
		if err != nil {
			return err
		}
	} else {
		return ErrSectionExists
	}

	return nil
}

func (ini *Ini) WriteKey(section string, key string, value string) error {
	offset, err := ini.sectionExists(section)
	if err != nil {
		return err
	}

	file, err := ini.open(os.O_RDWR)
	if err != nil {
		return err
	}
	defer file.Close()

	s, err := file.Stat()
	if err != nil {
		return err
	}

	if offset == -1 {
		err = ini.AddSection(section)
		if err != nil {
			return err
		}

		toWrite := ""
		if s.Size() == 0 {
			toWrite = fmt.Sprintf("[%s]\n%s = %s\n",
				section, key, value)
		} else {
			toWrite = fmt.Sprintf("\n[%s]\n%s = %s\n",
				section, key, value)
		}
		_, err = file.WriteAt([]byte(toWrite), s.Size())
		if err != nil {
			return err
		}

	} else {
		out := []string{}
		inserted := false

		_, err = file.Seek(offset, os.SEEK_SET)
		if err != nil {
			return err
		}

		lines := bytes.NewBuffer(make([]byte, 0, s.Size()))
		_, err = io.Copy(lines, file)
		if err != nil {
			return err
		}

		for {
			currLine, err := lines.ReadString('\n')
			if err != nil && err != io.EOF {
				return err
			}

			if currLine == "\n" || err == io.EOF && !inserted {
				fmt.Println("works")
				out = append(out, fmt.Sprintf("%s = %s\n", key, value))
				inserted = true
			}

			if err == io.EOF {
				str := join(out)
				fmt.Println(str)

				_, err := file.WriteAt([]byte(str), offset)
				if err != nil {
					log.Fatal(err)
				}

				break
			}

			out = append(out, currLine)
		}
	}

	return nil
}

func (ini *Ini) ReadKey(section string, key string) (string, error) {
	offset, err := ini.sectionExists(section)
	if err != nil {
		return "", err
	}

	if offset == -1 {
		return "", ErrSectionNotExists
	}

	file, err := ini.open(os.O_RDONLY)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Seek(offset, os.SEEK_SET)
	if err != nil {
		return "", err
	}

	s, err := file.Stat()
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(make([]byte, 0, s.Size()))
	_, err = io.Copy(buf, file)
	if err != nil {
		return "", err
	}

	for {
		line, err := buf.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", err
		}

		if err == io.EOF {
			return "", ErrKeyNotExists
		}

		data := strings.Split(line, "=")
		if strings.TrimSpace(data[0]) == key {
			return strings.TrimSpace(data[1]), nil
		}
	}
}

func (ini *Ini) DeleteKey(section string, key string) error {
	offset, err := ini.sectionExists(section)
	if err != nil {
		return err
	}

	if offset == -1 {
		return ErrSectionNotExists
	}

	file, err := ini.open(os.O_RDWR)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(offset, os.SEEK_SET)
	if err != nil {
		return err
	}

	s, err := file.Stat()
	if err != nil {
		return err
	}

	size := s.Size()
	buf := bytes.NewBuffer(make([]byte, 0, size))
	_, err = io.Copy(buf, file)
	if err != nil {
		return err
	}

	out := []string{}
	found := false

	for {
		line, err := buf.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}

		data := strings.Split(line, "=")
		if strings.TrimSpace(data[0]) == key {
			found = true
			size -= int64(len(line))
			continue
		}

		if err == io.EOF {
			break
		}

		out = append(out, line)
	}

	if !found {
		return ErrKeyNotExists
	}

	str := join(out)
	_, err = file.WriteAt([]byte(str), offset)
	if err != nil {
		return err
	}

	_, err = file.Seek(0, os.SEEK_END)
	if err != nil {
		return err
	}

	err = file.Truncate(size)
	if err != nil {
		log.Fatal(err)
	}
	return err
}
